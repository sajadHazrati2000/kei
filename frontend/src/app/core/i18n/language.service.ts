import { Injectable, inject } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

export type Lang = 'en' | 'fa';

@Injectable({ providedIn: 'root' })
export class LanguageService {
  private translate = inject(TranslateService);

  get current(): Lang {
    return (this.translate.currentLang ?? 'en') as Lang;
  }

  switch(lang: Lang): void {
    this.translate.use(lang);
  }
}
