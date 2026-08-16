import { Routes } from '@angular/router';

export const EDITOR_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./book-manager/book-manager.component').then((m) => m.BookManagerComponent),
  },
  {
    path: ':bookId',
    loadComponent: () =>
      import('./chapter-manager/chapter-manager.component').then((m) => m.ChapterManagerComponent),
  },
  {
    path: ':bookId/:chapterId',
    loadComponent: () =>
      import('./chapter-editor/chapter-editor.component').then((m) => m.ChapterEditorComponent),
  },
];
