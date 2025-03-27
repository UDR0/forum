package forum

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

func renderError(w http.ResponseWriter, tmpl string, errorMsg string) {
	t, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", tmpl))
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: errorMsg,
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

func CheckUserExists(db *sql.DB, email, pseudo string) (bool, bool, error) {
	var emailExists, pseudoExists bool
	var id int

	err := db.QueryRow("SELECT rowid FROM User WHERE EMAIL = ?", email).Scan(&id)
	if err == nil {
		emailExists = true
	} else if err != sql.ErrNoRows {
		return false, false, err
	}

	err = db.QueryRow("SELECT rowid FROM User WHERE USERNAME = ?", pseudo).Scan(&id)
	if err == nil {
		pseudoExists = true
	} else if err != sql.ErrNoRows {
		return false, false, err
	}

	return emailExists, pseudoExists, nil
}

func isValidPassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(password) >= 6 {
		hasMinLen = true
	}
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case regexp.MustCompile(`[!@#~$%^&*()_+|<>?:{}]`).MatchString(string(char)):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	pseudo := r.FormValue("pseudo")
	motDePasse := r.FormValue("mot_de_passe")
	confirmeMotDePasse := r.FormValue("confirme_mot_de_passe")
	photoURL := r.FormValue("photo_url") // Récupérer l'URL de l'avatar choisi

	// Validation du format de l'email
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(email) {
		renderError(w, "CreerCompte", "L'adresse email est invalide.")
		return
	}

	if motDePasse != confirmeMotDePasse {
		renderError(w, "CreerCompte", "Les mots de passe ne correspondent pas.")
		return
	}

	if !isValidPassword(motDePasse) {
		renderError(w, "CreerCompte", "Le mot de passe doit contenir au minimum\nune majuscule, une minuscule, un caractère spécial, un chiffre, et au minimum 6 caractères.")
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	emailExists, pseudoExists, err := CheckUserExists(db, email, pseudo)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la vérification des utilisateurs existants.")
		return
	}
	if emailExists {
		renderError(w, "CreerCompte", "L'email est déjà utilisé.")
		return
	}
	if pseudoExists {
		renderError(w, "CreerCompte", "Le pseudo est déjà utilisé.")
		return
	}

	motDePasseChiffre, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors du chiffrement du mot de passe.")
		return
	}

	emailChiffre, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors du chiffrement de l'email.")
		return
	}

	// Utilisez l'URL de l'avatar choisi ou une URL de photo par défaut
	if photoURL == "" {
		photoURL = "static/img/avatar/avatarFemme1.png"
	} else {
		photoURL = strings.TrimPrefix(photoURL, "http://localhost:8080/")
	}

	biographie := ""

	_, err = db.Exec("INSERT INTO User (USERNAME, PASSWORD, EMAIL, PHOTO_URL, BIOGRAPHY) VALUES (?, ?, ?, ?, ?)", pseudo, motDePasseChiffre, emailChiffre, photoURL, biographie)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

	fmt.Fprintln(w, "Compte créé avec succès")
}

func CheckCredentialsForConnection(w http.ResponseWriter, r *http.Request) {
	var hashedPassword string
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "SeConnecter", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	err = db.QueryRow("SELECT PASSWORD FROM User WHERE USERNAME = ?", username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			renderError(w, "SeConnecter", "Mot de passe ou identifiants introuvables")
		} else {
			http.Error(w, "Erreur interne lors de la vérification des identifiants : "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		renderError(w, "SeConnecter", "Mot de passe ou identifiants introuvables")
		return
	}

	//Changer cette ligne par la page qui est renvoyée au site avec sa connection + renvoyer true au js pour display les truc cachés
	//fmt.Fprintln(w, "Compte existe")
	http.Redirect(w, r, "/", http.StatusFound)
}

func AddChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.") /////////////////// change the name page html
		return
	}
	defer db.Close()

	nameChat := r.FormValue("nameChat")
	nameDep := r.FormValue("nameDep")
	chatType := r.FormValue("chatType")

	////////////////////////////questo non so se è meglio metterlo nello js e poi prenderlo da li o inizializzarlo direttamente qua
	chatDateTime := time.Now()

	_, err = db.Exec("INSERT INTO Chat (CHAT_NAME, DEPARTMENT_NAME, CHAT_DATETIME, CHAT_TYPE) VALUES (?, ?, ?, ?)", nameChat, nameDep, chatDateTime, chatType)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de l'ajout du message.") /////////////////// change the name page html
		return
	}

}

// func AddMessage(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	db, err := sql.Open("sqlite3", "./forum.db")
// 	if err != nil {
// 		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.") /////////////////// change the name page html
// 		return
// 	}
// 	defer db.Close()

// 	nameChat := r.FormValue("nameChat")
// 	username := r.FormValue("username")
// 	msg := r.FormValue("msg")

// 	////////////////////////////questo non so se è meglio metterlo nello js e poi prenderlo da li o inizializzarlo direttamente qua
// 	datetime := time.Now()

// 	_, err = db.Exec("INSERT INTO Message (CHAT_NAME, USERNAME, MESSAGE_DATETIME, MESSAGE_TEXT) VALUES (?, ?, ?, ?)", nameChat, username, datetime, msg)
// 	if err != nil {
// 		renderError(w, "CreerCompte", "Erreur lors de l'ajout du message.") /////////////////// change the name page html
// 		return
// 	}

// }

// /// this gets all the infos from the table Message poi posso fare la stessa cosa per elencare le chat

// func GetMessages(db *sql.DB) ([]struct {
// 	ID       int
// 	ChatName string
// 	Username string
// 	Datetime time.Time
// 	Text     string
// }, error) {
// 	rows, err := db.Query("SELECT MESSAGE_ID, CHAT_NAME, USERNAME, MESSAGE_DATETIME, MESSAGE_TEXT FROM Message")
// 	if err != nil {
// 		return nil, fmt.Errorf("error querying messages: %v", err)
// 	}
// 	defer rows.Close()

// 	var messages []struct {
// 		ID       int
// 		ChatName string
// 		Username string
// 		Datetime time.Time
// 		Text     string
// 	}

// 	for rows.Next() {
// 		var message struct {
// 			ID       int
// 			ChatName string
// 			Username string
// 			Datetime time.Time
// 			Text     string
// 		}

// 		if err := rows.Scan(&message.ID, &message.ChatName, &message.Username, &message.Datetime, &message.Text); err != nil {
// 			return nil, fmt.Errorf("error scanning message row: %v", err)
// 		}
// 		messages = append(messages, message)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error with rows: %v", err)
// 	}

// 	return messages, nil
// }
