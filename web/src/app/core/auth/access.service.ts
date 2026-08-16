import { Injectable, signal } from '@angular/core';
import { ACCESS_PASSWORD_HASH_HEX, ACCESS_STORAGE_KEY } from './access-config';

@Injectable({ providedIn: 'root' })
export class AccessService {
  readonly unlocked = signal(localStorage.getItem(ACCESS_STORAGE_KEY) === 'true');

  async unlock(candidate: string): Promise<boolean> {
    const digest = await this.sha256Hex(candidate);
    const matches = digest === ACCESS_PASSWORD_HASH_HEX;
    if (matches) {
      this.unlocked.set(true);
      localStorage.setItem(ACCESS_STORAGE_KEY, 'true');
    }
    return matches;
  }

  private async sha256Hex(value: string): Promise<string> {
    const bytes = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(value));
    return Array.from(new Uint8Array(bytes))
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');
  }
}
