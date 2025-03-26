package forum

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// ////////TO DO LIST////////////////
// get the infos from the database (ex => GetImageURLFromDB())
// write into the database (ex => WriteImageIntoDB())
// make sure all the tables in the database are right (don't want to)
// write all the url to get to the images

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

func CheckUserExists(db *sql.DB, query string, value string) (bool, error) {
	var elemCount int
	err := db.QueryRow(query, value).Scan(&elemCount)
	if err != nil {
		return false, fmt.Errorf("error checking existence: %v", err)
	}
	return elemCount > 0, nil
}

func isValidPassword(password string) bool {
	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		if 'A' <= char && char <= 'Z' {
			hasUpper = true
		} else if 'a' <= char && char <= 'z' {
			hasLower = true
		} else if '0' <= char && char <= '9' {
			hasNumber = true
		} else if regexp.MustCompile(`[!@#~$%^&*()_+|<>?:{}]`).MatchString(string(char)) {
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial && len(password) >= 6
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
		renderError(w, "CreerCompte", "Le mot de passe doit contenir au minimum<br>une majuscule, une minuscule, un caractère spécial, un chiffre, et au minimum 6 caractères.")
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	pseudoExists, err := CheckUserExists(db, `SELECT COUNT(*) FROM User WHERE USERNAME = ?`, pseudo)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la vérification des utilisateurs existants.")
		return
	}

	emailExists, err := CheckUserExists(db, `SELECT COUNT(*) FROM User WHERE EMAIL = ?`, email)
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

	// Utilisez une URL de photo par défaut
	urlPhoto := "static/img/photoProfil.png"

	_, err = db.Exec("INSERT INTO user (USERNAME, PASSWORD, EMAIL, PHOTO_URL) VALUES (?, ?, ?, ?)", pseudo, motDePasseChiffre, emailChiffre, urlPhoto)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

	//Changer cette ligne par la page qui est renvoyée en cas de succès de création de compte
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

	//Changer cette ligne par la page qui est renvoyée au site avec sa connection
	fmt.Fprintln(w, "Compte existe")
}

func AddMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	nameChat := r.FormValue("nameChat")
	nameDep := r.FormValue("nameDep")
	chatType := r.FormValue("chatType")

	////////////////////////////questo non so se è meglio metterlo nello js e poi prenderlo da li o inizializzarlo direttamente qua
	chatDateTime := time.Now()

	_, err = db.Exec("INSERT INTO user (CHAT_NAME, DEPARTMENT_NAME, CHAT_DATETIME, CHAT_TYPE) VALUES (?, ?, ?, ?)", nameChat, nameDep, chatDateTime, chatType)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

}
