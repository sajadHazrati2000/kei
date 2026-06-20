import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { provideRouter }            from '@angular/router';
import { provideHttpClient }        from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideAnimations }        from '@angular/platform-browser/animations';
import { TranslateModule }          from '@ngx-translate/core';
import { of }                       from 'rxjs';
import { SetupWizardComponent }     from './setup-wizard.component';
import { AuthService }              from '../../../core/auth/auth.service';

const mockUser = {
  id: '1', org_id: 'o1', name: 'Admin', email: 'a@a.com',
  role: 'admin' as const, timezone: 'UTC', language: 'en',
  calendar_pref: 'gregorian' as const, is_active: true,
};

describe('SetupWizardComponent', () => {
  let fixture: ComponentFixture<SetupWizardComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj('AuthService', ['setup', 'isSetupDone'], {
      currentUser: null, isAuthenticated: false,
    });

    await TestBed.configureTestingModule({
      imports: [SetupWizardComponent, TranslateModule.forRoot()],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        provideAnimations(),
        { provide: AuthService, useValue: authSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(SetupWizardComponent);
    fixture.detectChanges();
  });

  it('renders step 1 by default', () => {
    expect(fixture.componentInstance['step']()).toBe(1);
    const title = fixture.nativeElement.querySelector('.setup-title');
    expect(title).toBeTruthy();
  });

  it('Next is blocked when step 1 form is invalid', () => {
    // orgName is empty → form invalid
    fixture.componentInstance['orgForm'].controls.orgName.setValue('');
    fixture.componentInstance['next']();
    expect(fixture.componentInstance['step']()).toBe(1);
  });

  it('step advances to 2 when step 1 form is valid', fakeAsync(() => {
    fixture.componentInstance['orgForm'].setValue({
      orgName:      'Acme',
      timezone:     'UTC',
      workingStart: '09:00',
      workingEnd:   '17:00',
    });
    fixture.componentInstance['next']();
    tick();
    expect(fixture.componentInstance['step']()).toBe(2);
  }));
});
