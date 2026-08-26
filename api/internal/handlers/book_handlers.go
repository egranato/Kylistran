package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"kylistran-api/internal/models"
)

func (h *Handlers) listBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT slug, title, description FROM books WHERE hidden = 0 ORDER BY position, id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list books")
		return
	}
	defer rows.Close()

	books := []models.BookSummary{}
	for rows.Next() {
		var b models.BookSummary
		if err := rows.Scan(&b.Slug, &b.Title, &b.Description); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read books")
			return
		}
		books = append(books, b)
	}
	writeJSON(w, http.StatusOK, books)
}

func (h *Handlers) getBookDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("bookSlug")

	var bookID int64
	var detail models.BookDetail
	err := h.db.QueryRow(`SELECT id, slug, title, description FROM books WHERE slug = ?`, slug).
		Scan(&bookID, &detail.Slug, &detail.Title, &detail.Description)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "book not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load book")
		return
	}

	rows, err := h.db.Query(`SELECT slug, title FROM chapters WHERE book_id = ? ORDER BY position, id`, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load chapters")
		return
	}
	defer rows.Close()

	detail.Chapters = []models.ChapterSummary{}
	for rows.Next() {
		var c models.ChapterSummary
		if err := rows.Scan(&c.Slug, &c.Title); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read chapters")
			return
		}
		detail.Chapters = append(detail.Chapters, c)
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handlers) listAdminBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, slug, title, description, hidden FROM books ORDER BY position, id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list books")
		return
	}
	defer rows.Close()

	books := []models.AdminBookSummary{}
	for rows.Next() {
		var b models.AdminBookSummary
		if err := rows.Scan(&b.ID, &b.Slug, &b.Title, &b.Description, &b.Hidden); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read books")
			return
		}
		books = append(books, b)
	}
	writeJSON(w, http.StatusOK, books)
}

type bookInput struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Hidden      bool   `json:"hidden"`
}

func (h *Handlers) createBook(w http.ResponseWriter, r *http.Request) {
	var in bookInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Slug == "" || in.Title == "" {
		writeError(w, http.StatusBadRequest, "slug and title are required")
		return
	}

	var position int
	h.db.QueryRow(`SELECT COALESCE(MAX(position), -1) + 1 FROM books`).Scan(&position)

	res, err := h.db.Exec(`INSERT INTO books (slug, title, description, position, hidden) VALUES (?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Description, position, in.Hidden)
	if err != nil {
		writeError(w, http.StatusConflict, "a book with that slug already exists")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *Handlers) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var in bookInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Slug == "" || in.Title == "" {
		writeError(w, http.StatusBadRequest, "slug and title are required")
		return
	}

	res, err := h.db.Exec(
		`UPDATE books SET slug = ?, title = ?, description = ?, hidden = ?, updated_at = datetime('now') WHERE id = ?`,
		in.Slug, in.Title, in.Description, in.Hidden, id,
	)
	if err != nil {
		writeError(w, http.StatusConflict, "a book with that slug already exists")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	res, err := h.db.Exec(`DELETE FROM books WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete book")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) reorderBooks(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Order []int64 `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder books")
		return
	}
	defer tx.Rollback()

	for i, id := range in.Order {
		if _, err := tx.Exec(`UPDATE books SET position = ? WHERE id = ?`, i, id); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reorder books")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder books")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
