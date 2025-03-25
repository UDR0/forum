package forum

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"text/template"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

// ////////TO DO LIST////////////////
// get the infos from the database (ex => GetImageURLFromDB())
// write into the database (ex => WriteImageIntoDB())
// make sure all the tables in the database are right //
// write all the url to get to the images

func GetImageURLFromDB() string {
	// Open the SQLite database located at /forum.db
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return ""
	}
	defer db.Close()

	// Query the image_url for id_region = 8
	var imageURL string
	err = db.QueryRow("SELECT image_url FROM regions WHERE id_region = 8").Scan(&imageURL)
	if err != nil {
		fmt.Printf("Failed to query image_url: %v\n", err)
		return ""
	}

	return imageURL
}

func WriteImageIntoDB() {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Write (INSERT) data into the database
	query := `SELECT * FROM regions; UPDATE regions SET image_url = 'peoedj' WHERE id_region = 1;`
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return
	}

	fmt.Println("Data successfully inserted into the database!")
}

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

	err := db.QueryRow("SELECT rowid FROM user WHERE MAIL = ?", email).Scan(&id)
	if err == nil {
		emailExists = true
	} else if err != sql.ErrNoRows {
		return false, false, err
	}

	err = db.QueryRow("SELECT rowid FROM user WHERE PSEUDO = ?", pseudo).Scan(&id)
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

	// Utilisez une URL de photo par défaut
	urlPhoto := "static/img/photoProfil.png"
	biographie := ""

	_, err = db.Exec("INSERT INTO user (PSEUDO, MOT_DE_PASSE, MAIL, URL_PHOTO, BIOGRAPHIE) VALUES (?, ?, ?, ?, ?)", pseudo, motDePasseChiffre, emailChiffre, urlPhoto, biographie)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

	fmt.Fprintln(w, "Compte créé avec succès")
}

/*    HOW TO SHOW THE IMAGE ON THE REGION.TMPL FILE BUT I NEED TO CHANGE THE NAME IN THE DATABASE
func forumHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "region.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		ImageURL string
	}{
		ImageURL: getImageURLFromDB(),
	}

	tmpl.Execute(w, data)
}
*/
