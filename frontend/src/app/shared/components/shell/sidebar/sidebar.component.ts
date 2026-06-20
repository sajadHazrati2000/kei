import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { KeiIconComponent } from '../../icon/kei-icon.component';
import { KeiThemeService } from '../../../../core/services/kei-theme.service';

interface NavItem {
  icon: string;
  route: string;
  label: string;
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, KeiIconComponent],
  template: `
    <nav class="sidebar" role="navigation" aria-label="Main navigation">
      <!-- Logo mark -->
      <div class="sidebar__logo" aria-label="Kei">كي</div>

      <!-- Primary navigation -->
      <ul class="sidebar__nav" role="list">
        @for (item of navItems; track item.route) {
          <li>
            <a
              class="sidebar__btn"
              [routerLink]="item.route"
              routerLinkActive="sidebar__btn--active"
              [attr.aria-label]="item.label"
              [title]="item.label"
            >
              <kei-icon [name]="item.icon" [size]="20" />
            </a>
          </li>
        }
      </ul>

      <!-- Bottom: language toggle + user avatar -->
      <div class="sidebar__bottom">
        <button
          class="sidebar__btn"
          (click)="theme.toggleLocale()"
          [title]="theme.isRTL ? 'Switch to English' : 'تغییر به فارسی'"
          aria-label="Toggle language"
        >
          <kei-icon name="language" [size]="20" />
        </button>

        <a
          class="sidebar__avatar"
          routerLink="/settings"
          aria-label="Profile settings"
          title="Profile"
        >
          <kei-icon name="user" [size]="18" />
        </a>
      </div>
    </nav>
  `,
  styleUrl: './sidebar.component.scss',
})
export class SidebarComponent {
  protected readonly theme = inject(KeiThemeService);

  protected readonly navItems: NavItem[] = [
    { icon: 'layout-dashboard', route: '/dashboard',    label: 'Dashboard' },
    { icon: 'calendar',         route: '/availability', label: 'My Availability' },
    { icon: 'users',            route: '/team',         label: 'Team' },
    { icon: 'settings',         route: '/settings',     label: 'Settings' },
  ];
}
