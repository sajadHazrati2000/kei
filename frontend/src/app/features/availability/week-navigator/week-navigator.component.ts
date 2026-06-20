import { Component, input, output } from '@angular/core';
import { DatePipe } from '@angular/common';
import { inject } from '@angular/core';
import { KeiThemeService } from '../../../core/services/kei-theme.service';
import { KeiIconComponent } from '../../../shared/components/icon/kei-icon.component';
import { JalaliPipe } from '../../../shared/pipes/jalali.pipe';

@Component({
  selector: 'app-week-navigator',
  standalone: true,
  imports: [DatePipe, KeiIconComponent, JalaliPipe],
  template: `
    <div class="week-nav" [attr.dir]="'ltr'">
      <button
        class="week-nav__btn"
        (click)="prev()"
        aria-label="Previous week"
        type="button"
      >
        <kei-icon name="chevron-left" [size]="16" />
      </button>

      <div class="week-nav__label">
        <span class="week-nav__gregorian">
          {{ rangeLabel() }}
        </span>
        @if (isRTL()) {
          <span class="week-nav__jalali">
            ({{ weekStart() | jalali:'d MMM' }} – {{ weekEndDate() | jalali:'d MMM yyyy' }})
          </span>
        }
      </div>

      <button
        class="week-nav__btn"
        (click)="next()"
        aria-label="Next week"
        type="button"
      >
        <kei-icon name="chevron-right" [size]="16" />
      </button>
    </div>
  `,
  styleUrl: './week-navigator.component.scss',
})
export class WeekNavigatorComponent {
  private readonly theme = inject(KeiThemeService);

  weekStart  = input.required<Date>();
  weekChange = output<Date>();

  protected isRTL = () => this.theme.isRTL;

  protected weekEndDate(): Date {
    const d = new Date(this.weekStart());
    d.setDate(d.getDate() + 6);
    return d;
  }

  protected rangeLabel(): string {
    const start = this.weekStart();
    const end   = this.weekEndDate();
    const fmt   = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' });
    const year  = new Intl.DateTimeFormat('en-US', { year: 'numeric' });
    return `${fmt.format(start)} – ${fmt.format(end)}, ${year.format(end)}`;
  }

  protected prev(): void {
    const d = new Date(this.weekStart());
    d.setDate(d.getDate() - 7);
    this.weekChange.emit(d);
  }

  protected next(): void {
    const d = new Date(this.weekStart());
    d.setDate(d.getDate() + 7);
    this.weekChange.emit(d);
  }
}
