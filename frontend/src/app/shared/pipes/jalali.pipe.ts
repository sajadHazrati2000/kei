import { inject, Pipe, PipeTransform } from '@angular/core';
import { format } from 'date-fns-jalali';
import { KeiThemeService } from '../../core/services/kei-theme.service';

/**
 * Transforms a Date to a Jalali calendar string when locale is fa-IR,
 * or a Gregorian string (via Intl) when locale is en-US.
 *
 * Usage:
 *   {{ date | jalali }}               → "26 خرداد" (fa-IR) | "Jun 16" (en-US)
 *   {{ date | jalali:'yyyy/MM/dd' }}  → "1405/03/26" (fa-IR) | "2026/06/16" (en-US)
 */
@Pipe({
  name: 'jalali',
  standalone: true,
  pure: false, // re-evaluates when locale changes
})
export class JalaliPipe implements PipeTransform {
  private readonly theme = inject(KeiThemeService);

  transform(value: Date | string | null | undefined, formatStr = 'd MMM'): string {
    if (value == null) return '';
    const date = value instanceof Date ? value : new Date(value as string);
    if (isNaN(date.getTime())) return '';

    if (this.theme.currentLocale === 'fa-IR') {
      try {
        return format(date, formatStr); // date-fns-jalali
      } catch {
        return '';
      }
    }

    // Gregorian passthrough — map format string to Intl options
    return this.gregorian(date, formatStr);
  }

  private gregorian(date: Date, fmt: string): string {
    // Best-effort Gregorian format matching date-fns-jalali token conventions
    if (fmt.includes('yyyy') || fmt.includes('MM') || fmt.includes('dd')) {
      return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
      }).format(date).replace(/\//g, '/');
    }
    // Default: "d MMM" → "16 Jun"
    return new Intl.DateTimeFormat('en-US', {
      day: 'numeric',
      month: 'short',
    }).format(date);
  }
}
