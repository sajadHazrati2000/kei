import { Routes } from '@angular/router';
import { AppShellComponent } from './shared/components/shell/app-shell.component';
import { authGuard }   from './features/auth/guards/auth.guard';
import { noAuthGuard } from './features/auth/guards/no-auth.guard';
import { setupGuard }  from './features/auth/guards/setup.guard';

export const routes: Routes = [
  // Redirect bare root to /login (exact match only)
  { path: '', redirectTo: 'login', pathMatch: 'full' },

  // ── Public auth pages — no shell ─────────────────────────────────────────
  {
    path: 'login',
    canActivate: [noAuthGuard],
    loadComponent: () =>
      import('./features/auth/login/login.component').then(m => m.LoginComponent),
  },
  {
    path: 'setup',
    canActivate: [setupGuard],
    loadComponent: () =>
      import('./features/auth/setup/setup-wizard.component').then(m => m.SetupWizardComponent),
  },

  // ── Authenticated shell layout ────────────────────────────────────────────
  // AppShellComponent acts as the layout wrapper; its router-outlet
  // renders the active child route inside the main content area.
  {
    path: '',
    component: AppShellComponent,
    canActivate: [authGuard],
    children: [
      {
        path: 'team',
        loadChildren: () =>
          import('./features/dashboard/dashboard.routes').then(m => m.dashboardRoutes),
      },
      {
        path: 'availability',
        loadChildren: () =>
          import('./features/availability/availability.routes').then(m => m.availabilityRoutes),
      },
      {
        path: 'settings',
        loadChildren: () =>
          import('./features/settings/settings.routes').then(m => m.settingsRoutes),
      },
      { path: '**', redirectTo: 'team' },
    ],
  },
];
