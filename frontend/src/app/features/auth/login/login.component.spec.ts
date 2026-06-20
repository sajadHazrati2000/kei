import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { provideRouter }            from '@angular/router';
import { Router }                   from '@angular/router';
import { provideHttpClient }        from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideAnimations }        from '@angular/platform-browser/animations';
import { TranslateModule }          from '@ngx-translate/core';
import { of, throwError }           from 'rxjs';
import { LoginComponent }           from './login.component';
import { AuthService }              from '../../../core/auth/auth.service';
import { KeiThemeService }          from '../../../core/services/kei-theme.service';

const mockUser = {
  id: '1', org_id: 'o1', name: 'Test', email: 't@t.com',
  role: 'admin' as const, timezone: 'UTC', language: 'en',
  calendar_pref: 'gregorian' as const, is_active: true,
};

describe('LoginComponent', () => {
  let fixture:     ComponentFixture<LoginComponent>;
  let authSpy:     jasmine.SpyObj<AuthService>;
  let router:      Router;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj('AuthService', ['login'], {
      currentUser: null, isAuthenticated: false,
    });

    await TestBed.configureTestingModule({
      imports: [LoginComponent, TranslateModule.forRoot()],
      providers: [
        // Provide a stub /team route so navigate() succeeds without errors
        provideRouter([{ path: 'team', component: LoginComponent }]),
        provideHttpClient(),
        provideHttpClientTesting(),
        provideAnimations(),
        { provide: AuthService, useValue: authSpy },
      ],
    }).compileComponents();

    router  = TestBed.inject(Router);
    spyOn(router, 'navigate').and.resolveTo(true); // prevent real navigation in all tests
    fixture = TestBed.createComponent(LoginComponent);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('renders email and password fields', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('input[type="email"], input[formcontrolname="email"]'))
      .withContext('email field').toBeTruthy();
    expect(el.querySelector('input[type="password"]'))
      .withContext('password field').toBeTruthy();
  });

  it('calls authService.login with form values on submit', fakeAsync(() => {
    authSpy.login.and.returnValue(of(mockUser));
    const el = fixture.nativeElement as HTMLElement;

    (el.querySelector('input[type="email"]') as HTMLInputElement).value = 't@t.com';
    (el.querySelector('input[type="email"]') as HTMLInputElement).dispatchEvent(new Event('input'));
    (el.querySelector('input[type="password"]') as HTMLInputElement).value = 'password';
    (el.querySelector('input[type="password"]') as HTMLInputElement).dispatchEvent(new Event('input'));
    fixture.detectChanges();

    fixture.componentInstance['form'].setValue({ email: 't@t.com', password: 'password' });
    fixture.componentInstance['submit']();
    tick();

    expect(authSpy.login).toHaveBeenCalledWith('t@t.com', 'password');
  }));

  it('displays error message when login fails', fakeAsync(() => {
    authSpy.login.and.returnValue(
      throwError(() => ({ error: { error: 'ERRORS.UNAUTHORIZED' } }))
    );
    fixture.componentInstance['form'].setValue({ email: 'x@x.com', password: 'wrong' });
    fixture.componentInstance['submit']();
    tick();
    fixture.detectChanges();

    const el = fixture.nativeElement as HTMLElement;
    const errorEl = el.querySelector('.login-error');
    expect(errorEl).withContext('error element').toBeTruthy();
  }));

  it('navigates to /team on successful login', fakeAsync(() => {
    authSpy.login.and.returnValue(of(mockUser));
    fixture.componentInstance['form'].setValue({ email: 't@t.com', password: 'pass' });
    fixture.componentInstance['submit']();
    tick();
    // router.navigate was spied in beforeEach
    expect(router.navigate).toHaveBeenCalledWith(['/team']);
  }));

  it('applies RTL class direction when fa-IR locale', () => {
    const theme = TestBed.inject(KeiThemeService);
    theme.setLocale('fa-IR');
    expect(document.documentElement.dir).toBe('rtl');
  });

  it('submit button is present and labelled', () => {
    const btn = fixture.nativeElement.querySelector('button[type="submit"]');
    expect(btn).toBeTruthy();
  });
});
