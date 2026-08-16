import { inject } from '@angular/core';
import { CanMatchFn, Router } from '@angular/router';
import { EditorAuthService } from './editor-auth.service';

export const editorAuthGuard: CanMatchFn = (_route, segments) => {
  const auth = inject(EditorAuthService);
  if (auth.isAuthenticated()) {
    return true;
  }
  const router = inject(Router);
  return router.createUrlTree(['/editor/login'], {
    queryParams: { redirect: '/' + segments.map((segment) => segment.path).join('/') },
  });
};
