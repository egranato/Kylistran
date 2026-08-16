import { Routes } from '@angular/router';

export const LIBRARY_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./book-list/book-list.component').then((m) => m.BookListComponent),
  },
  {
    path: ':bookSlug',
    loadComponent: () =>
      import('./chapter-list/chapter-list.component').then((m) => m.ChapterListComponent),
  },
  {
    path: ':bookSlug/:chapterSlug',
    loadComponent: () =>
      import('./chapter-reader/chapter-reader.component').then((m) => m.ChapterReaderComponent),
  },
];
