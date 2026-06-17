import { Component, OnInit, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  template: '<router-outlet />',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  private translate = inject(TranslateService);

  ngOnInit(): void {
    this.translate.addLangs(['en', 'fa']);
    this.translate.setDefaultLang('en');

    const saved = localStorage.getItem('kei_lang') ?? 'en';
    this.setLanguage(saved);

    this.translate.onLangChange.subscribe(({ lang }) => {
      localStorage.setItem('kei_lang', lang);
      this.applyDir(lang);
    });
  }

  private setLanguage(lang: string): void {
    this.translate.use(lang);
    this.applyDir(lang);
  }

  private applyDir(lang: string): void {
    const dir = lang === 'fa' ? 'rtl' : 'ltr';
    document.documentElement.dir = dir;
    document.documentElement.lang = lang;
  }
}
