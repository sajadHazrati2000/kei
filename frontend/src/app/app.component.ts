import { Component, inject } from '@angular/core';
import { AppShellComponent } from './shared/components/shell/app-shell.component';
import { KeiThemeService } from './core/services/kei-theme.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [AppShellComponent],
  template: `<app-shell />`,
})
export class AppComponent {
  // Injecting KeiThemeService here ensures it is instantiated at bootstrap,
  // which restores the saved locale (dir + lang) before the first render.
  private readonly _theme = inject(KeiThemeService);
}
