import { Component, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { ContentService } from '../../../content/content.service';

@Component({
  selector: 'app-book-list',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './book-list.component.html',
})
export class BookListComponent {
  private readonly content = inject(ContentService);
  readonly books = toSignal(this.content.getBooks(), { initialValue: [] });
}
