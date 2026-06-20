import { TestBed }                  from '@angular/core/testing';
import { provideHttpClient }        from '@angular/common/http';
import { HttpTestingController,
         provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter }            from '@angular/router';
import { AvailabilityService }           from './availability.service';
import { AuthService, User }             from '../auth/auth.service';
import { Slot, UserAvailability }        from './availability.models';

const BASE    = 'http://localhost:8080/api/v1';
const MOCK_ID = 'user-123';
const MOCK_USER: User = {
  id: MOCK_ID, org_id: 'org-1', name: 'Test', email: 't@t.com',
  role: 'admin', timezone: 'UTC', language: 'en',
  calendar_pref: 'gregorian', is_active: true,
};
const WEEK_START = new Date('2026-06-16T00:00:00.000Z');
const MOCK_SLOT: Slot = {
  id: 's1', user_id: MOCK_ID,
  from: '2026-06-16T09:00:00Z', to: '2026-06-16T12:00:00Z',
  status: 'free', is_override: false, created_at: '2026-06-16T00:00:00Z',
};

describe('AvailabilityService', () => {
  let service: AvailabilityService;
  let http:    HttpTestingController;
  let authSpy: jasmine.SpyObj<AuthService>;

  beforeEach(() => {
    authSpy = jasmine.createSpyObj('AuthService', [], {
      currentUser: MOCK_USER,
    });

    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authSpy },
      ],
    });
    service = TestBed.inject(AvailabilityService);
    http    = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('getMySlots() calls GET /availability/:userId with from/to params', () => {
    let result: Slot[] | undefined;
    service.getMySlots(WEEK_START).subscribe(s => (result = s));

    const req = http.expectOne(r =>
      r.url === `${BASE}/availability/${MOCK_ID}` &&
      r.method === 'GET' &&
      r.params.has('from') &&
      r.params.has('to')
    );
    req.flush({ data: [MOCK_SLOT], total: 1 });
    expect(result).toEqual([MOCK_SLOT]);
  });

  it('setSlots() calls PUT /availability/:userId with slot body', () => {
    const input = [{ from: MOCK_SLOT.from, to: MOCK_SLOT.to, status: MOCK_SLOT.status }];
    let result: Slot[] | undefined;
    service.setSlots(WEEK_START, input).subscribe(s => (result = s));

    const req = http.expectOne(`${BASE}/availability/${MOCK_ID}`);
    expect(req.request.method).toBe('PUT');
    expect(req.request.body.slots).toEqual(input);
    req.flush({ data: [MOCK_SLOT], total: 1 });
    expect(result).toEqual([MOCK_SLOT]);
  });

  it('deleteSlot() calls setSlots (PUT) with the slot removed', () => {
    const existing: Slot[] = [MOCK_SLOT, { ...MOCK_SLOT, id: 's2', status: 'busy' }];
    let result: Slot[] | undefined;
    service.deleteSlot(WEEK_START, existing, 's1').subscribe(s => (result = s));

    const req = http.expectOne(`${BASE}/availability/${MOCK_ID}`);
    expect(req.request.method).toBe('PUT');
    // Only s2 should remain in the body
    expect(req.request.body.slots.length).toBe(1);
    expect(req.request.body.slots[0].status).toBe('busy');
    req.flush({ data: [{ ...MOCK_SLOT, id: 's2', status: 'busy' }], total: 1 });
    expect(result!.length).toBe(1);
  });

  it('getOverlapWindows() calls GET /team/overlap and returns OverlapResult', () => {
    const mockResult = {
      date: '2026-06-16',
      overlap_start: '09:00',
      overlap_end: '17:00',
      members: [] as UserAvailability[],
    };
    let result: typeof mockResult | undefined;
    service.getOverlapWindows(WEEK_START).subscribe(r => (result = r as typeof mockResult));

    const req = http.expectOne(r =>
      r.url === `${BASE}/team/overlap` && r.params.get('date') === '2026-06-16'
    );
    expect(req.request.method).toBe('GET');
    req.flush(mockResult);
    expect(result).toEqual(mockResult);
  });
});
