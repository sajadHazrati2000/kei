import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { AvailabilityService } from '../../../core/availability/availability.service';
import { AuthService } from '../../../core/auth/auth.service';
import { KeiThemeService } from '../../../core/services/kei-theme.service';
import { Slot, SlotInput, SlotStatus, WeekDay } from '../../../core/availability/availability.models';
import { WeekNavigatorComponent } from '../week-navigator/week-navigator.component';
import { SlotBlockComponent } from '../slot-block/slot-block.component';
import { SlotCreatorComponent } from '../slot-creator/slot-creator.component';
import { JalaliPipe } from '../../../shared/pipes/jalali.pipe';

const ROW_H = 24; // px per 30-min row
const HOURS = Array.from({ length: 24 }, (_, i) => i);

function getWeekStart(date: Date): Date {
  const d = new Date(date);
  // Start on Sunday (0)
  d.setDate(d.getDate() - d.getDay());
  d.setHours(0, 0, 0, 0);
  return d;
}

function getWeekDates(weekStart: Date): Date[] {
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(weekStart);
    d.setDate(d.getDate() + i);
    return d;
  });
}

function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

function dotStatus(slots: Slot[]): 'none' | 'busy' | 'focus' | 'free' {
  if (!slots.length) return 'none';
  if (slots.some(s => s.status === 'busy'))  return 'busy';
  if (slots.some(s => s.status === 'focus')) return 'focus';
  return 'free';
}

function formatHour(h: number): string {
  return `${String(h).padStart(2, '0')}:00`;
}

@Component({
  selector: 'app-my-week',
  standalone: true,
  imports: [
    DatePipe,
    TranslateModule,
    WeekNavigatorComponent,
    SlotBlockComponent,
    SlotCreatorComponent,
    JalaliPipe,
  ],
  templateUrl: './my-week.component.html',
  styleUrl:    './my-week.component.scss',
})
export class MyWeekComponent implements OnInit {
  private readonly avail = inject(AvailabilityService);
  protected readonly auth  = inject(AuthService);
  protected readonly theme = inject(KeiThemeService);

  // ── State ─────────────────────────────────────────────────────────────────
  protected weekStart    = signal(getWeekStart(new Date()));
  protected selectedDate = signal(new Date());
  protected allSlots     = signal<Slot[]>([]);
  protected editingSlotId = signal<string | null>(null);
  protected creatorAt    = signal<{ hour: number; minute: number } | null>(null);
  protected loading      = signal(false);

  // ── Derived ───────────────────────────────────────────────────────────────
  protected weekDays = computed<WeekDay[]>(() => {
    const dates = getWeekDates(this.weekStart());
    return dates.map(date => {
      const slots = this.allSlots().filter(s =>
        isSameDay(new Date(s.from), date)
      );
      return { date, dot: dotStatus(slots), slots };
    });
  });

  protected daySlots = computed(() =>
    this.allSlots().filter(s =>
      isSameDay(new Date(s.from), this.selectedDate())
    )
  );

  protected readonly hours   = HOURS;
  protected readonly rowH    = ROW_H;
  protected readonly formatH = formatHour;

  ngOnInit(): void {
    this.loadSlots();
  }

  // ── Data loading ───────────────────────────────────────────────────────────
  private loadSlots(): void {
    this.loading.set(true);
    this.avail.getMySlots(this.weekStart()).subscribe({
      next: slots => { this.allSlots.set(slots); this.loading.set(false); },
      error: () => this.loading.set(false),
    });
  }

  // ── Navigation ─────────────────────────────────────────────────────────────
  protected onWeekChange(newStart: Date): void {
    this.weekStart.set(newStart);
    this.selectedDate.set(newStart);
    this.creatorAt.set(null);
    this.editingSlotId.set(null);
    this.loadSlots();
  }

  protected selectDay(date: Date): void {
    this.selectedDate.set(date);
    this.creatorAt.set(null);
    this.editingSlotId.set(null);
  }

  protected isSameDay = isSameDay;
  protected isWeekend = (d: Date): boolean => d.getDay() === 0 || d.getDay() === 6;

  // ── Grid interaction ────────────────────────────────────────────────────────
  protected onGridClick(event: MouseEvent): void {
    // Ignore clicks on slot blocks or the creator
    if ((event.target as Element).closest('.slot-block, .slot-creator')) return;

    const grid = event.currentTarget as HTMLElement;
    const rect  = grid.getBoundingClientRect();
    const y     = event.clientY - rect.top + grid.scrollTop;
    const halfH = Math.floor(y / ROW_H);
    const hour   = Math.min(Math.floor(halfH / 2), 23);
    const minute = (halfH % 2) * 30;

    this.editingSlotId.set(null);
    this.creatorAt.set({ hour, minute });
  }

  protected editSlot(slot: Slot): void {
    this.creatorAt.set(null);
    this.editingSlotId.set(slot.id);
  }

  protected deleteSlot(slot: Slot): void {
    this.avail
      .deleteSlot(this.weekStart(), this.allSlots(), slot.id)
      .subscribe({ next: updated => this.allSlots.set(updated) });
    this.editingSlotId.set(null);
  }

  protected resizeSlot(event: { slot: Slot; from: string; to: string }): void {
    this.allSlots.update(slots =>
      slots.map(s => s.id === event.slot.id ? { ...s, from: event.from, to: event.to } : s)
    );
    const inputs: SlotInput[] = this.allSlots().map(s => ({
      from: s.from, to: s.to, status: s.status,
    }));
    this.avail.setSlots(this.weekStart(), inputs).subscribe({
      next: updated => this.allSlots.set(updated),
    });
  }

  protected onSlotCreated(slot: SlotInput): void {
    this.avail
      .createSlot(this.weekStart(), this.allSlots(), slot)
      .subscribe({ next: updated => this.allSlots.set(updated) });
    this.creatorAt.set(null);
  }

  protected closeCreator(): void {
    this.creatorAt.set(null);
    this.editingSlotId.set(null);
  }
}
