import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { AuthService } from '../auth/auth.service';
import {
  OverlapResult,
  RecurringTemplate,
  SetSlotsRequest,
  Slot,
  SlotInput,
  UserAvailability,
} from './availability.models';

const BASE = environment.apiUrl + '/api/v1';

/** Formats a Date as an ISO 8601 UTC string (full datetime). */
function toISO(d: Date): string {
  return d.toISOString();
}

/** Returns the Monday-anchored week end date (exclusive). */
function weekEnd(weekStart: Date): Date {
  const d = new Date(weekStart);
  d.setDate(d.getDate() + 7);
  return d;
}

@Injectable({ providedIn: 'root' })
export class AvailabilityService {
  private readonly http = inject(HttpClient);
  private readonly auth = inject(AuthService);

  private get userId(): string {
    return this.auth.currentUser?.id ?? '';
  }

  // ── Per-user slots ─────────────────────────────────────────────────────────

  /** Fetch the current user's slots for the 7-day window starting on weekStart. */
  getMySlots(weekStart: Date): Observable<Slot[]> {
    return this.http
      .get<{ data: Slot[]; total: number }>(
        `${BASE}/availability/${this.userId}`,
        { params: { from: toISO(weekStart), to: toISO(weekEnd(weekStart)) } }
      )
      .pipe(map(r => r.data ?? []));
  }

  /** Replace all slots in the given week window. Returns the saved slots. */
  setSlots(weekStart: Date, slots: SlotInput[]): Observable<Slot[]> {
    const body: SetSlotsRequest = {
      from:  toISO(weekStart),
      to:    toISO(weekEnd(weekStart)),
      slots,
    };
    return this.http
      .put<{ data: Slot[]; total: number }>(
        `${BASE}/availability/${this.userId}`,
        body
      )
      .pipe(map(r => r.data ?? []));
  }

  /** Add a new slot to the week, preserving existing ones. */
  createSlot(weekStart: Date, allCurrentSlots: Slot[], newSlot: SlotInput): Observable<Slot[]> {
    const inputs = [
      ...allCurrentSlots.map(s => ({ from: s.from, to: s.to, status: s.status })),
      newSlot,
    ];
    return this.setSlots(weekStart, inputs);
  }

  /** Remove a slot by id, then persist the remaining set via setSlots (PUT). */
  deleteSlot(weekStart: Date, allCurrentSlots: Slot[], slotId: string): Observable<Slot[]> {
    const inputs = allCurrentSlots
      .filter(s => s.id !== slotId)
      .map(s => ({ from: s.from, to: s.to, status: s.status }));
    return this.setSlots(weekStart, inputs);
  }

  // ── Recurring templates ────────────────────────────────────────────────────

  getTemplates(): Observable<RecurringTemplate[]> {
    return this.http
      .get<{ data: RecurringTemplate[] }>(`${BASE}/availability/${this.userId}/recurring`)
      .pipe(map(r => r.data ?? []));
  }

  setTemplates(templates: Partial<RecurringTemplate>[]): Observable<RecurringTemplate[]> {
    return this.http
      .put<{ data: RecurringTemplate[] }>(
        `${BASE}/availability/${this.userId}/recurring`,
        { templates }
      )
      .pipe(map(r => r.data ?? []));
  }

  // ── Team ──────────────────────────────────────────────────────────────────

  getTeamAvailability(weekStart: Date): Observable<UserAvailability[]> {
    return this.http
      .get<{ data: UserAvailability[] }>(`${BASE}/team/availability`, {
        params: { from: toISO(weekStart), to: toISO(weekEnd(weekStart)) },
      })
      .pipe(map(r => r.data ?? []));
  }

  getOverlapWindows(date: Date): Observable<OverlapResult> {
    const dateStr = date.toISOString().split('T')[0];
    return this.http
      .get<OverlapResult>(`${BASE}/team/overlap`, { params: { date: dateStr } });
  }
}
