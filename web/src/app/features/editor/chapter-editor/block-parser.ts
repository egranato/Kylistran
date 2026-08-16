import { ChapterBlock } from '../../../content/content.models';

const IMAGE_MARKER = /^\{\{image\s+src="([^"]*)"\s+alt="([^"]*)"(?:\s+caption="([^"]*)")?\s*\}\}$/;

/**
 * Splits raw textarea input into ChapterBlocks: one non-blank line per block,
 * ported from the old tools/chapter-formatter.html workflow. A line matching
 * the {{image ...}} marker becomes a ChapterImage instead of a paragraph.
 */
export function parseBlocks(raw: string): ChapterBlock[] {
  const lines = raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.length > 0);

  return lines.map((line): ChapterBlock => {
    const match = IMAGE_MARKER.exec(line);
    if (!match) {
      return line;
    }
    const [, src, alt, caption] = match;
    return caption ? { type: 'image', src, alt, caption } : { type: 'image', src, alt };
  });
}

/** Inverse of parseBlocks, for loading a chapter back into the textarea. */
export function serializeBlocks(blocks: ChapterBlock[]): string {
  return blocks
    .map((block) =>
      typeof block === 'string'
        ? block
        : `{{image src="${block.src}" alt="${block.alt}"${block.caption ? ` caption="${block.caption}"` : ''}}}`,
    )
    .join('\n\n');
}

export function imageMarker(src: string, alt: string, caption: string): string {
  return `{{image src="${src}" alt="${alt}"${caption ? ` caption="${caption}"` : ''}}}`;
}
