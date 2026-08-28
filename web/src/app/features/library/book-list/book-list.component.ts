import { Component, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { ContentService } from '../../../content/content.service';
import { BookSummary } from '../../../content/content.models';

interface BookSection {
  universeName: string | null;
  books: BookSummary[];
}

@Component({
  selector: 'app-book-list',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './book-list.component.html',
})
export class BookListComponent {
  private readonly content = inject(ContentService);
  readonly books = toSignal(this.content.getBooks(), { initialValue: [] });

  // Groups adjacent books that share a universe under one header; a run of
  // length 1 is flattened back to a headerless entry so a lone book in a
  // universe (or one with none) renders like any standalone book.
  readonly sections = computed<BookSection[]>(() => {
    const runs: BookSection[] = [];
    for (const book of this.books()) {
      const last = runs[runs.length - 1];
      if (book.universe && last?.universeName === book.universe.name) {
        last.books.push(book);
      } else {
        runs.push({ universeName: book.universe?.name ?? null, books: [book] });
      }
    }
    return runs.map((run) => (run.books.length === 1 ? { ...run, universeName: null } : run));
  });
}
