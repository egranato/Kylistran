import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { environment } from '../../../environments/environment';
import { EditorAuthService } from './editor-auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  if (!req.url.startsWith(environment.apiBaseUrl)) {
    return next(req);
  }

  const token = inject(EditorAuthService).currentToken();
  if (!token) {
    return next(req);
  }

  return next(req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }));
};
