import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { AdminCharacterSummary, AdminUniverseSummary, EditorContentService } from '../editor-content.service';

@Component({
  selector: 'app-character-manager',
  standalone: true,
  imports: [FormsModule, RouterLink],
  templateUrl: './character-manager.component.html',
  styleUrl: './character-manager.component.scss',
})
export class CharacterManagerComponent {
  private readonly content = inject(EditorContentService);

  readonly characters = signal<AdminCharacterSummary[]>([]);
  readonly universes = signal<AdminUniverseSummary[]>([]);
  readonly error = signal<string | null>(null);

  readonly newName = signal('');
  readonly newUniverseId = signal<number | null>(null);

  constructor() {
    void this.load();
  }

  private async load(): Promise<void> {
    try {
      const [characters, universes] = await Promise.all([
        firstValueFrom(this.content.listCharacters()),
        firstValueFrom(this.content.listUniverses()),
      ]);
      this.characters.set(characters);
      this.universes.set(universes);
    } catch {
      this.error.set('Failed to load characters.');
    }
  }

  async create(): Promise<void> {
    const name = this.newName().trim();
    if (!name) return;
    try {
      await firstValueFrom(this.content.createCharacter({ name, universeId: this.newUniverseId() }));
      this.newName.set('');
      this.newUniverseId.set(null);
      this.error.set(null);
      await this.load();
    } catch {
      this.error.set('Failed to create character.');
    }
  }

  async remove(character: AdminCharacterSummary): Promise<void> {
    if (!confirm(`Delete "${character.name}"? This can't be undone.`)) return;
    await firstValueFrom(this.content.deleteCharacter(character.id));
    await this.load();
  }

  async move(index: number, direction: -1 | 1): Promise<void> {
    const characters = this.characters();
    const target = index + direction;
    if (target < 0 || target >= characters.length) return;
    const reordered = [...characters];
    [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
    this.characters.set(reordered);
    await firstValueFrom(this.content.reorderCharacters(reordered.map((c) => c.id)));
  }
}
