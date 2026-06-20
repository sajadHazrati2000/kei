import { Component, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { KeiThemeService } from './core/services/kei-theme.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  // AppShellComponent is now a layout route; AppComponent is just the root outlet.
  template: `<router-outlet />`,
})
export class AppComponent {
  // Injecting KeiThemeService ensures locale/dir is applied at bootstrap.
  private readonly _theme = inject(KeiThemeService);
}
