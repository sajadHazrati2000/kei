import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from './sidebar/sidebar.component';
import { TopbarComponent } from './topbar/topbar.component';
import { MainContentComponent } from './main-content/main-content.component';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [RouterOutlet, SidebarComponent, TopbarComponent, MainContentComponent],
  template: `
    <div class="kei-shell">
      <app-sidebar />
      <div class="kei-shell__right">
        <app-topbar title="Kei" />
        <app-main-content>
          <router-outlet />
        </app-main-content>
      </div>
    </div>
  `,
  styleUrl: './app-shell.component.scss',
})
export class AppShellComponent {}
