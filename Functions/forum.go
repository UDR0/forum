package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/sessions" //go get github.com/gorilla/sessions
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	//"golang.org/x/exp/rand"
)

// Exporter le magasin de sessions
var Store = sessions.NewCookieStore([]byte("votre-clé-secrète"))

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
	var emailInDB string
	var pseudoExists bool
	var id int

	// Étape 1 : Vérifier l'existence de l'e-mail
	log.Println("Vérification de l'email dans la base de données...")
	err := db.QueryRow("SELECT EMAIL FROM User WHERE EMAIL = ?", email).Scan(&emailInDB)
	if err == nil {
		if emailInDB == email {
			log.Println("L'email existe :", emailInDB)
		}
	} else if err == sql.ErrNoRows {
		log.Println("Email non trouvé.")
	} else {
		log.Println("Erreur lors de la vérification de l'email :", err)
		return false, false, err
	}

	// Étape 2 : Vérifier l'existence du pseudo
	log.Println("Vérification du pseudo dans la base de données...")
	err = db.QueryRow("SELECT rowid FROM User WHERE USERNAME = ?", pseudo).Scan(&id)
	if err == nil {
		pseudoExists = true
		log.Println("Le pseudo existe :", pseudoExists)
	} else if err == sql.ErrNoRows {
		log.Println("Pseudo non trouvé.")
	} else {
		log.Println("Erreur lors de la vérification du pseudo :", err)
		return false, false, err
	}

	return emailInDB != "", pseudoExists, nil
}

func GetUserDetails(db *sql.DB, email, pseudo string) (bool, bool, string, string, error) {
	var emailInDB, pseudoInDB string
	var emailExists, pseudoExists bool

	// Vérifier l'existence de l'email et le récupérer
	err := db.QueryRow("SELECT EMAIL FROM User WHERE EMAIL = ?", email).Scan(&emailInDB)
	if err == nil {
		emailExists = true
	} else if err != sql.ErrNoRows {
		emailInDB = ""
	} else {
		return false, false, "", "", err
	}

	// Vérifier l'existence du pseudo et le récupérer
	err = db.QueryRow("SELECT USERNAME FROM User WHERE USERNAME = ?", pseudo).Scan(&pseudoInDB)
	if err == nil {
		pseudoExists = true
	} else if err != sql.ErrNoRows {
		pseudoInDB = ""
	} else {
		return false, false, "", "", err
	}

	return emailExists, pseudoExists, emailInDB, pseudoInDB, nil
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
	photoURL := r.FormValue("photo_url")

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

	// Utilisez l'URL de l'avatar choisi ou une URL de photo par défaut
	if photoURL == "" {
		photoURL = "static/img/avatar/avatarFemme1.png"
	} else {
		photoURL = strings.TrimPrefix(photoURL, "http://localhost:8080/")
	}

	biographie := ""

	// Insérer l'email en clair dans la base de données
	_, err = db.Exec("INSERT INTO User (USERNAME, PASSWORD, EMAIL, PHOTO_URL, BIOGRAPHY) VALUES (?, ?, ?, ?, ?)", pseudo, motDePasseChiffre, email, photoURL, biographie)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

	// Créer une nouvelle session et stocker le nom d'utilisateur
	session, _ := Store.Get(r, "session-name")
	session.Values["username"] = pseudo
	session.Save(r, w)

	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
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

	// Créer une nouvelle session et stocker le nom d'utilisateur
	session, _ := Store.Get(r, "session-name")
	session.Values["username"] = username
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

// Page de profil pour afficher les informations utilisateur
func ProfilPage(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)

	if !ok {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var pseudo, urlPhoto, biography string
	err = db.QueryRow("SELECT USERNAME, PHOTO_URL, BIOGRAPHY FROM User WHERE USERNAME = ?", username).Scan(&pseudo, &urlPhoto, &biography)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Pseudo    string
		PhotoURL  string
		Biography string
	}{
		Pseudo:    pseudo,
		PhotoURL:  urlPhoto,
		Biography: biography,
	}

	t, err := template.ParseFiles("templates/profil.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
}

func AddChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.") // Changez le nom de la page HTML
		return
	}
	defer db.Close()

	nameChat := r.FormValue("nameChat")
	nameDep := r.FormValue("nameDep")
	chatType := r.FormValue("chatType")

	chatDateTime := time.Now()

	_, err = db.Exec("INSERT INTO Chat (CHAT_NAME, DEPARTMENT_NAME, CHAT_DATETIME, CHAT_TYPE) VALUES (?, ?, ?, ?)", nameChat, nameDep, chatDateTime, chatType)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de l'ajout du message.") // Changez le nom de la page HTML
		return
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session-name")
		_, ok := session.Values["username"].(string)
		if !ok {
			http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func MyTripyNonHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	var pseudo, urlPhoto string

	if ok {
		db, err := sql.Open("sqlite3", "./forum.db")
		if err != nil {
			http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		err = db.QueryRow("SELECT USERNAME, PHOTO_URL FROM User WHERE USERNAME = ?", username).Scan(&pseudo, &urlPhoto)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	type RegionChat struct {
		RegionName  string
		ChatCount   int
		RegionImg   string
		RegionDescr string
	}

	data := struct {
		IsAuthenticated bool
		Pseudo          string
		URLPhoto        string
		Regions         []RegionChat
	}{
		IsAuthenticated: ok,
		Pseudo:          pseudo,
		URLPhoto:        urlPhoto,
	}

	// Fetch popular regions
	db, err := sql.Open("sqlite3", "./forum.db") // Adjust connection details
	if err != nil {

		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
        SELECT r.REGION_NAME, COUNT(c.CHAT_NAME) AS CHAT_COUNT, r.REGION_IMG_URL, r.DESCRI
        FROM Region r
        JOIN Department d ON r.REGION_NAME = d.REGION_NAME
        JOIN Chat c ON d.DEPARTMENT_NAME = c.DEPARTMENT_NAME
        GROUP BY r.REGION_NAME, r.REGION_IMG_URL, r.DESCRI
        ORDER BY CHAT_COUNT DESC
        LIMIT 3;
    `

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Erreur lors de l'exécution de la requête.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var region RegionChat
		if err := rows.Scan(&region.RegionName, &region.ChatCount, &region.RegionImg, &region.RegionDescr); err != nil {
			http.Error(w, "Erreur lors du scan des résultats.", http.StatusInternalServerError)
			return
		}
		data.Regions = append(data.Regions, region)
	}

	// Render the template with the data
	tmpl, err := template.ParseFiles("templates/mytripy-non.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func AllRegions(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	var pseudo, urlPhoto string

	if ok {
		db, err := sql.Open("sqlite3", "./forum.db")
		if err != nil {
			http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		err = db.QueryRow("SELECT USERNAME, PHOTO_URL FROM User WHERE USERNAME = ?", username).Scan(&pseudo, &urlPhoto)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	type RegionChat struct {
		RegionName  string
		RegionImg   string
		RegionDescr string
	}

	data := struct {
		IsAuthenticated bool
		Pseudo          string
		URLPhoto        string
		Regions         []RegionChat
	}{
		IsAuthenticated: ok,
		Pseudo:          pseudo,
		URLPhoto:        urlPhoto,
	}

	// Fetch popular regions
	db, err := sql.Open("sqlite3", "./forum.db") // Adjust connection details
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
        SELECT REGION_NAME, REGION_IMG_URL, DESCRI
        FROM Region;
    `

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution de la requête.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var region RegionChat
		if err := rows.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr); err != nil {
			http.Error(w, "Erreur lors du scan des résultats.", http.StatusInternalServerError)
			return
		}
		data.Regions = append(data.Regions, region)
	}

	// Render the template with the data
	tmpl, err := template.ParseFiles("templates/destinations.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

func RegionPage(w http.ResponseWriter, r *http.Request) {
	// Récupérer le paramètre "region" depuis l'URL
	regionName := strings.TrimSpace(r.URL.Query().Get("region")) // Suppression des espaces inutiles
	if regionName == "" {
		http.Error(w, "Nom de région manquant dans l'URL.", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Vérifier si la région existe et récupérer ses informations
	var regionDetails struct {
		Name  string
		Image string
		Fils  string
	}
	err = db.QueryRow("SELECT REGION_NAME, REGION_IMG_URL, FILS FROM Region WHERE REGION_NAME = ?", regionName).
		Scan(&regionDetails.Name, &regionDetails.Image, &regionDetails.Fils)
	if err == sql.ErrNoRows {
		// Si aucune ligne n'est retournée, la région n'existe pas
		http.Error(w, "La région spécifiée n'existe pas.", http.StatusNotFound)
		return
	} else if err != nil {
		// Loguer l'erreur pour des informations plus détaillées
		log.Printf("Erreur lors du Scan : %v", err)
		http.Error(w, "Erreur lors de la récupération des informations de la région.", http.StatusInternalServerError)
		return
	}

	// Préparer les données pour le template
	type Chat struct {
		Name     string
		DateTime string
	}
	data := struct {
		Name  string
		Image string
		Fils  string
		Chats []Chat
	}{
		Name:  regionDetails.Name,
		Image: regionDetails.Image,
		Fils:  regionDetails.Fils,
	}

	// Charger et exécuter le template filsDiscussion.html
	tmpl, err := template.ParseFiles("templates/filsDiscussion.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

// Fonction utilitaire pour vérifier si une valeur existe dans un tableau
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func SearchSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q") // prend la requete qui suit le 'q' dans l'URL
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// requete sql aui donne les regions et departements qui contiennent la requete de l'utilisateur
	sqlQuery := `
    SELECT D.DEPARTMENT_NAME, R.REGION_NAME
    FROM Department D
    JOIN Region R ON D.REGION_NAME = R.REGION_NAME
    WHERE D.DEPARTMENT_NAME LIKE ? OR R.REGION_NAME LIKE ?
    LIMIT 5;
    `

	rows, err := db.Query(sqlQuery, "%"+query+"%", "%"+query+"%") // Exécute la requête SQL avec le terme à chercher parmi les departements et regions
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Suggestion struct {
		DepartmentName string `json:"department_name"`
		RegionName     string `json:"region_name"`
	}
	var suggestions []Suggestion // Crée un tableau pour stocker les suggestions

	for rows.Next() { // Parcourt les lignes renvoyées par la requête SQL
		var suggestion Suggestion
		if err := rows.Scan(&suggestion.DepartmentName, &suggestion.RegionName); err != nil { // Récupère les données des colonnes dans la structure suggestion
			http.Error(w, "Error scanning rows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		suggestions = append(suggestions, suggestion) // Ajoute la suggestion au tableau
	}

	w.Header().Set("Content-Type", "application/json") // Avertit que la reponse est en JSON
	json.NewEncoder(w).Encode(suggestions)             // Encode la reponse en JSON
}

//////////////////////////////////// PROFIL /////////////////////////////////////////////////////////

type UpdateProfileRequest struct {
	Pseudo string `json:"pseudo"`
	Bio    string `json:"bio"`
	Avatar string `json:"avatar"`
}

// Fonction pour mettre à jour les informations utilisateur dans la base de données
func updateUserInDB(db *sql.DB, userID int, pseudo, bio, avatar string) error {
	// Construire dynamiquement la requête SQL en fonction des champs renseignés
	query := "UPDATE User SET"
	params := []interface{}{}
	if pseudo != "" {
		query += " USERNAME = ?,"
		params = append(params, pseudo)
	}
	if bio != "" {
		query += " BIOGRAPHY = ?,"
		params = append(params, bio)
	}
	if avatar != "" {
		query += " PHOTO_URL = ?,"
		params = append(params, avatar)
	}

	// Supprimer la dernière virgule et ajouter la condition WHERE
	query = query[:len(query)-1] + " WHERE rowid = ?"
	params = append(params, userID)

	// Exécuter la requête SQL
	result, err := db.Exec(query, params...)
	if err != nil {
		return err
	}

	// Vérifier combien de lignes ont été affectées
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// API pour mettre à jour le profil utilisateur
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	session, _ := Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Utilisateur non connecté", http.StatusUnauthorized)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT rowid FROM User WHERE USERNAME = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	// Mise à jour des informations dans la base de données
	err = updateUserInDB(db, userID, req.Pseudo, req.Bio, req.Avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Aucune mise à jour effectuée", http.StatusBadRequest)
		} else {
			http.Error(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		}
		return
	}

	// Répondre avec succès
	w.WriteHeader(http.StatusOK)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Email           string `json:"email"`
		Username        string `json:"username"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur lors de la connexion à la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Utiliser la nouvelle fonction pour vérifier et récupérer les détails utilisateur
	emailExists, pseudoExists, emailInDB, pseudoInDB, err := GetUserDetails(db, requestData.Email, requestData.Username)
	if err != nil {
		http.Error(w, "Utilisateur introuvable.", http.StatusInternalServerError)
		return
	}
	if !emailExists || !pseudoExists {
		http.Error(w, "Utilisateur introuvable.", http.StatusBadRequest)
		return
	}

	// Validation des mots de passe
	if requestData.NewPassword != requestData.ConfirmPassword {
		http.Error(w, "Les mots de passe ne correspondent pas.", http.StatusBadRequest)
		return
	}
	if !isValidPassword(requestData.NewPassword) {
		http.Error(w, "Le mot de passe ne respecte pas les critères de sécurité.\nune majuscule, une minuscule, un caractère spécial, un chiffre, et au minimum 6 caractères.", http.StatusBadRequest)
		return
	}

	// Hacher le nouveau mot de passe
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(requestData.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erreur lors du hachage du mot de passe.", http.StatusInternalServerError)
		return
	}

	// Mettre à jour le mot de passe dans la base de données
	_, err = db.Exec("UPDATE User SET PASSWORD = ? WHERE EMAIL = ? AND USERNAME = ?", passwordHash, emailInDB, pseudoInDB)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour du mot de passe.", http.StatusInternalServerError)
		return
	}

	// Répondre avec succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Mot de passe réinitialisé avec succès."))
}
