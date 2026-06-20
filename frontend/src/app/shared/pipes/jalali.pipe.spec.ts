import { TestBed }          from '@angular/core/testing';
import { TranslateModule }  from '@ngx-translate/core';
import { JalaliPipe }       from './jalali.pipe';
import { KeiThemeService }  from '../../core/services/kei-theme.service';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

// June 16 2026 in Jalali = 26 Khordad 1405
const TEST_DATE = new Date(2026, 5, 16); // month is 0-indexed

describe('JalaliPipe', () => {
  let pipe:  JalaliPipe;
  let theme: KeiThemeService;

  beforeEach(() => {
    localStorage.clear();
    document.documentElement.dir  = 'ltr';
    document.documentElement.lang = 'en';

    TestBed.configureTestingModule({
      imports: [TranslateModule.forRoot()],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    theme = TestBed.inject(KeiThemeService);
    pipe  = TestBed.runInInjectionContext(() => new JalaliPipe());
  });

  it('transforms date to Jalali string when locale is fa-IR', () => {
    theme.setLocale('fa-IR');
    const result = pipe.transform(TEST_DATE, 'yyyy/MM/dd');
    expect(result).toBe('1405/03/26');
  });

  it('returns Gregorian formatted date when locale is en-US', () => {
    theme.setLocale('en-US');
    const result = pipe.transform(TEST_DATE, 'd MMM');
    // Should contain the day number
    expect(result).toContain('16');
  });

  it('returns empty string for null input', () => {
    expect(pipe.transform(null)).toBe('');
    expect(pipe.transform(undefined)).toBe('');
  });
});
