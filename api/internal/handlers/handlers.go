package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"kylistran-api/internal/auth"
)

type Handlers struct {
	db   *sql.DB
	auth *auth.Service
}

func New(db *sql.DB, authSvc *auth.Service) *Handlers {
	return &Handlers{db: db, auth: authSvc}
}

func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /auth/login", h.login)

	// Public reads
	mux.HandleFunc("GET /books", h.listBooks)
	mux.HandleFunc("GET /books/{bookSlug}", h.getBookDetail)
	mux.HandleFunc("GET /books/{bookSlug}/chapters/{chapterSlug}", h.getChapterPublic)

	// Admin: books
	admin := func(fn http.HandlerFunc) http.Handler {
		return h.auth.RequireAuth(auth.RequireAdmin(fn))
	}
	mux.Handle("GET /admin/books", admin(h.listAdminBooks))
	mux.Handle("POST /admin/books", admin(h.createBook))
	mux.Handle("PATCH /admin/books/reorder", admin(h.reorderBooks))
	mux.Handle("PATCH /admin/books/{id}", admin(h.updateBook))
	mux.Handle("DELETE /admin/books/{id}", admin(h.deleteBook))

	// Admin: chapters
	mux.Handle("GET /admin/books/{id}/chapters", admin(h.listAdminChapters))
	mux.Handle("POST /admin/books/{id}/chapters", admin(h.createChapter))
	mux.Handle("PATCH /admin/books/{id}/chapters/reorder", admin(h.reorderChapters))
	mux.Handle("GET /admin/chapters/{id}", admin(h.getAdminChapter))
	mux.Handle("PUT /admin/chapters/{id}", admin(h.putChapter))
	mux.Handle("DELETE /admin/chapters/{id}", admin(h.deleteChapter))

	// Admin: universes
	mux.Handle("GET /admin/universes", admin(h.listAdminUniverses))
	mux.Handle("POST /admin/universes", admin(h.createUniverse))
	mux.Handle("PATCH /admin/universes/reorder", admin(h.reorderUniverses))
	mux.Handle("PATCH /admin/universes/{id}", admin(h.updateUniverse))
	mux.Handle("DELETE /admin/universes/{id}", admin(h.deleteUniverse))
	mux.Handle("GET /admin/universes/{id}/characters/export", admin(h.exportUniverseCharacters))

	// Admin: characters
	mux.Handle("GET /admin/characters", admin(h.listAdminCharacters))
	mux.Handle("POST /admin/characters", admin(h.createCharacter))
	mux.Handle("PATCH /admin/characters/reorder", admin(h.reorderCharacters))
	mux.Handle("GET /admin/characters/{id}", admin(h.getAdminCharacter))
	mux.Handle("PUT /admin/characters/{id}", admin(h.putCharacter))
	mux.Handle("DELETE /admin/characters/{id}", admin(h.deleteCharacter))
}

func (h *Handlers) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
