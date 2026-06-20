import { Routes } from '@angular/router';

export const availabilityRoutes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./my-week/my-week.component').then(m => m.MyWeekComponent),
  },
];
