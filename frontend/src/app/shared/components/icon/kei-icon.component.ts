import { Component, computed, inject, input } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { KEI_ICONS } from '../../icons/kei-icons';

/**
 * Renders a Tabler outline icon by name as inline SVG.
 * Sizes are driven by the [size] input (default 20 px).
 * Colour inherits from `currentColor` so CSS handles it.
 *
 * Usage: <kei-icon name="calendar" />
 *        <kei-icon name="settings" [size]="16" />
 */
@Component({
  selector: 'kei-icon',
  standalone: true,
  template: `<span [innerHTML]="svg()" aria-hidden="true"></span>`,
  styles: [`
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }
    span {
      display: contents;
    }
  `],
})
export class KeiIconComponent {
  private readonly sanitizer = inject(DomSanitizer);

  name = input.required<string>();
  size = input<number>(20);

  protected svg = computed((): SafeHtml => {
    const raw = KEI_ICONS[this.name()] ?? '';
    // Replace the default 20px dimensions with the requested size.
    const sized = raw
      .replace(/width="\d+"/, `width="${this.size()}"`)
      .replace(/height="\d+"/, `height="${this.size()}"`);
    return this.sanitizer.bypassSecurityTrustHtml(sized);
  });
}
