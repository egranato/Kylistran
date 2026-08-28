package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kylistran-api/internal/models"
)

func (h *Handlers) listAdminUniverses(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, slug, name, description FROM universes ORDER BY position, id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list universes")
		return
	}
	defer rows.Close()

	universes := []models.AdminUniverseSummary{}
	for rows.Next() {
		var u models.AdminUniverseSummary
		if err := rows.Scan(&u.ID, &u.Slug, &u.Name, &u.Description); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read universes")
			return
		}
		universes = append(universes, u)
	}
	writeJSON(w, http.StatusOK, universes)
}

type universeInput struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handlers) createUniverse(w http.ResponseWriter, r *http.Request) {
	var in universeInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Slug == "" || in.Name == "" {
		writeError(w, http.StatusBadRequest, "slug and name are required")
		return
	}

	var position int
	h.db.QueryRow(`SELECT COALESCE(MAX(position), -1) + 1 FROM universes`).Scan(&position)

	res, err := h.db.Exec(`INSERT INTO universes (slug, name, description, position) VALUES (?, ?, ?, ?)`,
		in.Slug, in.Name, in.Description, position)
	if err != nil {
		writeError(w, http.StatusConflict, "a universe with that slug already exists")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *Handlers) updateUniverse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid universe id")
		return
	}

	var in universeInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Slug == "" || in.Name == "" {
		writeError(w, http.StatusBadRequest, "slug and name are required")
		return
	}

	res, err := h.db.Exec(
		`UPDATE universes SET slug = ?, name = ?, description = ?, updated_at = datetime('now') WHERE id = ?`,
		in.Slug, in.Name, in.Description, id,
	)
	if err != nil {
		writeError(w, http.StatusConflict, "a universe with that slug already exists")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "universe not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) deleteUniverse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid universe id")
		return
	}

	res, err := h.db.Exec(`DELETE FROM universes WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete universe")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "universe not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) reorderUniverses(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Order []int64 `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder universes")
		return
	}
	defer tx.Rollback()

	for i, id := range in.Order {
		if _, err := tx.Exec(`UPDATE universes SET position = ? WHERE id = ?`, i, id); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reorder universes")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder universes")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
