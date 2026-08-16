import { Routes } from '@angular/router';
import { authGuard } from './core/auth/auth.guard';
import { editorAuthGuard } from './core/editor-auth/editor-auth.guard';
import { MAINTENANCE_MODE } from './core/maintenance';

const maintenanceRoutes: Routes = [
  {
    path: '**',
    loadComponent: () =>
      import('./features/maintenance/maintenance.component').then((m) => m.MaintenanceComponent),
  },
];

const siteRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./features/gate/gate.component').then((m) => m.GateComponent),
  },
  {
    path: 'library',
    canMatch: [authGuard],
    loadChildren: () => import('./features/library/library.routes').then((m) => m.LIBRARY_ROUTES),
  },
  {
    path: 'editor/login',
    loadComponent: () =>
      import('./features/editor/login/editor-login.component').then((m) => m.EditorLoginComponent),
  },
  {
    path: 'editor',
    canMatch: [editorAuthGuard],
    loadChildren: () => import('./features/editor/editor.routes').then((m) => m.EDITOR_ROUTES),
  },
  { path: '**', redirectTo: 'library' },
];

export const routes: Routes = MAINTENANCE_MODE ? maintenanceRoutes : siteRoutes;
