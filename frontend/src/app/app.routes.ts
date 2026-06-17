import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'auth',
    loadChildren: () => import('./features/auth/auth.routes').then(m => m.authRoutes),
  },
  {
    path: 'dashboard',
    loadChildren: () => import('./features/dashboard/dashboard.routes').then(m => m.dashboardRoutes),
  },
  {
    path: 'availability',
    loadChildren: () => import('./features/availability/availability.routes').then(m => m.availabilityRoutes),
  },
  {
    path: 'settings',
    loadChildren: () => import('./features/settings/settings.routes').then(m => m.settingsRoutes),
  },
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: '**', redirectTo: 'dashboard' },
];
