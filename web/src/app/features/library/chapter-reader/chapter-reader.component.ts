import { Component, computed, effect, inject, input } from '@angular/core';
import { toObservable, toSignal } from '@angular/core/rxjs-interop';
import { Router, RouterLink } from '@angular/router';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { combineLatest, switchMap } from 'rxjs';
import { ChapterBlock, ChapterImage, ChapterSummary } from '../../../content/content.models';
import { ContentService } from '../../../content/content.service';

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

/**
 * Drops author-only content before it reaches the reader: single paragraphs
 * starting with "#", plus everything (including images) between a pair of
 * "###"-only paragraphs. An unclosed trailing "###" hides the rest of the chapter.
 */
function filterReaderBlocks(blocks: ChapterBlock[]): ChapterBlock[] {
  const visible: ChapterBlock[] = [];
  let inCommentBlock = false;

  for (const block of blocks) {
    if (typeof block !== 'string') {
      if (!inCommentBlock) {
        visible.push(block);
      }
      continue;
    }

    const trimmed = block.trim();
    if (trimmed === '###') {
      inCommentBlock = !inCommentBlock;
      continue;
    }
    if (inCommentBlock || trimmed.startsWith('#')) {
      continue;
    }
    visible.push(block);
  }

  return visible;
}

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
  private readonly sanitizer = inject(DomSanitizer);

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
  readonly visibleParagraphs = computed(() => filterReaderBlocks(this.chapter()?.paragraphs ?? []));

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

  formatParagraph(text: string): SafeHtml {
    const withItalics = escapeHtml(text).replace(/\*(\S(?:[^*]*\S)?)\*/g, '<i>$1</i>');
    return this.sanitizer.bypassSecurityTrustHtml(withItalics);
  }
}
