export interface MusicLink {
  url: string;
  label?: string;
}

export interface ChapterImage {
  type: 'image';
  src: string;
  alt: string;
  caption?: string;
}

export type ChapterBlock = string | ChapterImage;

export interface Chapter {
  slug: string;
  title: string;
  paragraphs: ChapterBlock[];
  musicLinks?: MusicLink[];
}

export interface UniverseRef {
  slug: string;
  name: string;
}

export interface BookSummary {
  slug: string;
  title: string;
  description?: string;
  universe?: UniverseRef;
}

export interface ChapterSummary {
  slug: string;
  title: string;
}

export interface BookDetail extends BookSummary {
  chapters: ChapterSummary[];
}
