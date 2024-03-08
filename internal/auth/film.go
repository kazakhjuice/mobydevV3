package auth

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"mobydevLogin/internal/helpers"
	"net/http"
)

func (h *dbHandler) UpdateFilm(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger) {

	var film FilmDetails

	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		helpers.ServeError(err, w, "Invalid JSON format", log, http.StatusBadRequest)
		return
	}

	tagsJSON, err := json.Marshal(film.Tags)
	if err != nil {
		helpers.ServeError(err, w, "internal error", log, http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		INSERT INTO films (name, category, project_type, year, duration, tags, description, director, producer)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, film.Name, film.Category, film.ProjectType, film.Year, film.Duration, string(tagsJSON), film.Description, film.Director, film.Producer)
	if err != nil {
		helpers.ServeError(err, w, "internal error", log, http.StatusInternalServerError)
		return
	}

	log.Info("Film details inserted successfully.")
	w.Write([]byte("film updated"))
}

func (h *dbHandler) ViewFilm(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger) {
	id, err := helpers.RetrieveID(r)

	if err != nil {
		helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
		SELECT name, category, project_type, year, duration, tags, description, director, producer
		FROM films
		WHERE id = ?
	`, id)

	if err != nil {
		helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var film FilmDetails

		var tags string

		err := rows.Scan(
			&film.Name,
			&film.Category,
			&film.ProjectType,
			&film.Year,
			&film.Duration,
			&tags,
			&film.Description,
			&film.Director,
			&film.Producer,
		)
		if err != nil {
			helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal([]byte(tags), &film.Tags)
		if err != nil {
			helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
			return
		}

		responseJSON, err := json.Marshal(film)
		if err != nil {
			helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	} else {
		http.Error(w, "Film not found", http.StatusNotFound)
	}
}
