package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"kylistran-api/internal/models"
)

func loadBlocks(db *sql.DB, chapterID int64) ([]models.Block, error) {
	rows, err := db.Query(
		`SELECT block_type, text, image_src, image_alt, image_caption
		 FROM chapter_blocks WHERE chapter_id = ? ORDER BY position, id`, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocks := []models.Block{}
	for rows.Next() {
		var blockType, text, src, alt, caption string
		if err := rows.Scan(&blockType, &text, &src, &alt, &caption); err != nil {
			return nil, err
		}
		if blockType == "image" {
			blocks = append(blocks, models.Block{Image: &models.ImageBlock{Type: "image", Src: src, Alt: alt, Caption: caption}})
		} else {
			blocks = append(blocks, models.Block{Paragraph: text})
		}
	}
	return blocks, rows.Err()
}

func loadMusicLinks(db *sql.DB, chapterID int64) ([]models.MusicLink, error) {
	rows, err := db.Query(`SELECT url, label FROM chapter_music_links WHERE chapter_id = ? ORDER BY position, id`, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []models.MusicLink{}
	for rows.Next() {
		var l models.MusicLink
		if err := rows.Scan(&l.URL, &l.Label); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (h *Handlers) getChapterPublic(w http.ResponseWriter, r *http.Request) {
	bookSlug := r.PathValue("bookSlug")
	chapterSlug := r.PathValue("chapterSlug")

	var chapterID int64
	var chapter models.Chapter
	err := h.db.QueryRow(
		`SELECT c.id, c.slug, c.title FROM chapters c
		 JOIN books b ON b.id = c.book_id
		 WHERE b.slug = ? AND c.slug = ?`, bookSlug, chapterSlug,
	).Scan(&chapterID, &chapter.Slug, &chapter.Title)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapter")
		return
	}

	blocks, err := loadBlocks(h.db, chapterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapter content")
		return
	}
	links, err := loadMusicLinks(h.db, chapterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapter content")
		return
	}
	chapter.Paragraphs = blocks
	chapter.MusicLinks = links

	writeJSON(w, http.StatusOK, chapter)
}

func (h *Handlers) listAdminChapters(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	rows, err := h.db.Query(`SELECT id, slug, title, position FROM chapters WHERE book_id = ? ORDER BY position, id`, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list chapters")
		return
	}
	defer rows.Close()

	chapters := []models.AdminChapterSummary{}
	for rows.Next() {
		var c models.AdminChapterSummary
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Position); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read chapters")
			return
		}
		chapters = append(chapters, c)
	}
	writeJSON(w, http.StatusOK, chapters)
}

type chapterCreateInput struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

func (h *Handlers) createChapter(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var in chapterCreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Slug == "" || in.Title == "" {
		writeError(w, http.StatusBadRequest, "slug and title are required")
		return
	}

	var position int
	h.db.QueryRow(`SELECT COALESCE(MAX(position), -1) + 1 FROM chapters WHERE book_id = ?`, bookID).Scan(&position)

	res, err := h.db.Exec(`INSERT INTO chapters (book_id, slug, title, position) VALUES (?, ?, ?, ?)`,
		bookID, in.Slug, in.Title, position)
	if err != nil {
		writeError(w, http.StatusConflict, "a chapter with that slug already exists in this book")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *Handlers) reorderChapters(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var in struct {
		Order []int64 `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder chapters")
		return
	}
	defer tx.Rollback()

	for i, id := range in.Order {
		if _, err := tx.Exec(`UPDATE chapters SET position = ? WHERE id = ? AND book_id = ?`, i, id, bookID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reorder chapters")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder chapters")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) getAdminChapter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	var chapter models.AdminChapter
	chapter.ID = id
	err = h.db.QueryRow(`SELECT slug, title FROM chapters WHERE id = ?`, id).Scan(&chapter.Slug, &chapter.Title)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapter")
		return
	}

	blocks, err := loadBlocks(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapter content")
		return
	}
	links, err := loadMusicLinks(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapter content")
		return
	}
	chapter.Paragraphs = blocks
	chapter.MusicLinks = links

	writeJSON(w, http.StatusOK, chapter)
}

type chapterPutInput struct {
	Slug       string             `json:"slug"`
	Title      string             `json:"title"`
	Paragraphs []models.Block     `json:"paragraphs"`
	MusicLinks []models.MusicLink `json:"musicLinks"`
}

func (h *Handlers) putChapter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	var in chapterPutInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Slug == "" || in.Title == "" {
		writeError(w, http.StatusBadRequest, "slug and title are required")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save chapter")
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE chapters SET slug = ?, title = ?, updated_at = datetime('now') WHERE id = ?`, in.Slug, in.Title, id)
	if err != nil {
		writeError(w, http.StatusConflict, "a chapter with that slug already exists in this book")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}

	if _, err := tx.Exec(`DELETE FROM chapter_blocks WHERE chapter_id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save chapter")
		return
	}
	for i, b := range in.Paragraphs {
		if b.Image != nil {
			_, err = tx.Exec(
				`INSERT INTO chapter_blocks (chapter_id, position, block_type, image_src, image_alt, image_caption) VALUES (?, ?, 'image', ?, ?, ?)`,
				id, i, b.Image.Src, b.Image.Alt, b.Image.Caption,
			)
		} else {
			_, err = tx.Exec(
				`INSERT INTO chapter_blocks (chapter_id, position, block_type, text) VALUES (?, ?, 'paragraph', ?)`,
				id, i, b.Paragraph,
			)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save chapter")
			return
		}
	}

	if _, err := tx.Exec(`DELETE FROM chapter_music_links WHERE chapter_id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save chapter")
		return
	}
	for i, l := range in.MusicLinks {
		if _, err := tx.Exec(`INSERT INTO chapter_music_links (chapter_id, position, url, label) VALUES (?, ?, ?, ?)`, id, i, l.URL, l.Label); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save chapter")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save chapter")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) deleteChapter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	res, err := h.db.Exec(`DELETE FROM chapters WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete chapter")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
