import { TestBed } from '@angular/core/testing';
import { TranslateModule } from '@ngx-translate/core';
import { KeiThemeService } from './kei-theme.service';

describe('KeiThemeService', () => {
  let service: KeiThemeService;

  beforeEach(() => {
    localStorage.clear();
    // Reset document state before each test
    document.documentElement.dir  = 'ltr';
    document.documentElement.lang = 'en';

    TestBed.configureTestingModule({
      imports: [TranslateModule.forRoot()],
    });
    service = TestBed.inject(KeiThemeService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('defaults to en-US when no preference is stored', () => {
    expect(service.currentLocale).toBe('en-US');
    expect(service.isRTL).toBeFalse();
  });

  it('toggleLocale() switches to fa-IR and applies RTL', () => {
    service.toggleLocale();

    expect(service.currentLocale).toBe('fa-IR');
    expect(service.isRTL).toBeTrue();
    expect(document.documentElement.dir).toBe('rtl');
    expect(document.documentElement.lang).toBe('fa');
  });

  it('toggleLocale() called twice returns to en-US and LTR', () => {
    service.toggleLocale(); // → fa-IR
    service.toggleLocale(); // → en-US

    expect(service.currentLocale).toBe('en-US');
    expect(document.documentElement.dir).toBe('ltr');
    expect(document.documentElement.lang).toBe('en');
  });

  it('persists locale choice to localStorage', () => {
    service.toggleLocale();
    expect(localStorage.getItem('kei_locale')).toBe('fa-IR');
  });

  it('setLocale() is a no-op when locale is already active', () => {
    const spy = spyOn(document.documentElement, 'setAttribute');
    service.setLocale('en-US'); // already en-US — should not fire
    expect(spy).not.toHaveBeenCalled();
  });
});
