import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, catchError, of } from 'rxjs';
import { environment } from '../../environments/environment';
import { BookDetail, BookSummary, Chapter, ChapterSummary } from './content.models';

export interface AdjacentChapters {
  previous: ChapterSummary | null;
  next: ChapterSummary | null;
}

@Injectable({ providedIn: 'root' })
export class ContentService {
  private readonly http = inject(HttpClient);
  private readonly base = environment.apiBaseUrl;

  getBooks(): Observable<BookSummary[]> {
    return this.http.get<BookSummary[]>(`${this.base}/books`);
  }

  getBook(bookSlug: string): Observable<BookDetail | null> {
    return this.http
      .get<BookDetail>(`${this.base}/books/${bookSlug}`)
      .pipe(catchError(() => of(null)));
  }

  getChapter(bookSlug: string, chapterSlug: string): Observable<Chapter | null> {
    return this.http
      .get<Chapter>(`${this.base}/books/${bookSlug}/chapters/${chapterSlug}`)
      .pipe(catchError(() => of(null)));
  }

  getChapterNumber(book: BookDetail | null | undefined, chapterSlug: string): number | undefined {
    const chapters = book?.chapters ?? [];
    const index = chapters.findIndex((chapter) => chapter.slug === chapterSlug);
    return index >= 0 ? index + 1 : undefined;
  }

  formatChapterTitle(title: string, chapterNumber?: number): string {
    const baseTitle = title.replace(/^\s*\d+\s*[:.)-]\s*/, '').trim();
    if (!chapterNumber) {
      return baseTitle;
    }
    return `Chapter ${chapterNumber}: ${baseTitle}`;
  }

  getAdjacentChapters(book: BookDetail | null | undefined, chapterSlug: string): AdjacentChapters {
    const chapters = book?.chapters ?? [];
    const index = chapters.findIndex((chapter) => chapter.slug === chapterSlug);
    return {
      previous: index > 0 ? chapters[index - 1] : null,
      next: index >= 0 && index < chapters.length - 1 ? chapters[index + 1] : null,
    };
  }
}
