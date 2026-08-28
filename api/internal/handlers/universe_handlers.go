package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// exportUniverseCharacters renders every character in a universe as a single
// markdown file (# universe name, then ## character name + notes per
// character) for the author to save externally.
func (h *Handlers) exportUniverseCharacters(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid universe id")
		return
	}

	var universeSlug, universeName string
	err = h.db.QueryRow(`SELECT slug, name FROM universes WHERE id = ?`, id).Scan(&universeSlug, &universeName)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "universe not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load universe")
		return
	}

	rows, err := h.db.Query(`SELECT name, notes FROM characters WHERE universe_id = ? ORDER BY position, id`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load characters")
		return
	}
	defer rows.Close()

	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n", universeName)
	for rows.Next() {
		var name, notes string
		if err := rows.Scan(&name, &notes); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read characters")
			return
		}
		fmt.Fprintf(&sb, "\n## %s\n\n%s\n", name, strings.TrimSpace(notes))
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read characters")
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, universeSlug))
	w.Write([]byte(sb.String()))
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
