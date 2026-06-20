import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { AuthService } from './auth.service';

const BASE = 'http://localhost:8080/api/v1';

describe('AuthService', () => {
  let service: AuthService;
  let http:    HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(AuthService);
    http    = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('login() POSTs to /auth/login and updates currentUser$', () => {
    const mockUser = { id: '1', org_id: 'o1', name: 'Test', email: 't@t.com',
                       role: 'admin', timezone: 'UTC', language: 'en',
                       calendar_pref: 'gregorian', is_active: true } as const;
    let emitted: unknown;
    service.login('t@t.com', 'pass').subscribe(u => (emitted = u));

    const req = http.expectOne(`${BASE}/auth/login`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ email: 't@t.com', password: 'pass' });
    req.flush({ user: mockUser });

    expect(emitted).toEqual(mockUser);
    expect(service.currentUser).toEqual(mockUser);
  });

  it('me() rehydrates currentUser$ on success', () => {
    const mockUser = { id: '2', org_id: 'o1', name: 'Me', email: 'm@t.com',
                       role: 'member', timezone: 'UTC', language: 'en',
                       calendar_pref: 'gregorian', is_active: true } as const;
    let result: unknown = 'not set';
    service.me().subscribe(u => (result = u));

    http.expectOne(`${BASE}/users/me`).flush({ data: mockUser });

    expect(result).toEqual(mockUser);
    expect(service.currentUser).toEqual(mockUser);
  });

  it('me() returns null and clears currentUser$ on 401', () => {
    let result: unknown = 'not set';
    service.me().subscribe(u => (result = u));

    http.expectOne(`${BASE}/users/me`).flush('Unauthorized', {
      status: 401, statusText: 'Unauthorized',
    });

    expect(result).toBeNull();
    expect(service.isAuthenticated).toBeFalse();
  });

  it('isAuthenticated$ reflects user state changes', () => {
    const states: boolean[] = [];
    service.isAuthenticated$.subscribe(v => states.push(v));

    service['_user$'].next({
      id: '3', org_id: 'o1', name: 'U', email: 'u@t.com',
      role: 'viewer', timezone: 'UTC', language: 'en',
      calendar_pref: 'gregorian', is_active: true,
    });
    service.clearSession();

    expect(states).toEqual([false, true, false]);
  });
});
