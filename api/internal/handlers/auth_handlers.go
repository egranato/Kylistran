package handlers

import (
	"encoding/json"
	"net/http"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var id int64
	var hash string
	var isAdmin bool
	err := h.db.QueryRow(`SELECT id, password_hash, is_admin FROM users WHERE username = ?`, creds.Username).Scan(&id, &hash, &isAdmin)
	if err != nil || !h.auth.CheckPassword(hash, creds.Password) {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, err := h.auth.IssueToken(id, creds.Username, isAdmin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
