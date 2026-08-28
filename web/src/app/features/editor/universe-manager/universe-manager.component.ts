import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { AdminUniverseSummary, EditorContentService, UniverseInput } from '../editor-content.service';

@Component({
  selector: 'app-universe-manager',
  standalone: true,
  imports: [FormsModule, RouterLink],
  templateUrl: './universe-manager.component.html',
  styleUrl: './universe-manager.component.scss',
})
export class UniverseManagerComponent {
  private readonly content = inject(EditorContentService);

  readonly universes = signal<AdminUniverseSummary[]>([]);
  readonly error = signal<string | null>(null);

  readonly newSlug = signal('');
  readonly newName = signal('');
  readonly newDescription = signal('');

  constructor() {
    void this.load();
  }

  private async load(): Promise<void> {
    try {
      this.universes.set(await firstValueFrom(this.content.listUniverses()));
    } catch {
      this.error.set('Failed to load universes.');
    }
  }

  async create(): Promise<void> {
    const slug = this.newSlug().trim();
    const name = this.newName().trim();
    if (!slug || !name) return;
    try {
      await firstValueFrom(
        this.content.createUniverse({ slug, name, description: this.newDescription().trim() }),
      );
      this.newSlug.set('');
      this.newName.set('');
      this.newDescription.set('');
      this.error.set(null);
      await this.load();
    } catch {
      this.error.set('Failed to create universe — slug may already be in use.');
    }
  }

  async save(universe: AdminUniverseSummary): Promise<void> {
    const input: UniverseInput = {
      slug: universe.slug,
      name: universe.name,
      description: universe.description ?? '',
    };
    try {
      await firstValueFrom(this.content.updateUniverse(universe.id, input));
      this.error.set(null);
    } catch {
      this.error.set('Failed to save universe — slug may already be in use.');
    }
  }

  async remove(universe: AdminUniverseSummary): Promise<void> {
    if (!confirm(`Delete "${universe.name}"? Books and characters in it will become unassigned.`)) return;
    await firstValueFrom(this.content.deleteUniverse(universe.id));
    await this.load();
  }

  async exportCharacters(universe: AdminUniverseSummary): Promise<void> {
    try {
      const blob = await firstValueFrom(this.content.exportUniverseCharacters(universe.id));
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${universe.slug}-characters.md`;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      this.error.set('Failed to export characters.');
    }
  }

  async move(index: number, direction: -1 | 1): Promise<void> {
    const universes = this.universes();
    const target = index + direction;
    if (target < 0 || target >= universes.length) return;
    const reordered = [...universes];
    [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
    this.universes.set(reordered);
    await firstValueFrom(this.content.reorderUniverses(reordered.map((u) => u.id)));
  }
}
