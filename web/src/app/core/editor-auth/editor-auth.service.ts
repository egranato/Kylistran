import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { environment } from '../../../environments/environment';

const TOKEN_STORAGE_KEY = 'kylistran:editorToken';

@Injectable({ providedIn: 'root' })
export class EditorAuthService {
  private readonly http = inject(HttpClient);

  private readonly token = signal(localStorage.getItem(TOKEN_STORAGE_KEY));
  readonly isAuthenticated = computed(() => this.token() !== null);

  currentToken(): string | null {
    return this.token();
  }

  async login(username: string, password: string): Promise<boolean> {
    try {
      const res = await firstValueFrom(
        this.http.post<{ token: string }>(`${environment.apiBaseUrl}/auth/login`, { username, password }),
      );
      this.token.set(res.token);
      localStorage.setItem(TOKEN_STORAGE_KEY, res.token);
      return true;
    } catch {
      return false;
    }
  }

  logout(): void {
    this.token.set(null);
    localStorage.removeItem(TOKEN_STORAGE_KEY);
  }
}
