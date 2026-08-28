package models

import (
	"bytes"
	"encoding/json"
)

// Block mirrors the frontend's ChapterBlock union (string | ChapterImage) by
// giving itself custom JSON (de)serialization, so the wire format needs no
// translation layer on the Angular side.
type Block struct {
	Paragraph string      // set when Image is nil
	Image     *ImageBlock // set when this is an image block
}

type ImageBlock struct {
	Type    string `json:"type"`
	Src     string `json:"src"`
	Alt     string `json:"alt"`
	Caption string `json:"caption,omitempty"`
}

func (b Block) MarshalJSON() ([]byte, error) {
	if b.Image != nil {
		return json.Marshal(b.Image)
	}
	return json.Marshal(b.Paragraph)
}

func (b *Block) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		return json.Unmarshal(data, &b.Paragraph)
	}
	var img ImageBlock
	if err := json.Unmarshal(data, &img); err != nil {
		return err
	}
	img.Type = "image"
	b.Image = &img
	return nil
}

type MusicLink struct {
	URL   string `json:"url"`
	Label string `json:"label,omitempty"`
}

// UniverseRef is the read-only display shape embedded wherever a universe
// needs to be shown without exposing its numeric id (public book summaries,
// admin character/book rows).
type UniverseRef struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type BookSummary struct {
	Slug        string       `json:"slug"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Universe    *UniverseRef `json:"universe,omitempty"`
}

type Universe struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ChapterSummary struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type BookDetail struct {
	BookSummary
	Chapters []ChapterSummary `json:"chapters"`
}

type Chapter struct {
	Slug       string      `json:"slug"`
	Title      string      `json:"title"`
	Paragraphs []Block     `json:"paragraphs"`
	MusicLinks []MusicLink `json:"musicLinks,omitempty"`
}

// AdminBookSummary/AdminChapterSummary add the numeric ids the editor needs
// for admin CRUD/reorder endpoints, without leaking ids into public reads.
type AdminBookSummary struct {
	ID         int64  `json:"id"`
	Hidden     bool   `json:"hidden"`
	UniverseID *int64 `json:"universeId,omitempty"`
	BookSummary
}

type AdminUniverseSummary struct {
	ID int64 `json:"id"`
	Universe
}

type AdminCharacterSummary struct {
	ID         int64        `json:"id"`
	Name       string       `json:"name"`
	Position   int          `json:"position"`
	UniverseID *int64       `json:"universeId,omitempty"`
	Universe   *UniverseRef `json:"universe,omitempty"`
}

type AdminCharacter struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Notes      string `json:"notes"`
	UniverseID *int64 `json:"universeId,omitempty"`
}

type AdminChapterSummary struct {
	ID       int64  `json:"id"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Position int    `json:"position"`
}

type AdminChapter struct {
	ID int64 `json:"id"`
	Chapter
}
