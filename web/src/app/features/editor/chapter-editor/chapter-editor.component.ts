import { Component, ElementRef, effect, inject, input, signal, viewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { MusicLink } from '../../../content/content.models';
import { EditorContentService } from '../editor-content.service';
import { imageMarker, parseBlocks, serializeBlocks } from './block-parser';

@Component({
  selector: 'app-chapter-editor',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './chapter-editor.component.html',
  styleUrl: './chapter-editor.component.scss',
})
export class ChapterEditorComponent {
  private readonly content = inject(EditorContentService);
  private readonly router = inject(Router);

  readonly bookId = input.required<string>();
  readonly chapterId = input.required<string>();

  readonly textareaEl = viewChild<ElementRef<HTMLTextAreaElement>>('textareaEl');

  readonly loaded = signal(false);
  readonly slug = signal('');
  readonly title = signal('');
  readonly body = signal('');
  readonly musicLinks = signal<MusicLink[]>([]);

  readonly imageSrc = signal('');
  readonly imageAlt = signal('');
  readonly imageCaption = signal('');

  readonly status = signal<string | null>(null);
  readonly error = signal<string | null>(null);

  constructor() {
    // chapterId is a router-bound input — deferred via effect() for the same
    // reason as ChapterManagerComponent (see its constructor comment).
    effect(() => {
      void this.load(Number(this.chapterId()));
    });
  }

  private async load(chapterId: number): Promise<void> {
    try {
      const chapter = await firstValueFrom(this.content.getChapter(chapterId));
      this.slug.set(chapter.slug);
      this.title.set(chapter.title);
      this.body.set(serializeBlocks(chapter.paragraphs));
      this.musicLinks.set(chapter.musicLinks ?? []);
      this.loaded.set(true);
    } catch {
      this.error.set('Failed to load chapter.');
    }
  }

  insertImage(): void {
    const src = this.imageSrc().trim();
    const alt = this.imageAlt().trim();
    if (!src) return;
    const marker = imageMarker(src, alt, this.imageCaption().trim());

    const el = this.textareaEl()?.nativeElement;
    const current = this.body();
    if (!el) {
      this.body.set(current ? `${current}\n\n${marker}` : marker);
    } else {
      const start = el.selectionStart ?? current.length;
      const end = el.selectionEnd ?? current.length;
      const before = current.slice(0, start);
      const after = current.slice(end);
      const needsLeadingBreak = before.length > 0 && !before.endsWith('\n\n');
      const needsTrailingBreak = after.length > 0 && !after.startsWith('\n\n');
      const insertion = `${needsLeadingBreak ? '\n\n' : ''}${marker}${needsTrailingBreak ? '\n\n' : ''}`;
      this.body.set(before + insertion + after);
    }

    this.imageSrc.set('');
    this.imageAlt.set('');
    this.imageCaption.set('');
  }

  addMusicLink(): void {
    this.musicLinks.set([...this.musicLinks(), { url: '', label: '' }]);
  }

  removeMusicLink(index: number): void {
    this.musicLinks.set(this.musicLinks().filter((_, i) => i !== index));
  }

  async save(): Promise<void> {
    const slug = this.slug().trim();
    const title = this.title().trim();
    if (!slug || !title) {
      this.error.set('Title and slug are required.');
      return;
    }
    try {
      await firstValueFrom(
        this.content.saveChapter(Number(this.chapterId()), {
          slug,
          title,
          paragraphs: parseBlocks(this.body()),
          musicLinks: this.musicLinks().filter((link) => link.url.trim()),
        }),
      );
      this.error.set(null);
      this.status.set('Saved — live for readers now.');
      setTimeout(() => this.status.set(null), 3000);
    } catch {
      this.error.set('Failed to save — slug may already be in use in this book.');
    }
  }

  backToChapters(): void {
    void this.router.navigate(['/editor', this.bookId()]);
  }
}
