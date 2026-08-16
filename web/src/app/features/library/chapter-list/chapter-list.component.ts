import { Component, effect, inject, input } from '@angular/core';
import { toObservable, toSignal } from '@angular/core/rxjs-interop';
import { Router, RouterLink } from '@angular/router';
import { switchMap } from 'rxjs';
import { ContentService } from '../../../content/content.service';

@Component({
  selector: 'app-chapter-list',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './chapter-list.component.html',
})
export class ChapterListComponent {
  private readonly content = inject(ContentService);
  private readonly router = inject(Router);

  readonly bookSlug = input.required<string>();

  readonly book = toSignal(
    toObservable(this.bookSlug).pipe(switchMap((slug) => this.content.getBook(slug))),
    { initialValue: undefined },
  );

  constructor() {
    effect(() => {
      if (this.book() === null) {
        void this.router.navigate(['/library']);
      }
    });
  }

  chapterTitle(title: string, chapterNumber: number): string {
    return this.content.formatChapterTitle(title, chapterNumber);
  }
}
