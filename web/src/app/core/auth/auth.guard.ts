import { inject } from '@angular/core';
import { CanMatchFn, Router } from '@angular/router';
import { AccessService } from './access.service';

export const authGuard: CanMatchFn = (_route, segments) => {
  const access = inject(AccessService);
  if (access.unlocked()) {
    return true;
  }
  const router = inject(Router);
  return router.createUrlTree(['/'], {
    queryParams: { redirect: '/' + segments.map((segment) => segment.path).join('/') },
  });
};
