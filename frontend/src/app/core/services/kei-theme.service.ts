import { Injectable, inject } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import { TranslateService } from '@ngx-translate/core';

export type KeiLocale = 'en-US' | 'fa-IR';

@Injectable({ providedIn: 'root' })
export class KeiThemeService {
  private readonly translate = inject(TranslateService);

  private readonly _locale$ = new BehaviorSubject<KeiLocale>(
    (localStorage.getItem('kei_locale') as KeiLocale | null) ?? 'en-US'
  );

  /** Observable stream of the active locale. */
  readonly currentLocale$ = this._locale$.asObservable();

  get currentLocale(): KeiLocale {
    return this._locale$.value;
  }

  get isRTL(): boolean {
    return this._locale$.value === 'fa-IR';
  }

  constructor() {
    // Apply saved preference on startup — sets dir, lang, and ngx-translate lang.
    this.applyLocale(this._locale$.value);
  }

  /** Flip between en-US and fa-IR, persisting the choice to localStorage. */
  toggleLocale(): void {
    this.setLocale(this._locale$.value === 'en-US' ? 'fa-IR' : 'en-US');
  }

  /** Explicitly set a locale and apply all side-effects. */
  setLocale(locale: KeiLocale): void {
    if (this._locale$.value === locale) return;
    this._locale$.next(locale);
    localStorage.setItem('kei_locale', locale);
    this.applyLocale(locale);
  }

  private applyLocale(locale: KeiLocale): void {
    const isRTL = locale === 'fa-IR';
    const lang  = isRTL ? 'fa' : 'en';
    document.documentElement.dir  = isRTL ? 'rtl' : 'ltr';
    document.documentElement.lang = lang;
    this.translate.use(lang);
  }
}
