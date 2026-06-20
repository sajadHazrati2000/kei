import { Component } from '@angular/core';

@Component({
  selector: 'app-main-content',
  standalone: true,
  template: `
    <main class="main-content">
      <ng-content />
    </main>
  `,
  styleUrl: './main-content.component.scss',
})
export class MainContentComponent {}
