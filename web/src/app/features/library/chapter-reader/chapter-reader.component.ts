import { Component, computed, effect, inject, input } from '@angular/core';
import { toObservable, toSignal } from '@angular/core/rxjs-interop';
import { Router, RouterLink } from '@angular/router';
import { combineLatest, switchMap } from 'rxjs';
import { ChapterBlock, ChapterImage, ChapterSummary } from '../../../content/content.models';
import { ContentService } from '../../../content/content.service';

@Component({
  selector: 'app-chapter-reader',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './chapter-reader.component.html',
  styleUrl: './chapter-reader.component.scss',
})
export class ChapterReaderComponent {
  private readonly content = inject(ContentService);
  private readonly router = inject(Router);

  readonly bookSlug = input.required<string>();
  readonly chapterSlug = input.required<string>();

  private readonly data = toSignal(
    combineLatest([toObservable(this.bookSlug), toObservable(this.chapterSlug)]).pipe(
      switchMap(([bookSlug, chapterSlug]) =>
        combineLatest([this.content.getChapter(bookSlug, chapterSlug), this.content.getBook(bookSlug)]),
      ),
    ),
    { initialValue: undefined },
  );

  readonly chapter = computed(() => this.data()?.[0]);
  readonly book = computed(() => this.data()?.[1]);

  readonly chapterNumber = computed(() => this.content.getChapterNumber(this.book(), this.chapterSlug()));
  readonly adjacent = computed(() => this.content.getAdjacentChapters(this.book(), this.chapterSlug()));

  constructor() {
    effect(() => {
      if (this.data() && !this.chapter()) {
        void this.router.navigate(['/library']);
      }
    });
  }

  chapterTitle(title: string, chapterNumber?: number): string {
    return this.content.formatChapterTitle(title, chapterNumber);
  }

  adjacentChapterTitle(chapter: ChapterSummary): string {
    const number = this.content.getChapterNumber(this.book(), chapter.slug);
    return this.content.formatChapterTitle(chapter.title, number);
  }

  isImage(block: ChapterBlock): block is ChapterImage {
    return typeof block === 'object';
  }

  isAuthorComment(block: ChapterBlock): boolean {
    return typeof block === 'string' && block.trimStart().startsWith('#');
  }
}
