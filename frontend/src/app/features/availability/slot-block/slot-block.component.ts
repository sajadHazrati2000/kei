import { Component, computed, inject, input, output } from '@angular/core';
import { DomSanitizer } from '@angular/platform-browser';
import { Slot } from '../../../core/availability/availability.models';
import { KeiIconComponent } from '../../../shared/components/icon/kei-icon.component';

const ROW_H = 24; // px per 30-min row

/** Maps a slot's ISO `from` time to a pixel top offset in the timeline grid. */
export function slotTop(isoFrom: string): number {
  const d = new Date(isoFrom);
  return (d.getUTCHours() * 2 + Math.floor(d.getUTCMinutes() / 30)) * ROW_H;
}

/** Maps a slot's ISO `from`→`to` duration to a pixel height. */
export function slotHeight(isoFrom: string, isoTo: string): number {
  const ms = new Date(isoTo).getTime() - new Date(isoFrom).getTime();
  const halfHours = ms / (30 * 60 * 1000);
  return Math.max(halfHours, 1) * ROW_H;
}

const ICON_MAP: Record<string, string> = {
  free:  'circle-check',
  busy:  'circle-x',
  focus: 'focus-2',
};

@Component({
  selector: 'app-slot-block',
  standalone: true,
  imports: [KeiIconComponent],
  template: `
    <div
      class="slot-block"
      [class]="'slot-block--' + slot().status"
      [class.slot-block--editing]="isEditing()"
      [style.top.px]="top()"
      [style.height.px]="height()"
      role="button"
      tabindex="0"
      [attr.aria-label]="slot().status + ' slot'"
      (click)="edit.emit(slot())"
      (keydown.enter)="edit.emit(slot())"
    >
      <!-- Resize handle — top -->
      <div
        class="slot-block__handle slot-block__handle--top"
        (pointerdown)="onResizeDown($event, 'top')"
        aria-hidden="true"
      ></div>

      <div class="slot-block__content">
        <kei-icon [name]="icon()" [size]="12" class="slot-block__icon" />
        @if (slot().status === 'busy') {
          <span class="slot-block__label">Busy</span>
        } @else if (slot().status === 'focus') {
          <span class="slot-block__label">Focus</span>
        }
      </div>

      @if (isEditing()) {
        <button
          class="slot-block__action slot-block__action--delete"
          (click)="$event.stopPropagation(); delete.emit(slot())"
          aria-label="Delete slot"
          type="button"
        >
          <kei-icon name="trash" [size]="12" />
        </button>
      }

      <!-- Resize handle — bottom -->
      <div
        class="slot-block__handle slot-block__handle--bottom"
        (pointerdown)="onResizeDown($event, 'bottom')"
        aria-hidden="true"
      ></div>
    </div>
  `,
  styleUrl: './slot-block.component.scss',
})
export class SlotBlockComponent {
  slot      = input.required<Slot>();
  isEditing = input<boolean>(false);

  edit   = output<Slot>();
  delete = output<Slot>();
  resize = output<{ slot: Slot; from: string; to: string }>();

  protected top    = computed(() => slotTop(this.slot().from));
  protected height = computed(() => slotHeight(this.slot().from, this.slot().to));
  protected icon   = computed(() => ICON_MAP[this.slot().status] ?? 'circle-check');

  /** Pointer-event based resize — works for mouse and touch. */
  protected onResizeDown(event: PointerEvent, handle: 'top' | 'bottom'): void {
    event.preventDefault();
    event.stopPropagation();

    const el = event.currentTarget as HTMLElement;
    el.setPointerCapture(event.pointerId);

    const startY     = event.clientY;
    const slotFromMs = new Date(this.slot().from).getTime();
    const slotToMs   = new Date(this.slot().to).getTime();
    const SNAP_MS    = 30 * 60 * 1000;

    let pendingFrom = this.slot().from;
    let pendingTo   = this.slot().to;

    const onMove = (e: PointerEvent): void => {
      const delta      = e.clientY - startY;
      const snaps      = Math.round(delta / ROW_H);
      const deltaMs    = snaps * SNAP_MS;

      if (handle === 'bottom') {
        const newToMs = Math.max(slotToMs + deltaMs, slotFromMs + SNAP_MS);
        pendingTo = new Date(newToMs).toISOString();
      } else {
        const newFromMs = Math.min(slotFromMs + deltaMs, slotToMs - SNAP_MS);
        pendingFrom = new Date(newFromMs).toISOString();
      }
    };

    const onUp = (): void => {
      el.releasePointerCapture(event.pointerId);
      el.removeEventListener('pointermove', onMove as EventListener);
      el.removeEventListener('pointerup', onUp);
      this.resize.emit({ slot: this.slot(), from: pendingFrom, to: pendingTo });
    };

    el.addEventListener('pointermove', onMove as EventListener);
    el.addEventListener('pointerup', onUp);
  }
}
