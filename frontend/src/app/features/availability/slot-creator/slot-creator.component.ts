import { Component, input, output, signal } from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { SlotInput, SlotStatus } from '../../../core/availability/availability.models';

const ROW_H = 24;

function toTimeISO(date: Date, hour: number, minute: number): string {
  const d = new Date(date);
  d.setUTCHours(hour, minute, 0, 0);
  return d.toISOString();
}

@Component({
  selector: 'app-slot-creator',
  standalone: true,
  imports: [DecimalPipe, FormsModule, TranslateModule],
  template: `
    <div
      class="slot-creator"
      [style.top.px]="top()"
      role="dialog"
      aria-label="Create availability slot"
    >
      <!-- Type selector -->
      <div class="slot-creator__types">
        @for (t of types; track t.value) {
          <button
            class="slot-creator__type-btn"
            [class.slot-creator__type-btn--active]="selectedType() === t.value"
            [class]="'slot-creator__type-btn--' + t.value"
            (click)="selectedType.set(t.value)"
            type="button"
          >{{ t.label }}</button>
        }
      </div>

      <!-- Optional label -->
      <input
        class="slot-creator__label-input"
        [(ngModel)]="label"
        placeholder="{{ 'AVAILABILITY.LABEL_PLACEHOLDER' | translate }}"
        maxlength="40"
        (keydown.enter)="confirm()"
        (keydown.escape)="cancelled.emit()"
        type="text"
      />

      <!-- Time display -->
      <div class="slot-creator__time">
        {{ startHour() | number:'2.0-0' }}:{{ startMinute() | number:'2.0-0' }} –
        {{ endHour()   | number:'2.0-0' }}:{{ endMinute()   | number:'2.0-0' }}
      </div>

      <!-- Actions -->
      <div class="slot-creator__actions">
        <button class="slot-creator__cancel" (click)="cancelled.emit()" type="button">
          {{ 'SETUP.BACK' | translate }}
        </button>
        <button class="slot-creator__confirm" (click)="confirm()" type="button">
          {{ 'AVAILABILITY.SAVE' | translate }}
        </button>
      </div>
    </div>
  `,
  styleUrl: './slot-creator.component.scss',
})
export class SlotCreatorComponent {
  forDate    = input.required<Date>();
  startHour  = input.required<number>();
  startMinute = input<number>(0);

  created   = output<SlotInput>();
  cancelled = output<void>();

  protected selectedType = signal<SlotStatus>('free');
  protected label        = '';

  protected endHour   = () => this.startMinute() === 30 ? this.startHour() + 1 : this.startHour();
  protected endMinute = () => this.startMinute() === 30 ? 0 : 30;

  protected top = () => (this.startHour() * 2 + this.startMinute() / 30) * ROW_H;

  protected readonly types = [
    { value: 'free'  as SlotStatus, label: 'Free' },
    { value: 'focus' as SlotStatus, label: 'Focus' },
    { value: 'busy'  as SlotStatus, label: 'Busy' },
  ];

  protected confirm(): void {
    const from = toTimeISO(this.forDate(), this.startHour(), this.startMinute());
    const to   = toTimeISO(this.forDate(), this.endHour(), this.endMinute());
    this.created.emit({ from, to, status: this.selectedType() });
  }
}
