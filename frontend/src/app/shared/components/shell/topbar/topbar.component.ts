import { Component, input } from '@angular/core';

@Component({
  selector: 'app-topbar',
  standalone: true,
  template: `
    <header class="topbar">
      <div class="topbar__titles">
        <h1 class="topbar__title">{{ title() }}</h1>
        @if (subtitle()) {
          <span class="topbar__subtitle">{{ subtitle() }}</span>
        }
      </div>
      <ng-content />
    </header>
  `,
  styleUrl: './topbar.component.scss',
})
export class TopbarComponent {
  title    = input<string>('Kei');
  subtitle = input<string>('');
}
