import { inject } from '@angular/core';
import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { Router } from '@angular/router';
import { throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { AuthService } from './auth.service';

// Endpoints where 401 should NOT trigger a redirect — these are called before
// the user is authenticated or are part of the auth flow itself.
const AUTH_BYPASS_PATTERNS = [
  '/auth/login',
  '/auth/setup',
  '/auth/logout',
  '/auth/refresh',
  '/auth/password-reset',
  '/auth/setup/status',
  '/users/me',
];

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth   = inject(AuthService);
  const router = inject(Router);

  // Always attach credentials so httpOnly cookies are sent cross-origin.
  const withCreds = req.clone({ withCredentials: true });

  return next(withCreds).pipe(
    catchError(err => {
      if (
        err instanceof HttpErrorResponse &&
        err.status === 401 &&
        !AUTH_BYPASS_PATTERNS.some(p => req.url.includes(p))
      ) {
        auth.clearSession();
        router.navigate(['/login']);
      }
      return throwError(() => err);
    })
  );
};
