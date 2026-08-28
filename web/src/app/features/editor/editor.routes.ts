import { Routes } from '@angular/router';

export const EDITOR_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./book-manager/book-manager.component').then((m) => m.BookManagerComponent),
  },
  {
    path: 'universes',
    loadComponent: () =>
      import('./universe-manager/universe-manager.component').then((m) => m.UniverseManagerComponent),
  },
  {
    path: 'characters',
    loadComponent: () =>
      import('./character-manager/character-manager.component').then((m) => m.CharacterManagerComponent),
  },
  {
    path: 'characters/:characterId',
    loadComponent: () =>
      import('./character-editor/character-editor.component').then((m) => m.CharacterEditorComponent),
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
