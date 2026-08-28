import { Component, computed, effect, inject, input, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { marked } from 'marked';
import { AdminUniverseSummary, EditorContentService } from '../editor-content.service';

@Component({
  selector: 'app-character-editor',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './character-editor.component.html',
  styleUrl: './character-editor.component.scss',
})
export class CharacterEditorComponent {
  private readonly content = inject(EditorContentService);
  private readonly router = inject(Router);

  readonly characterId = input.required<string>();

  readonly loaded = signal(false);
  readonly name = signal('');
  readonly notes = signal('');
  readonly universeId = signal<number | null>(null);
  readonly universes = signal<AdminUniverseSummary[]>([]);

  readonly status = signal<string | null>(null);
  readonly error = signal<string | null>(null);

  readonly previewing = signal(false);
  readonly notesHtml = computed(() => marked.parse(this.notes(), { async: false }) as string);

  constructor() {
    // characterId is a router-bound input — deferred via effect() for the same
    // reason as ChapterEditorComponent (see its constructor comment).
    effect(() => {
      void this.load(Number(this.characterId()));
    });
  }

  private async load(characterId: number): Promise<void> {
    try {
      const [character, universes] = await Promise.all([
        firstValueFrom(this.content.getCharacter(characterId)),
        firstValueFrom(this.content.listUniverses()),
      ]);
      this.name.set(character.name);
      this.notes.set(character.notes);
      this.universeId.set(character.universeId ?? null);
      this.universes.set(universes);
      this.loaded.set(true);
    } catch {
      this.error.set('Failed to load character.');
    }
  }

  async save(): Promise<void> {
    const name = this.name().trim();
    if (!name) {
      this.error.set('Name is required.');
      return;
    }
    try {
      await firstValueFrom(
        this.content.saveCharacter(Number(this.characterId()), {
          name,
          notes: this.notes(),
          universeId: this.universeId(),
        }),
      );
      this.error.set(null);
      this.status.set('Saved.');
      setTimeout(() => this.status.set(null), 3000);
    } catch {
      this.error.set('Failed to save character.');
    }
  }

  togglePreview(): void {
    this.previewing.update((v) => !v);
  }

  backToCharacters(): void {
    void this.router.navigate(['/editor/characters']);
  }
}
