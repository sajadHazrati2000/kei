import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface User {
  id:            string;
  org_id:        string;
  name:          string;
  email:         string;
  role:          'admin' | 'member' | 'viewer';
  timezone:      string;
  language:      string;
  calendar_pref: 'gregorian' | 'jalali';
  is_active:     boolean;
}

export interface SetupPayload {
  org_name:   string;
  org_slug:   string;
  admin_name: string;
  email:      string;
  password:   string;
  timezone:   string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http   = inject(HttpClient);
  private readonly router = inject(Router);
  private readonly base   = environment.apiUrl + '/api/v1';

  private readonly _user$ = new BehaviorSubject<User | null>(null);
  readonly currentUser$     = this._user$.asObservable();
  readonly isAuthenticated$ = this._user$.pipe(map(u => u !== null));

  get currentUser(): User | null    { return this._user$.value; }
  get isAuthenticated(): boolean    { return this._user$.value !== null; }

  // ── Session rehydration ────────────────────────────────────────────────────

  /** Called by APP_INITIALIZER. Reads the httpOnly cookie via the API.
   *  Returns null (never throws) so a missing session is handled gracefully. */
  me(): Observable<User | null> {
    return this.http
      .get<{ data: User }>(`${this.base}/users/me`, { withCredentials: true })
      .pipe(
        map(r => r.data),
        tap(u  => this._user$.next(u)),
        catchError(() => { this._user$.next(null); return of(null); })
      );
  }

  // ── Auth flows ─────────────────────────────────────────────────────────────

  login(email: string, password: string): Observable<User> {
    return this.http
      .post<{ user: User }>(`${this.base}/auth/login`, { email, password }, { withCredentials: true })
      .pipe(
        map(r => r.user),
        tap(u => this._user$.next(u))
      );
  }

  setup(payload: SetupPayload): Observable<User> {
    return this.http
      .post<{ user: User }>(`${this.base}/auth/setup`, payload, { withCredentials: true })
      .pipe(
        map(r => r.user),
        tap(u => this._user$.next(u))
      );
  }

  logout(): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/auth/logout`, { withCredentials: true })
      .pipe(
        tap(() => this._user$.next(null)),
        catchError(() => { this._user$.next(null); return of(undefined); })
      );
  }

  /** Used by the SetupGuard to decide whether to show the wizard. */
  isSetupDone(): Observable<boolean> {
    return this.http
      .get<{ setup_done: boolean }>(`${this.base}/auth/setup/status`)
      .pipe(
        map(r => r.setup_done),
        catchError(() => of(true)) // fail safe — assume done to avoid wizard loop
      );
  }

  /** Clears the session without a server round-trip (used by interceptor). */
  clearSession(): void {
    this._user$.next(null);
  }

  updateSettings(payload: {
    timezone: string;
    working_start: string;
    working_end: string;
  }): Observable<void> {
    return this.http
      .put<void>(`${this.base}/settings`, {
        name:             this._user$.value?.org_id ?? '',
        timezone:         payload.timezone,
        overlap_start:    payload.working_start,
        overlap_end:      payload.working_end,
        default_language: this._user$.value?.language ?? 'en',
        working_days:     [1, 2, 3, 4, 5],
        working_start:    payload.working_start,
        working_end:      payload.working_end,
      }, { withCredentials: true })
      .pipe(map(() => undefined));
  }
}
