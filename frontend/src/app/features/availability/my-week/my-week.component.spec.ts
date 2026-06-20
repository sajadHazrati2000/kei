import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { provideRouter }            from '@angular/router';
import { provideHttpClient }        from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideAnimations }        from '@angular/platform-browser/animations';
import { TranslateModule }          from '@ngx-translate/core';
import { of }                       from 'rxjs';
import { MyWeekComponent }          from './my-week.component';
import { AvailabilityService }      from '../../../core/availability/availability.service';
import { AuthService, User }        from '../../../core/auth/auth.service';
import { Slot }                     from '../../../core/availability/availability.models';

const MOCK_USER: User = {
  id: 'u1', org_id: 'o1', name: 'Test', email: 't@t.com',
  role: 'admin', timezone: 'UTC', language: 'en',
  calendar_pref: 'gregorian', is_active: true,
};

function makeSlot(status: 'free' | 'busy' | 'focus', id = 's1'): Slot {
  return {
    id, user_id: 'u1',
    from: '2026-06-16T09:00:00Z', to: '2026-06-16T10:00:00Z',
    status, is_override: false, created_at: '2026-06-01T00:00:00Z',
  };
}

describe('MyWeekComponent', () => {
  let fixture:  ComponentFixture<MyWeekComponent>;
  let availSpy: jasmine.SpyObj<AvailabilityService>;
  let authSpy:  jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    availSpy = jasmine.createSpyObj('AvailabilityService', [
      'getMySlots', 'setSlots', 'createSlot', 'deleteSlot',
    ]);
    availSpy.getMySlots.and.returnValue(of([]));

    authSpy = jasmine.createSpyObj('AuthService', [], {
      currentUser: MOCK_USER, isAuthenticated: true,
    });

    await TestBed.configureTestingModule({
      imports: [MyWeekComponent, TranslateModule.forRoot()],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        provideAnimations(),
        { provide: AvailabilityService, useValue: availSpy },
        { provide: AuthService,         useValue: authSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(MyWeekComponent);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('renders 7 day buttons in the week strip', () => {
    const buttons = fixture.nativeElement.querySelectorAll('.day-btn');
    expect(buttons.length).toBe(7);
  });

  it('day dot reflects slot status — busy slots show dot--busy', fakeAsync(() => {
    // Pin the week to Jun 15 (Sun) so Jun 16 (Tue) is day index 2 (0-based)
    // and directly inject a busy slot for Jun 16 into the component's state.
    const weekStart = new Date('2026-06-15T00:00:00.000Z'); // Sunday
    fixture.componentInstance['weekStart'].set(weekStart);
    fixture.componentInstance['allSlots'].set([makeSlot('busy', 's1')]);
    fixture.detectChanges();

    // Jun 16 is day index 1 in a week starting Jun 15
    const dayButtons = fixture.nativeElement.querySelectorAll('.day-btn');
    const mondayDot = (dayButtons[1] as HTMLElement).querySelector('.day-dot');
    expect(mondayDot?.classList).toContain('day-dot--busy');
  }));

  it('clicking an empty timeline cell shows the slot creator', () => {
    const grid = fixture.nativeElement.querySelector('.timeline-scroll') as HTMLElement;
    expect(grid).toBeTruthy();

    // Simulate click on empty area
    fixture.componentInstance['creatorAt'].set({ hour: 9, minute: 0 });
    fixture.detectChanges();

    const creator = fixture.nativeElement.querySelector('app-slot-creator');
    expect(creator).withContext('slot-creator should appear').toBeTruthy();
  });

  it('clicking an existing slot sets editingSlotId', fakeAsync(() => {
    const slot = makeSlot('free');
    availSpy.getMySlots.and.returnValue(of([slot]));
    fixture.componentInstance['loadSlots' as keyof MyWeekComponent]();
    tick();
    fixture.detectChanges();

    fixture.componentInstance['editSlot'](slot);
    expect(fixture.componentInstance['editingSlotId']()).toBe(slot.id);
  }));

  it('changing week via navigator reloads slots', fakeAsync(() => {
    const newStart = new Date('2026-06-23T00:00:00.000Z');
    availSpy.getMySlots.and.returnValue(of([]));

    fixture.componentInstance['onWeekChange'](newStart);
    tick();

    expect(availSpy.getMySlots).toHaveBeenCalledWith(newStart);
  }));
});
