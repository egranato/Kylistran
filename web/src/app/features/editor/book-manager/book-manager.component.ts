import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { EditorAuthService } from '../../../core/editor-auth/editor-auth.service';
import { AdminBookSummary, AdminUniverseSummary, BookInput, EditorContentService } from '../editor-content.service';

@Component({
  selector: 'app-book-manager',
  standalone: true,
  imports: [FormsModule, RouterLink],
  templateUrl: './book-manager.component.html',
  styleUrl: './book-manager.component.scss',
})
export class BookManagerComponent {
  private readonly content = inject(EditorContentService);
  private readonly auth = inject(EditorAuthService);
  private readonly router = inject(Router);

  readonly books = signal<AdminBookSummary[]>([]);
  readonly universes = signal<AdminUniverseSummary[]>([]);
  readonly error = signal<string | null>(null);

  readonly newSlug = signal('');
  readonly newTitle = signal('');
  readonly newDescription = signal('');
  readonly newUniverseId = signal<number | null>(null);

  constructor() {
    void this.load();
  }

  private async load(): Promise<void> {
    try {
      const [books, universes] = await Promise.all([
        firstValueFrom(this.content.listBooks()),
        firstValueFrom(this.content.listUniverses()),
      ]);
      this.books.set(books);
      this.universes.set(universes);
    } catch {
      this.error.set('Failed to load books.');
    }
  }

  async create(): Promise<void> {
    const slug = this.newSlug().trim();
    const title = this.newTitle().trim();
    if (!slug || !title) return;
    try {
      await firstValueFrom(
        this.content.createBook({
          slug,
          title,
          description: this.newDescription().trim(),
          hidden: false,
          universeId: this.newUniverseId(),
        }),
      );
      this.newSlug.set('');
      this.newTitle.set('');
      this.newDescription.set('');
      this.newUniverseId.set(null);
      this.error.set(null);
      await this.load();
    } catch {
      this.error.set('Failed to create book — slug may already be in use.');
    }
  }

  async save(book: AdminBookSummary): Promise<void> {
    const input: BookInput = {
      slug: book.slug,
      title: book.title,
      description: book.description ?? '',
      hidden: book.hidden,
      universeId: book.universeId,
    };
    try {
      await firstValueFrom(this.content.updateBook(book.id, input));
      this.error.set(null);
    } catch {
      this.error.set('Failed to save book — slug may already be in use.');
    }
  }

  async remove(book: AdminBookSummary): Promise<void> {
    if (!confirm(`Delete "${book.title}" and all of its chapters? This can't be undone.`)) return;
    await firstValueFrom(this.content.deleteBook(book.id));
    await this.load();
  }

  async move(index: number, direction: -1 | 1): Promise<void> {
    const books = this.books();
    const target = index + direction;
    if (target < 0 || target >= books.length) return;
    const reordered = [...books];
    [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
    this.books.set(reordered);
    await firstValueFrom(this.content.reorderBooks(reordered.map((b) => b.id)));
  }

  logout(): void {
    this.auth.logout();
    void this.router.navigateByUrl('/editor/login');
  }
}
