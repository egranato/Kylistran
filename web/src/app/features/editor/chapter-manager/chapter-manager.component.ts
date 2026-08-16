import { Component, effect, inject, input, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { AdminChapterSummary, EditorContentService } from '../editor-content.service';

@Component({
  selector: 'app-chapter-manager',
  standalone: true,
  imports: [FormsModule, RouterLink],
  templateUrl: './chapter-manager.component.html',
  styleUrl: './chapter-manager.component.scss',
})
export class ChapterManagerComponent {
  private readonly content = inject(EditorContentService);

  readonly bookId = input.required<string>();

  readonly bookTitle = signal('');
  readonly chapters = signal<AdminChapterSummary[]>([]);
  readonly error = signal<string | null>(null);

  readonly newSlug = signal('');
  readonly newTitle = signal('');

  constructor() {
    // bookId is a router-bound input — it isn't guaranteed to be set yet if
    // read synchronously in the constructor, so defer via effect() until
    // Angular has applied it (and re-run if it later changes).
    effect(() => {
      void this.load(Number(this.bookId()));
    });
  }

  private async load(bookId: number): Promise<void> {
    try {
      const [books, chapters] = await Promise.all([
        firstValueFrom(this.content.listBooks()),
        firstValueFrom(this.content.listChapters(bookId)),
      ]);
      this.bookTitle.set(books.find((b) => b.id === bookId)?.title ?? '');
      this.chapters.set(chapters);
    } catch {
      this.error.set('Failed to load chapters.');
    }
  }

  async create(): Promise<void> {
    const slug = this.newSlug().trim();
    const title = this.newTitle().trim();
    if (!slug || !title) return;
    try {
      await firstValueFrom(this.content.createChapter(Number(this.bookId()), { slug, title }));
      this.newSlug.set('');
      this.newTitle.set('');
      this.error.set(null);
      await this.load(Number(this.bookId()));
    } catch {
      this.error.set('Failed to create chapter — slug may already be in use in this book.');
    }
  }

  async remove(chapter: AdminChapterSummary): Promise<void> {
    if (!confirm(`Delete "${chapter.title}"? This can't be undone.`)) return;
    await firstValueFrom(this.content.deleteChapter(chapter.id));
    await this.load(Number(this.bookId()));
  }

  async move(index: number, direction: -1 | 1): Promise<void> {
    const chapters = this.chapters();
    const target = index + direction;
    if (target < 0 || target >= chapters.length) return;
    const reordered = [...chapters];
    [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
    this.chapters.set(reordered);
    await firstValueFrom(this.content.reorderChapters(Number(this.bookId()), reordered.map((c) => c.id)));
  }
}
