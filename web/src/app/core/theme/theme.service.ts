import { Injectable, effect, signal } from '@angular/core';

export type Theme = 'dark' | 'light';

const STORAGE_KEY = 'kylistran:theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  readonly theme = signal<Theme>(this.readStoredTheme() ?? 'dark');

  constructor() {
    effect(() => {
      document.documentElement.dataset['theme'] = this.theme();
    });
  }

  toggle(): void {
    this.setTheme(this.theme() === 'dark' ? 'light' : 'dark');
  }

  setTheme(theme: Theme): void {
    this.theme.set(theme);
    localStorage.setItem(STORAGE_KEY, theme);
  }

  private readStoredTheme(): Theme | null {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === 'dark' || stored === 'light' ? stored : null;
  }
}
