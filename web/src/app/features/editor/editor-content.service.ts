import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { ChapterBlock, MusicLink, UniverseRef } from '../../content/content.models';

export interface AdminBookSummary {
  id: number;
  slug: string;
  title: string;
  description?: string;
  hidden: boolean;
  universeId?: number | null;
  universe?: UniverseRef;
}

export interface AdminUniverseSummary {
  id: number;
  slug: string;
  name: string;
  description?: string;
}

export interface UniverseInput {
  slug: string;
  name: string;
  description: string;
}

export interface AdminCharacterSummary {
  id: number;
  name: string;
  position: number;
  universeId?: number | null;
  universe?: UniverseRef;
}

export interface AdminCharacter {
  id: number;
  name: string;
  notes: string;
  universeId?: number | null;
}

export interface CharacterCreateInput {
  name: string;
  universeId?: number | null;
}

export interface CharacterSaveInput {
  name: string;
  notes: string;
  universeId?: number | null;
}

export interface AdminChapterSummary {
  id: number;
  slug: string;
  title: string;
  position: number;
}

export interface AdminChapter {
  id: number;
  slug: string;
  title: string;
  paragraphs: ChapterBlock[];
  musicLinks?: MusicLink[];
}

export interface BookInput {
  slug: string;
  title: string;
  description: string;
  hidden: boolean;
  universeId?: number | null;
}

export interface ChapterCreateInput {
  slug: string;
  title: string;
}

export interface ChapterSaveInput {
  slug: string;
  title: string;
  paragraphs: ChapterBlock[];
  musicLinks: MusicLink[];
}

@Injectable({ providedIn: 'root' })
export class EditorContentService {
  private readonly http = inject(HttpClient);
  private readonly base = environment.apiBaseUrl;

  listBooks(): Observable<AdminBookSummary[]> {
    return this.http.get<AdminBookSummary[]>(`${this.base}/admin/books`);
  }

  createBook(input: BookInput): Observable<{ id: number }> {
    return this.http.post<{ id: number }>(`${this.base}/admin/books`, input);
  }

  updateBook(id: number, input: BookInput): Observable<void> {
    return this.http.patch<void>(`${this.base}/admin/books/${id}`, input);
  }

  deleteBook(id: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/admin/books/${id}`);
  }

  reorderBooks(order: number[]): Observable<void> {
    return this.http.patch<void>(`${this.base}/admin/books/reorder`, { order });
  }

  listChapters(bookId: number): Observable<AdminChapterSummary[]> {
    return this.http.get<AdminChapterSummary[]>(`${this.base}/admin/books/${bookId}/chapters`);
  }

  createChapter(bookId: number, input: ChapterCreateInput): Observable<{ id: number }> {
    return this.http.post<{ id: number }>(`${this.base}/admin/books/${bookId}/chapters`, input);
  }

  reorderChapters(bookId: number, order: number[]): Observable<void> {
    return this.http.patch<void>(`${this.base}/admin/books/${bookId}/chapters/reorder`, { order });
  }

  getChapter(id: number): Observable<AdminChapter> {
    return this.http.get<AdminChapter>(`${this.base}/admin/chapters/${id}`);
  }

  saveChapter(id: number, input: ChapterSaveInput): Observable<void> {
    return this.http.put<void>(`${this.base}/admin/chapters/${id}`, input);
  }

  deleteChapter(id: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/admin/chapters/${id}`);
  }

  listUniverses(): Observable<AdminUniverseSummary[]> {
    return this.http.get<AdminUniverseSummary[]>(`${this.base}/admin/universes`);
  }

  createUniverse(input: UniverseInput): Observable<{ id: number }> {
    return this.http.post<{ id: number }>(`${this.base}/admin/universes`, input);
  }

  updateUniverse(id: number, input: UniverseInput): Observable<void> {
    return this.http.patch<void>(`${this.base}/admin/universes/${id}`, input);
  }

  deleteUniverse(id: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/admin/universes/${id}`);
  }

  reorderUniverses(order: number[]): Observable<void> {
    return this.http.patch<void>(`${this.base}/admin/universes/reorder`, { order });
  }

  exportUniverseCharacters(id: number): Observable<Blob> {
    return this.http.get(`${this.base}/admin/universes/${id}/characters/export`, { responseType: 'blob' });
  }

  listCharacters(): Observable<AdminCharacterSummary[]> {
    return this.http.get<AdminCharacterSummary[]>(`${this.base}/admin/characters`);
  }

  createCharacter(input: CharacterCreateInput): Observable<{ id: number }> {
    return this.http.post<{ id: number }>(`${this.base}/admin/characters`, input);
  }

  reorderCharacters(order: number[]): Observable<void> {
    return this.http.patch<void>(`${this.base}/admin/characters/reorder`, { order });
  }

  getCharacter(id: number): Observable<AdminCharacter> {
    return this.http.get<AdminCharacter>(`${this.base}/admin/characters/${id}`);
  }

  saveCharacter(id: number, input: CharacterSaveInput): Observable<void> {
    return this.http.put<void>(`${this.base}/admin/characters/${id}`, input);
  }

  deleteCharacter(id: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/admin/characters/${id}`);
  }
}
