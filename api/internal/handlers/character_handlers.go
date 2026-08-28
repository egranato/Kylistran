package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"kylistran-api/internal/models"
)

func (h *Handlers) listAdminCharacters(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(
		`SELECT c.id, c.name, c.position, c.universe_id, u.slug, u.name
		 FROM characters c LEFT JOIN universes u ON u.id = c.universe_id
		 ORDER BY c.position, c.id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list characters")
		return
	}
	defer rows.Close()

	characters := []models.AdminCharacterSummary{}
	for rows.Next() {
		var c models.AdminCharacterSummary
		var universeID sql.NullInt64
		var universeSlug, universeName sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Position, &universeID, &universeSlug, &universeName); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read characters")
			return
		}
		if universeID.Valid {
			id := universeID.Int64
			c.UniverseID = &id
			c.Universe = &models.UniverseRef{Slug: universeSlug.String, Name: universeName.String}
		}
		characters = append(characters, c)
	}
	writeJSON(w, http.StatusOK, characters)
}

type characterCreateInput struct {
	Name       string `json:"name"`
	UniverseID *int64 `json:"universeId"`
}

func (h *Handlers) createCharacter(w http.ResponseWriter, r *http.Request) {
	var in characterCreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var position int
	h.db.QueryRow(`SELECT COALESCE(MAX(position), -1) + 1 FROM characters`).Scan(&position)

	res, err := h.db.Exec(`INSERT INTO characters (universe_id, name, position) VALUES (?, ?, ?)`,
		in.UniverseID, in.Name, position)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create character")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *Handlers) getAdminCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	var character models.AdminCharacter
	character.ID = id
	var universeID sql.NullInt64
	err = h.db.QueryRow(`SELECT name, notes, universe_id FROM characters WHERE id = ?`, id).
		Scan(&character.Name, &character.Notes, &universeID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "character not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load character")
		return
	}
	if universeID.Valid {
		character.UniverseID = &universeID.Int64
	}
	writeJSON(w, http.StatusOK, character)
}

type characterSaveInput struct {
	Name       string `json:"name"`
	Notes      string `json:"notes"`
	UniverseID *int64 `json:"universeId"`
}

func (h *Handlers) putCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	var in characterSaveInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	res, err := h.db.Exec(
		`UPDATE characters SET name = ?, notes = ?, universe_id = ?, updated_at = datetime('now') WHERE id = ?`,
		in.Name, in.Notes, in.UniverseID, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save character")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "character not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) deleteCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	res, err := h.db.Exec(`DELETE FROM characters WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete character")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "character not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) reorderCharacters(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Order []int64 `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder characters")
		return
	}
	defer tx.Rollback()

	for i, id := range in.Order {
		if _, err := tx.Exec(`UPDATE characters SET position = ? WHERE id = ?`, i, id); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reorder characters")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder characters")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
