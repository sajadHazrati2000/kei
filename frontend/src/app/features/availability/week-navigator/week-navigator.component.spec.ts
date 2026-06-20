import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule }           from '@ngx-translate/core';
import { provideHttpClient }         from '@angular/common/http';
import { provideHttpClientTesting }  from '@angular/common/http/testing';
import { provideAnimations }         from '@angular/platform-browser/animations';
import { WeekNavigatorComponent }    from './week-navigator.component';
import { KeiThemeService }           from '../../../core/services/kei-theme.service';

// Monday June 16 2026
const WEEK_START = new Date(2026, 5, 16);

describe('WeekNavigatorComponent', () => {
  let fixture: ComponentFixture<WeekNavigatorComponent>;
  let theme:   KeiThemeService;

  beforeEach(async () => {
    localStorage.clear();

    await TestBed.configureTestingModule({
      imports:   [WeekNavigatorComponent, TranslateModule.forRoot()],
      providers: [
        provideAnimations(),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    theme   = TestBed.inject(KeiThemeService);
    fixture = TestBed.createComponent(WeekNavigatorComponent);
    fixture.componentRef.setInput('weekStart', WEEK_START);
    fixture.detectChanges();
  });

  it('displays the correct Gregorian date range', () => {
    const label = fixture.nativeElement.querySelector('.week-nav__gregorian') as HTMLElement;
    expect(label.textContent).toContain('Jun 16');
    expect(label.textContent).toContain('Jun 22');
  });

  it('prev button emits weekStart minus 7 days', () => {
    const emitted: Date[] = [];
    fixture.componentInstance.weekChange.subscribe((d: Date) => emitted.push(d));
    const btn = fixture.nativeElement.querySelector('[aria-label="Previous week"]') as HTMLButtonElement;
    btn.click();
    expect(emitted.length).toBe(1);
    expect(emitted[0].getDate()).toBe(9); // Jun 9
  });

  it('next button emits weekStart plus 7 days', () => {
    const emitted: Date[] = [];
    fixture.componentInstance.weekChange.subscribe((d: Date) => emitted.push(d));
    const btn = fixture.nativeElement.querySelector('[aria-label="Next week"]') as HTMLButtonElement;
    btn.click();
    expect(emitted.length).toBe(1);
    expect(emitted[0].getDate()).toBe(23); // Jun 23
  });
});
