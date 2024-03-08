package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"mobydevLogin/internal/helpers"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type DBHandler interface {
	Login(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger)
	Register(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger)
	UpdateFilm(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger)
	ViewFilm(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger)
}

type dbHandler struct{}

var jwtSecret = []byte("your-secret-key") // TODO: keep in env

func (h *dbHandler) Login(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger) {

	var request LoginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		helpers.ServeError(err, w, "Invalid JSON format", log, http.StatusBadRequest)
		return
	}

	email := request.Email
	password := request.Password

	var storedPassword string
	var id int
	var isAdmin bool

	if !helpers.IsValidEmail(email) {
		helpers.ServeError(errors.New("wrong email format"), w, "wrong email format", log, http.StatusBadRequest)
		return
	}

	err = db.QueryRow("SELECT password FROM users WHERE email=?", email).Scan(&storedPassword)
	if err == sql.ErrNoRows {
		helpers.ServeError(err, w, "user doesn't exist", log, http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		helpers.ServeError(err, w, "wrong credentials", log, http.StatusBadRequest)
		return
	}

	err = db.QueryRow("SELECT isAdmin, id FROM users WHERE email=?", email).Scan(&isAdmin, &id)

	if err != nil {
		helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
		return
	}

	token, err := generateJWT(isAdmin, id)
	if err != nil {
		helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})
	log.Info("User logged:", slog.String("email", email))

	w.Write([]byte("Login successful!"))

}

func (h *dbHandler) Register(w http.ResponseWriter, r *http.Request, db *sql.DB, log *slog.Logger) {

	var request RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		helpers.ServeError(err, w, "Invalid JSON format", log, http.StatusBadRequest)
		return
	}

	email := request.Email
	password := request.Password
	verPassword := request.VerPassword

	// TODO: use context for multple errors return

	if !helpers.IsValidEmail(email) {
		helpers.ServeError(errors.New("wrong email format"), w, "wrong email format", log, http.StatusBadRequest)
		return
	}

	if password != verPassword {
		helpers.ServeError(errors.New("passowrd dont match"), w, "passowrd dont match", log, http.StatusBadRequest)
		return
	}

	if email == "" || password == "" {
		helpers.ServeError(errors.New("input is empty"), w, "Input is empty", log, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
		return
	}

	err = createUser(email, string(hashedPassword), db)

	if err != nil {

		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			helpers.ServeError(err, w, "email alreadt exists", log, http.StatusConflict)
			return
		}

		helpers.ServeError(err, w, "Internal server error", log, http.StatusInternalServerError)
		return
	}

	log.Info("User registered:", slog.String("email", email))

	w.Write([]byte("User registered successfully!"))

}

func createUser(email, password string, db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO users (email, password)
		VALUES (?, ?)
	`, email, password)

	if err != nil {
		return err
	}

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email=?", email).Scan(&userID)
	if err != nil {
		return err
	}

	return err
}

func generateJWT(isAdmin bool, id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      id,
		"isAdmin": isAdmin,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

func NewDBHandler() DBHandler {
	return &dbHandler{}
}
