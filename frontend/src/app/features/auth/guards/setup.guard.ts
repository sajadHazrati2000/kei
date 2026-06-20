import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map, catchError } from 'rxjs/operators';
import { of } from 'rxjs';
import { AuthService } from '../../../core/auth/auth.service';

/** Protects /setup:
 *  - Already authenticated → /team (setup is done and user is logged in)
 *  - Setup done but not logged in → /login
 *  - Setup not done → allow access to the wizard
 */
export const setupGuard: CanActivateFn = () => {
  const auth   = inject(AuthService);
  const router = inject(Router);

  if (auth.isAuthenticated) return router.createUrlTree(['/team']);

  return auth.isSetupDone().pipe(
    map(done => done ? router.createUrlTree(['/login']) : true),
    catchError(() => of(true)) // on error, allow the wizard (it will handle failures)
  );
};
