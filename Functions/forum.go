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

	"github.com/gorilla/sessions"   // gérer les sessions utilisateur de manière sécurisée
	_ "github.com/mattn/go-sqlite3" // SQLite3 pour les interactions avec la base de données
	"golang.org/x/crypto/bcrypt"    // utilisé pour le hachage sécurisé des mots de passe
)

// déclare des variables globales pour gérer les sessions et la base de données
var (
	Store = sessions.NewCookieStore([]byte("votre-clé-secrète"))
	db    *sql.DB
)

// Query for fetching main chat data
var queryMain = `
    SELECT 
        c.name AS chat_name, 
        COUNT(m.id) AS message_count, 
        c.descri, 
        r.REGION_IMG_URL, 
        COALESCE(like_counts.total_likes, 0) AS total_likes,
        COALESCE(cl.liked, FALSE) AS user_liked
    FROM 
        chats c
    LEFT JOIN 
        messages m ON c.name = m.chat_name
    LEFT JOIN 
        Region r ON c.region = r.REGION_NAME
    LEFT JOIN (
        SELECT chatID, COUNT(*) AS total_likes
        FROM Chat_Liked
        GROUP BY chatID
    ) like_counts ON c.name = like_counts.chatID
    LEFT JOIN 
        Chat_Liked cl ON c.name = cl.chatID AND cl.Username = ?
    WHERE 
        c.principal = TRUE AND c.region = ?
    GROUP BY c.name, c.descri, r.REGION_IMG_URL, like_counts.total_likes, cl.liked;
`

// Query for fetching user chat data
var queryChats = `
    SELECT 
        c.name AS chat_name, 
        COUNT(m.id) AS message_count, 
        c.descri, 
        u.PHOTO_URL, 
        u.USERNAME, 
        COALESCE(like_counts.total_likes, 0) AS total_likes,
        COALESCE(cl.liked, FALSE) AS user_liked
    FROM 
        chats c
    LEFT JOIN 
        messages m ON c.name = m.chat_name
    LEFT JOIN 
        User u ON c.creator = u.USERNAME
    LEFT JOIN (
        SELECT chatID, COUNT(*) AS total_likes
        FROM Chat_Liked
        GROUP BY chatID
    ) like_counts ON c.name = like_counts.chatID
    LEFT JOIN 
        Chat_Liked cl ON c.name = cl.chatID AND cl.Username = ?
    WHERE 
        c.principal = FALSE AND c.region = ?
    GROUP BY c.name, c.descri, u.PHOTO_URL, u.USERNAME, like_counts.total_likes, cl.liked;
`

// //////////////  STRUCTURES  ////////////////

// prend les regions selon les requêtes sql
type Region struct {
	RegionName  string `json:"region_name"`
	RegionImg   string `json:"region_imgurl"`
	RegionDescr string `json:"region_description"`
}

// prend les fils de discussions créés par l'utilisateur
type ChatInfo struct {
	Name         string `json:"name"`
	MessageCount int    `json:"message_count"`
	Description  string `json:"description"`
	PhotoURL     string `json:"photo_url"`
	Username     string `json:"username"`
}

// prend les chats likés par l'utilisateur
type LikedChat struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	MessageCount int    `json:"message_count"`
	PhotoURL     string `json:"photo_url"`
	Creator      string `json:"creator"`
}

// prend les destinations likés par l'utilisateur pour mettre ou pas le coeur rouge
type LikeRequest struct {
	Region string `json:"region"`
	Liked  bool   `json:"liked"`
}

// prend les chats likés par l'utilisateur pour mettre ou pas le coeur rouge
type LikeChatRequest struct {
	Chat  string `json:"chat"`
	Liked bool   `json:"liked"`
}

// prend les informations sur les différentes regions
type RegionChat struct {
	RegionName  string
	ChatCount   int
	RegionImg   string
	RegionDescr string
	RegionLiked bool
}

// prend le fil de discussion principal de chaque region
type MainChat struct {
	Name         string
	MessageCount int
	Descri       string
	ImageURL     string
	TotalLikes   int
	UserLiked    bool
}

// prend les autres fils de discussion créé par les utilisateurs
type UserChat struct {
	Name         string
	MessageCount int
	Descri       string
	PhotoURL     string
	Creator      string
	TotalLikes   int
	UserLiked    bool
}

// prend les informations sur les messages des chats
type Message struct {
	MessageID     int    `json:"message_id"`
	Sender        string `json:"sender"`
	Message       string `json:"message"`
	TimeElapsed   string `json:"time_elapsed"`
	ImgUser       string `json:"img_user"`
	NumberOfLikes int    `json:"number_of_likes"`
	UserLiked     bool   `json:"user_liked"`
}

// affiche une page d'erreur en utilisant le template HTML donné en paramètre
func renderError(w http.ResponseWriter, tmpl string, errorMsg string) {
	t, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", tmpl))
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	// structure de données pour passer le message d'erreur au template
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: errorMsg,
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

// ouvrir une connexion à la base de données SQLite
func openDatabase() (*sql.DB, error) {
	return sql.Open("sqlite3", "./forum.db")
}

// ////////////////////////////// PAGE DE CRÉATION DE COMPTE ///////////////////////////////////

// vérifie si l'utilisateur existe déjà dans la base de données et renvoie deux booléens (si le email ou le pseudo existent)
func CheckUserExists(db *sql.DB, email, pseudo string) (bool, bool, error) {
	var emailExists, pseudoExists bool
	var id int

	// vérifie si un utilisateur avec cet email existe déjà dans la table "User"
	err := db.QueryRow("SELECT rowid FROM User WHERE EMAIL = ?", email).Scan(&id)
	if err == nil {
		emailExists = true
	} else if err != sql.ErrNoRows {
		return false, false, err
	}

	// vérifie si un utilisateur avec ce pseudo existe déjà dans la table "User"
	err = db.QueryRow("SELECT rowid FROM User WHERE USERNAME = ?", pseudo).Scan(&id)
	if err == nil {
		pseudoExists = true
	} else if err != sql.ErrNoRows {
		return false, false, err
	}

	return emailExists, pseudoExists, nil
}

// vérifie si le mot de passe fourni répond aux critères de sécurité (une lettre majuscule, une lettre minuscule,un chiffre, un caractère spécial et il doit avoir 6 caractères.)
func isValidPassword(password string) bool {
	var ( // initialise les indicateurs pour vérifier les critères
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 6 { // vérifie la longueur
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z': // vérifie s'il contient une majuscule
			hasUpper = true
		case 'a' <= char && char <= 'z': // vérifie s'il contient une minuscule
			hasLower = true
		case '0' <= char && char <= '9': // vérifie s'il contient un chiffre
			hasNumber = true
		case regexp.MustCompile(`[!@#~$%^&*()_+|<>?:{}]`).MatchString(string(char)): // vérifie s'il contient un caractère spécial
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

// prende les info données par l'utilisateur et les rentre dans la base de données
func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// récupération des données du formulaire
	email := r.FormValue("email")
	pseudo := r.FormValue("pseudo")
	motDePasse := r.FormValue("mot_de_passe")
	confirmeMotDePasse := r.FormValue("confirme_mot_de_passe")
	photoURL := r.FormValue("photo_url") // récupérer l'URL de l'avatar choisi

	// valide le format de l'email
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(email) { // renvoie un message d'erreur sinon
		renderError(w, "CreerCompte", "L'adresse email est invalide.")
		return
	}

	// vérifie que les deux mots de passe sont les mêmes
	if motDePasse != confirmeMotDePasse {
		renderError(w, "CreerCompte", "Les mots de passe ne correspondent pas.")
		return
	}

	// vérifie la forme du mot-de-passe
	if !isValidPassword(motDePasse) {
		renderError(w, "CreerCompte", "Le mot de passe doit contenir au minimum\nune majuscule, une minuscule, un caractère spécial, un chiffre, et au minimum 6 caractères.")
		return
	}

	// ouvre la base de données
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// vérifie l'email et le pseudo avec la fonction CheckUserExists()
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

	// chiffre le mot de passe avant de le sauvegarder dans la base de données
	motDePasseChiffre, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors du chiffrement du mot de passe.")
		return
	}

	// Utilisez l'URL de la photo par défaut
	if photoURL == "" {
		photoURL = "static/img/photoProfil.png"
	} else {
		photoURL = strings.TrimPrefix(photoURL, "http://localhost:8080/")
	}

	biographie := "" // initialize la biographie a une string vide

	// sauvegarde les informations dans la table User de la base de donnée
	_, err = db.Exec("INSERT INTO User (USERNAME, PASSWORD, EMAIL, PHOTO_URL, BIOGRAPHY) VALUES (?, ?, ?, ?, ?)", pseudo, motDePasseChiffre, email, photoURL, biographie)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

	// créé une nouvelle session et stocker le nom d'utilisateur
	session, _ := Store.Get(r, "session-name")
	session.Values["username"] = pseudo
	session.Save(r, w)

	// redirige vers la page mytripy-non après la création du compte
	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
}

// ////////////////////////////// PAGE DE CONNEXION ///////////////////////////////////

// vérifie les identifients lors de la connexion
func CheckCredentialsForConnection(w http.ResponseWriter, r *http.Request) {
	var hashedPassword string

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// récupère les données du formulaire
	username := r.FormValue("username")
	password := r.FormValue("password")

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "SeConnecter", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// vérifie l'existence du pseudo
	err = db.QueryRow("SELECT PASSWORD FROM User WHERE USERNAME = ?", username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			renderError(w, "SeConnecter", "Mot de passe ou identifiants introuvables")
		} else {
			http.Error(w, "Erreur interne lors de la vérification des identifiants : "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// compare le mot de passe donné avec celui crypté à l'aide de CompareHashAndPassword
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		renderError(w, "SeConnecter", "Mot de passe ou identifiants introuvables")
		return
	}

	// créer une nouvelle session et stocker le nom d'utilisateur
	session, _ := Store.Get(r, "session-name")
	session.Values["username"] = username
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

// ////////////////////////////// PAGE DE PROFIL ///////////////////////////////////

// execute des requêtes sql
func executeQuery(db *sql.DB, query string, args []interface{}, scanner func(*sql.Rows) error) error { // interface{} permet de stocker des informations de types différents
	// exécute la requête SQL donnée en paramètre
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// parcourt chaque ligne retournée par la requête
	for rows.Next() {
		if err := scanner(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// récupère les régions aimées par un utilisateur
func getLikedRegions(db *sql.DB, username string) ([]Region, error) {
	query := `
        SELECT Region.REGION_NAME, Region.REGION_IMG_URL, Region.DESCRI
        FROM Region
        JOIN USER_LIKES ON Region.REGION_NAME = USER_LIKES.REGION_NAME
        WHERE USER_LIKES.USER_ID = ? AND USER_LIKES.LIKED = TRUE;
    `
	var regions []Region
	// exécute la requête et traite chaque ligne avec la fonction executeQuery()
	err := executeQuery(db, query, []interface{}{username}, func(rows *sql.Rows) error {
		var region Region // structure qui stockee les données de chaque région
		if err := rows.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr); err != nil {
			return err
		}
		regions = append(regions, region) // ajoute la région à la liste
		return nil
	})

	return regions, err // renvoi la liste des régions et une éventuelle erreur
}

// récupère les messages aimées par un utilisateur
func getLikedChats(db *sql.DB, username string) ([]LikedChat, error) {
	query := `
        SELECT c.name, COUNT(m.id) AS message_count, c.descri, u.PHOTO_URL, u.USERNAME AS creator
        FROM chats c
        LEFT JOIN messages m ON c.name = m.chat_name
        LEFT JOIN Chat_Liked cl ON c.name = cl.chatID
        LEFT JOIN User u ON c.creator = u.USERNAME
        WHERE cl.Username = ? AND cl.liked = TRUE
        GROUP BY c.name, c.descri, u.PHOTO_URL, u.USERNAME;
    `
	var likedChats []LikedChat
	// exécute la requête et traite chaque ligne avec une executeQuery()
	err := executeQuery(db, query, []interface{}{username}, func(rows *sql.Rows) error {
		var likedChat LikedChat // structure pour stocker les données de chaque chat aimé
		if err := rows.Scan(&likedChat.Name, &likedChat.MessageCount, &likedChat.Description, &likedChat.PhotoURL, &likedChat.Creator); err != nil {
			return err
		}
		likedChats = append(likedChats, likedChat) // ajoute le chat à la liste de chat aimé
		return nil
	})

	return likedChats, err // renvoi la liste des chats aimés créés et une éventuelle erreur
}

// récupérer les chats créés par un utilisateur
func getUserChats(db *sql.DB, username string) ([]ChatInfo, error) {
	query := `
        SELECT c.name, COUNT(m.id) AS message_count, c.descri, u.PHOTO_URL, u.USERNAME
        FROM chats c
        LEFT JOIN messages m ON c.name = m.chat_name
        LEFT JOIN User u ON c.creator = u.USERNAME
        WHERE c.creator = ?
        GROUP BY c.name, c.descri, u.PHOTO_URL, u.USERNAME;
    `
	var chats []ChatInfo
	// exécute la requête et traite chaque ligne avec une executeQuery()
	err := executeQuery(db, query, []interface{}{username}, func(rows *sql.Rows) error {
		var chat ChatInfo // structure pour stocker les données de chaque chat
		if err := rows.Scan(&chat.Name, &chat.MessageCount, &chat.Description, &chat.PhotoURL, &chat.Username); err != nil {
			return err
		}
		chats = append(chats, chat) // ajoute le chat à la liste
		return nil
	})

	return chats, err // renvoi la liste des chats créés et une éventuelle erreur
}

// afficher la page de profil d'un utilisateur
func ProfilPage(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")          // récupère la session de l'utilisateur
	username, ok := session.Values["username"].(string) // prend le nom d'utilisateur de la session

	if !ok { // vérifie si l'utilisateur est connecté sinon il est redirigé vers la page de connexion
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	// connexion à la base de données
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// récupère les informations de l'utilisateur depuis la base de données
	var pseudo, urlPhoto, biography string
	err = db.QueryRow("SELECT USERNAME, PHOTO_URL, BIOGRAPHY FROM User WHERE USERNAME = ?", username).Scan(&pseudo, &urlPhoto, &biography)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// récupère les régions aimées par l'utilisateur
	regions, err := getLikedRegions(db, username)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des régions : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// récupère les chats créés par l'utilisateur
	chats, err := getUserChats(db, username)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des chats : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch chats liked by the user
	likedChats, err := getLikedChats(db, username)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des chats aimés : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// vérifie si l'utilisateur est connecté
	connected := ok && username != ""

	// préparer les données à transmettre au template
	data := struct {
		Pseudo      string
		PhotoURL    string
		Biography   string
		IsConnected bool
		Regions     []Region
		Chats       []ChatInfo
		LikedChats  []LikedChat
	}{
		Pseudo:      pseudo,
		PhotoURL:    urlPhoto,
		Biography:   biography,
		IsConnected: connected,
		Regions:     regions,
		Chats:       chats,
		LikedChats:  likedChats,
	}

	// charger et exécuter le template pour afficher le profil
	t, err := template.ParseFiles("templates/profil.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

// met à jour les informations du profil de l'utilisateur
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// déclare la structure pour les données JSON envoyées par le client
	type ProfileData struct {
		Pseudo string `json:"pseudo"`
		Bio    string `json:"bio"`
	}

	var data ProfileData
	err := json.NewDecoder(r.Body).Decode(&data) // décoder la requête JSON pour extraire les données du profil
	if err != nil {
		http.Error(w, "Erreur de décodage JSON", http.StatusBadRequest)
		return
	}

	// connexion à la base de données SQLite
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	// met à jour de la biographie de l'utilisateur dans la base de données
	_, err = db.Exec(`UPDATE User SET BIOGRAPHY = ? WHERE USERNAME = ?;`,
		data.Bio, data.Pseudo)

	// Répondre avec un message de succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Profil mis à jour avec succès !"))
}

// met à jour l'avatar de l'utilisateur
func UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// déclare la structure pour les données JSON envoyées par le client
	type AvatarData struct {
		Avatar string `json:"avatar"`
	}

	var data AvatarData
	err := json.NewDecoder(r.Body).Decode(&data) // décoder la requête JSON pour extraire l'URL de l'avatar
	if err != nil {
		http.Error(w, "Erreur lors du décodage du corps de la requête", http.StatusBadRequest)
		return
	}

	// récupérer la session de l'utilisateur
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" { // Vérifie si l'utilisateur est connecté
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	// validate l'URL de l'avatar
	if data.Avatar == "" {
		http.Error(w, "URL d'avatar non valide ou vide", http.StatusBadRequest)
		return
	}

	// connexion à la base de données SQLite
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// met à jour de l'URL de l'avatar dans la base de données
	_, err = db.Exec(`UPDATE User SET PHOTO_URL = ? WHERE USERNAME = ?;`,
		data.Avatar, username)

	// répondre avec un message de succès
	response := map[string]string{
		"message": "Avatar mis à jour avec succès !",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Deconnecte l'utilisateur de sa session
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name") // récupère la session actuelle
	session.Options.MaxAge = -1                // fixe l'âge maximum du cookie de session à -1 pour l'expirer immédiatement
	session.Save(r, w)
	http.Redirect(w, r, "/mytripy-non", http.StatusFound) // redirige l'utilisateur vers la page mytripy-non.html
}

// ///////////////////////////////////////// LIKES ////////////////////////////////////////////////////////
// vérifier si un enregistrement existe déjà dans la base de données
func recordExists(db *sql.DB, query string, args ...interface{}) (bool, error) {
	var exists bool
	err := db.QueryRow(query, args...).Scan(&exists) // exécute la requête SQL et scanne le résultat dans "exists" avec args... qui sépare les valeurs de type ...interface{}
	return exists, err
}

// insérer ou mettre à jour un enregistrement dans la base de données
func insertOrUpdateRecord(db *sql.DB, insertQuery string, updateQuery string, exists bool, args []interface{}) error {
	var err error
	if exists {
		_, err = db.Exec(updateQuery, args...) // Si l'enregistrement existe, exécute une mise à jour avec args... qui sépare les valeurs de type ...interface{}
	} else {
		_, err = db.Exec(insertQuery, args...) // Sinon, insère un nouvel enregistrement avec args... qui sépare les valeurs de type ...interface{}
	}
	return err
}

// gère les likes sur les régions
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var likeRequest struct {
		Region string `json:"region"`
		Liked  bool   `json:"liked"`
	}
	err := json.NewDecoder(r.Body).Decode(&likeRequest)
	if err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Retrieve user session
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, `{"error": "Unauthorized. Please log in."}`, http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check if the like record already exists
	queryExists := `SELECT EXISTS(SELECT 1 FROM USER_LIKES WHERE USER_ID = ? AND REGION_NAME = ?);`
	exists, err := recordExists(db, queryExists, username, likeRequest.Region)
	if err != nil {
		log.Println("Error checking existing like:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	if likeRequest.Liked {
		// Insert or update the like status
		insertQuery := `INSERT INTO USER_LIKES (USER_ID, REGION_NAME, LIKED) VALUES (?, ?, ?);`
		updateQuery := `UPDATE USER_LIKES SET LIKED = ? WHERE USER_ID = ? AND REGION_NAME = ?;`
		err = insertOrUpdateRecord(db, insertQuery, updateQuery, exists, []interface{}{username, likeRequest.Region, likeRequest.Liked})
	} else {
		// Delete the record if disliked
		deleteQuery := `DELETE FROM USER_LIKES WHERE USER_ID = ? AND REGION_NAME = ?;`
		_, err = db.Exec(deleteQuery, username, likeRequest.Region)
	}

	if err != nil {
		log.Println("Error executing SQL query:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Region '%s' liked status updated: %t", likeRequest.Region, likeRequest.Liked),
	})
}

// gère les likes sur les chats
func LikeChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var likeChatRequest struct {
		ChatID string `json:"chat_id"`
		Liked  bool   `json:"liked"`
	}
	err := json.NewDecoder(r.Body).Decode(&likeChatRequest)
	if err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Retrieve user session
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, `{"error": "Unauthorized. Please log in."}`, http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check if the like record already exists
	queryExists := `SELECT EXISTS(SELECT 1 FROM Chat_Liked WHERE Username = ? AND chatID = ?);`
	exists, err := recordExists(db, queryExists, username, likeChatRequest.ChatID)
	if err != nil {
		log.Println("Error checking existing like:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	if likeChatRequest.Liked {
		// Insert or update the like status
		insertQuery := `INSERT INTO Chat_Liked (Username, chatID, liked) VALUES (?, ?, ?);`
		updateQuery := `UPDATE Chat_Liked SET liked = ? WHERE Username = ? AND chatID = ?;`
		err = insertOrUpdateRecord(db, insertQuery, updateQuery, exists, []interface{}{username, likeChatRequest.ChatID, likeChatRequest.Liked})
	} else {
		// Delete the record if disliked
		deleteQuery := `DELETE FROM Chat_Liked WHERE Username = ? AND chatID = ?;`
		_, err = db.Exec(deleteQuery, username, likeChatRequest.ChatID)
	}

	if err != nil {
		log.Println("Error executing SQL query:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Chat ID '%s' liked status updated: %t", likeChatRequest.ChatID, likeChatRequest.Liked),
	})
}

func LikeMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Définit la structure pour le payload
	var likeData struct {
		MessageID int  `json:"message_id"`
		Liked     bool `json:"liked"`
	}

	err := json.NewDecoder(r.Body).Decode(&likeData)
	if err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Récupère la session utilisateur
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, `{"error": "Unauthorized. Please log in."}`, http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	// Connexion à la base de données
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Vérifie si un like existe déjà pour le message
	queryExists := `SELECT EXISTS(SELECT 1 FROM Msg_Liked WHERE Username = ? AND message_id = ?);`
	exists, err := recordExists(db, queryExists, username, likeData.MessageID)
	if err != nil {
		log.Println("Error checking existing like:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	if likeData.Liked {
		// Si "like", insère ou met à jour le statut "LIKED".
		insertQuery := `INSERT INTO Msg_Liked (Username, message_id, liked) VALUES (?, ?, ?);`
		updateQuery := `UPDATE Msg_Liked SET LIKED = ? WHERE Username = ? AND message_id = ?;`
		err = insertOrUpdateRecord(db, insertQuery, updateQuery, exists, []interface{}{username, likeData.MessageID, likeData.Liked})
	} else {
		// Si "dislike", supprime l'enregistrement de la base de données.
		deleteQuery := `DELETE FROM Msg_Liked WHERE Username = ? AND message_id = ?;`
		_, err = db.Exec(deleteQuery, username, likeData.MessageID)
	}

	if err != nil {
		log.Println("Error executing SQL query:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	// Répond avec un message de succès
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Message ID '%d' liked status updated: %t", likeData.MessageID, likeData.Liked),
	})
}

//////////////////////////////////////// PAGE MYTRIPY-NON ////////////////////////////////////////////////////////////

// exécute une requête SQL et scanner les lignes de résultats
func executeQueryAndScan[T any](db *sql.DB, query string, args []interface{}, scanner func(*sql.Rows) (T, error)) ([]T, error) {
	rows, err := db.Query(query, args...) // exécute la requête SQL avec les arguments données en paramètres
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T // initialise une liste pour stocker les objets scannés
	for rows.Next() {
		item, err := scanner(rows) // scanne les résultats
		if err != nil {
			return nil, err
		}
		results = append(results, item) // ajoute l'élément scanné à la liste des résultats
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil // renvoie la liste des résultats et aucune erreur
}

// rend un template HTML
func renderTemplate(w http.ResponseWriter, templateFile string, data interface{}) error {
	tmpl, err := template.ParseFiles(templateFile) // Charge le fichier du template
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data) // exécute le template avec les données fournies en paramètre
}

// affiche les informations nécessaires pour mytripy-non.html
func MyTripyNonHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, connected := session.Values["username"].(string)

	db, err := openDatabase()
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// requête pour récupérer les informations des 3 régions les plus populaires
	query := `
        SELECT r.REGION_NAME, 
               COUNT(c.CHAT_NAME) AS CHAT_COUNT, 
               r.REGION_IMG_URL, 
               r.DESCRI,
               COALESCE(ul.LIKED, FALSE) AS LIKED
        FROM Region r
        JOIN Department d ON r.REGION_NAME = d.REGION_NAME
        JOIN Chat c ON d.DEPARTMENT_NAME = c.DEPARTMENT_NAME
        LEFT JOIN USER_LIKES ul ON r.REGION_NAME = ul.REGION_NAME AND ul.USER_ID = ?
        GROUP BY r.REGION_NAME, r.REGION_IMG_URL, r.DESCRI
        ORDER BY CHAT_COUNT DESC
        LIMIT 3;
    `
	// scanne chaque région
	scanRegionChat := func(rows *sql.Rows) (RegionChat, error) {
		var region RegionChat
		err := rows.Scan(&region.RegionName, &region.ChatCount, &region.RegionImg, &region.RegionDescr, &region.RegionLiked)
		return region, err
	}

	// exécute la requête SQL et scanne les résultats
	regions, err := executeQueryAndScan(db, query, []interface{}{username}, scanRegionChat)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution de la requête.", http.StatusInternalServerError)
		return
	}

	// initialise les données pour le template
	data := struct {
		IsConnected bool
		Regions     []RegionChat
	}{
		IsConnected: connected,
		Regions:     regions,
	}

	// rend le template HTML avec les données
	if err := renderTemplate(w, "templates/mytripy-non.html", data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

///////////////////////////// DESTINATIONS //////////////////////////////////////

// récupère toutes les régions et leur statut aimé par l'utilisateur
func AllRegions(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, isConnected := session.Values["username"].(string)

	db, err := openDatabase()
	if err != nil {
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// requête SQL pour récupérer les informations des régions et leur statut aimé
	query := `
        SELECT Region.REGION_NAME, Region.REGION_IMG_URL, Region.DESCRI, 
               COALESCE(USER_LIKES.LIKED, FALSE) AS LIKED
        FROM Region
        LEFT JOIN USER_LIKES 
        ON Region.REGION_NAME = USER_LIKES.REGION_NAME AND USER_LIKES.USER_ID = ?;
    `
	// scanne chaque ligne retournée par la requête et remplir la structure RegionChat
	scanRegionChat := func(rows *sql.Rows) (RegionChat, error) {
		var region RegionChat
		err := rows.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr, &region.RegionLiked)
		return region, err
	}

	// exécute la requête SQL et scanne les résultats
	regions, err := executeQueryAndScan(db, query, []interface{}{username}, scanRegionChat)
	if err != nil {
		http.Error(w, "Error querying database.", http.StatusInternalServerError)
		return
	}

	// initialise les données à transmettre au template HTML
	data := struct {
		IsConnected bool
		Regions     []RegionChat
	}{
		IsConnected: isConnected,
		Regions:     regions,
	}

	if err := renderTemplate(w, "templates/destinations.html", data); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

///////////////////////// SEARCHBAR ///////////////////////////

// fournit des suggestions de recherche basées sur les entrées utilisateur
func SearchSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q") // récupère la valeur du paramètre de requête "q" pour la recherche
	if len(query) < 2 {             // vérifie que la requête contient au moins 2 caractères
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{}) // retourne un tableau vide si la requête est trop courte
		return
	}

	db, err := openDatabase()
	if err != nil {
		http.Error(w, "Error opening database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	searchPattern := "%" + query + "%"
	sqlQuery := `
        SELECT D.DEPARTMENT_NAME, R.REGION_NAME
        FROM Department D
        JOIN Region R ON D.REGION_NAME = R.REGION_NAME
        WHERE D.DEPARTMENT_NAME LIKE ? OR R.REGION_NAME LIKE ?
        LIMIT 5;
    `
	// scanne les résultats et retourne une suite de chaînes (departmentName,regionName)
	scanSearchResults := func(rows *sql.Rows) (map[string]string, error) {
		var departmentName, regionName string
		err := rows.Scan(&departmentName, &regionName)
		return map[string]string{
			"departmentName": departmentName,
			"regionName":     regionName,
		}, err
	}

	// exécute la requête SQL et scanne les résultats pour remplir les suggestions
	suggestions, err := executeQueryAndScan(db, sqlQuery, []interface{}{searchPattern, searchPattern}, scanSearchResults)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// renvoi les suggestions sous forme de réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// ///////////////////////////////////////////////// FIL DISCUSSION //////////////////////////////////////////////////
// scanner la ligne de fils de discussion principal pour remplir la structure MainChat
func scanMainChat(row *sql.Row) (MainChat, error) {
	var chat MainChat
	err := row.Scan(&chat.Name, &chat.MessageCount, &chat.Descri, &chat.ImageURL, &chat.TotalLikes, &chat.UserLiked)
	return chat, err
}

// scanner les lignes de fils de discussions pour remplir la structure UserChat
func scanUserChat(rows *sql.Rows) (UserChat, error) {
	var chat UserChat
	err := rows.Scan(&chat.Name, &chat.MessageCount, &chat.Descri, &chat.PhotoURL, &chat.Creator, &chat.TotalLikes, &chat.UserLiked)
	return chat, err
}

// exécute une requête et scanne une seule ligne pour le chat principal
func fetchSingleRow[T any](db *sql.DB, query string, args []interface{}, scanner func(*sql.Row) (T, error)) (T, error) {
	row := db.QueryRow(query, args...)
	return scanner(row)
}

// exécute une requête et scanne plusieurs lignes pour les autres chats
func fetchMultipleRows[T any](db *sql.DB, query string, args []interface{}, scanner func(*sql.Rows) (T, error)) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T // initialise un slice pour stocker les résultats
	for rows.Next() {
		item, err := scanner(rows) // scanne chaque ligne et convertit les résultats dans le type T
		if err != nil {
			return nil, err
		}
		results = append(results, item) // ajoute l'objet scanné au slice
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// affiche la discussion principale et les chats utilisateur d'une région spécifique
func FileDiscussion(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
	}

	username, _ := session.Values["username"].(string)
	region, ok := session.Values["region"].(string)
	if !ok || region == "" {
		http.Redirect(w, r, "/region-selection", http.StatusSeeOther)
		return
	}

	db, err := openDatabase()
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// récupère les données du chat principal
	mainChat, err := fetchSingleRow(db, queryMain, []interface{}{username, region}, scanMainChat)
	if err != nil {
		http.Error(w, "Failed to fetch main chat data", http.StatusInternalServerError)
		return
	}

	// rcupère les données de autres chats
	chats, err := fetchMultipleRows(db, queryChats, []interface{}{username, region}, scanUserChat)
	if err != nil {
		http.Error(w, "Failed to fetch user chats data", http.StatusInternalServerError)
		return
	}

	// initialise les données pour le rendu du template
	data := struct {
		IsConnected bool
		Username    string
		Region      string
		MainChat    MainChat
		Chats       []UserChat
	}{
		IsConnected: username != "",
		Username:    username,
		Region:      region,
		MainChat:    mainChat,
		Chats:       chats,
	}

	// rend le template HTML avec les données
	if err := template.Must(template.ParseFiles("templates/welcome.html")).Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// créer un nouveau chat
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	creator, ok := session.Values["username"].(string) // récupère le nom du créateur depuis la session
	if !ok || creator == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	chatName := r.FormValue("chatname")           // récupère le nom du chat depuis le formulaire
	chatDescription := r.FormValue("description") // // récupère la description du chat depuis le formulaire
	region := r.FormValue("region")               // récupère la région où le chat sera créé

	// validation des données d'entrée
	if chatName == "" || region == "" {
		http.Error(w, "Chat name or region missing", http.StatusBadRequest)
		log.Println("Chat creation failed: missing chat name or region.")
		return
	}

	db, err := openDatabase()
	if err != nil {
		log.Println("Database connection error:", err)
		http.Error(w, "Failed to connect to the database.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// insertion du chat dans la base de données
	query := "INSERT INTO chats (name, creator, region, descri, principal) VALUES (?, ?, ?, ?, ?)"
	_, err = db.Exec(query, chatName, creator, region, chatDescription, false)
	if err != nil {
		log.Printf("Chat creation failed: %v", err)
		http.Error(w, "Chat creation failed. Please try again later.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/welcome", http.StatusSeeOther) // redirige vers la page "welcome" après succès
}

// garde le nom du chat sélectionné et le stocker dans la session utilisateur
func SelectChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chatName := r.URL.Query().Get("chatname") // récupère le nom du chat depuis les paramètres de requête
	if chatName == "" {
		http.Error(w, "Chat name missing", http.StatusBadRequest)
		log.Println("Chat selection failed: missing chat name in request.")
		return
	}

	session, err := Store.Get(r, "session-name") // récupère la session utilisateur
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session.", http.StatusInternalServerError)
		return
	}

	session.Values["chatname"] = chatName // stocke le nom du chat dans la session utilisateur
	err = session.Save(r, w)              // sauvegarde la session
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Failed to save session. Please try again.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// récupérer et renvoyer des informations sur les chats d'une région pour envoyer dynamiquement la création des fils de discussions dans toutes les sessions
func FetchChatsHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region") // récupère la region depuis les paramètres de requête
	if region == "" {
		http.Error(w, "Region is required", http.StatusBadRequest)
		return
	}

	session, _ := Store.Get(r, "session-name")         // récupère la session
	username, _ := session.Values["username"].(string) // récupère le nom de l'utilisateur

	db, err := openDatabase()
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// récupère les données du chat principal
	mainChat, err := fetchSingleRow(db, queryMain, []interface{}{username, region}, scanMainChat)
	if err != nil {
		http.Error(w, "Failed to fetch main chat data", http.StatusInternalServerError)
		return
	}

	// récupère les données des autres chats
	chats, err := fetchMultipleRows(db, queryChats, []interface{}{username, region}, scanUserChat)
	if err != nil {
		http.Error(w, "Failed to fetch user chats data", http.StatusInternalServerError)
		return
	}

	// initialise les données pour les envoyer en réponse
	data := struct {
		IsConnected bool
		MainChat    MainChat
		Chats       []UserChat
	}{
		IsConnected: username != "",
		MainChat:    mainChat,
		Chats:       chats,
	}

	// définit le type de contenu de la réponse comme JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil { // encode les données dans un format JSON et les envoie comme réponse HTTP
		http.Error(w, "Error encoding data", http.StatusInternalServerError)
	}
}

// ///////////////////////////////////////////// MESSAGES //////////////////////////////////////////

// récupère des valeurs spécifiques depuis une session
func getSessionValues(session *sessions.Session, keys ...string) (map[string]string, error) {
	values := make(map[string]string) // initialise une map pour stocker les valeurs
	for _, key := range keys {
		value, ok := session.Values[key].(string)
		if !ok || value == "" {
			return nil, fmt.Errorf("Missing or invalid session value: %s", key)
		}
		values[key] = value
	}
	return values, nil
}

// récupère les messages d'un chat spécifique depuis la base de données
func fetchMessages(db *sql.DB, username, chatName string) ([]Message, error) {
	// requête SQL pour récupérer les messages et leurs informations associées
	query := `
        SELECT 
            m.id, 
            m.sender, 
            m.message, 
            strftime('%Y-%m-%d %H:%M:%S', m.timestamp) AS timestamp, 
            u.PHOTO_URL, 
            COALESCE(like_count, 0) AS number_of_likes,
            CASE WHEN ul.message_id IS NOT NULL THEN TRUE ELSE FALSE END AS user_liked
        FROM 
            messages m
        LEFT JOIN 
            User u ON m.sender = u.USERNAME
        LEFT JOIN (
            SELECT message_id, COUNT(*) AS like_count
            FROM Msg_Liked
            GROUP BY message_id
        ) likes ON m.id = likes.message_id
        LEFT JOIN 
            Msg_Liked ul ON m.id = ul.message_id AND ul.username = ?
        WHERE 
            m.chat_name = ?
        ORDER BY 
            m.timestamp ASC;`
	rows, err := db.Query(query, username, chatName) // exécute la requête SQL
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message // initialise une liste pour stocker les messages
	for rows.Next() {
		var msg Message // structure pour stocker un message
		var timestamp string
		if err := rows.Scan(&msg.MessageID, &msg.Sender, &msg.Message, &timestamp, &msg.ImgUser, &msg.NumberOfLikes, &msg.UserLiked); err != nil {
			log.Println("Error scanning message data:", err)
			continue
		}

		// formate l'heure en temps écoulé.
		elapsedTime, err := formatElapsedTime(timestamp)
		if err != nil {
			log.Println("Error formatting timestamp:", err)
			continue
		}
		msg.TimeElapsed = elapsedTime // ajoute le temps écoulé au message

		messages = append(messages, msg) // ajoute le message à la liste
	}
	return messages, nil
}

// valide une session et récupére des clés obligatoires
func validateSession(session *sessions.Session, requiredKeys []string) (map[string]string, error) {
	values := make(map[string]string)
	for _, key := range requiredKeys {
		value, ok := session.Values[key].(string)
		if !ok || value == "" {
			return nil, fmt.Errorf("Session key '%s' is missing or invalid", key)
		}
		values[key] = value
	}
	return values, nil
}

// récupère les messages d'un chat spécifique avec le même fonctionnement que fetchMessages
func getMessages(db *sql.DB, username, chatName string) ([]Message, error) {
	query := `
        SELECT 
            m.id, 
            m.sender, 
            m.message, 
            strftime('%Y-%m-%d %H:%M:%S', m.timestamp) AS timestamp, 
            u.PHOTO_URL, 
            COALESCE(like_count, 0) AS number_of_likes,
            CASE WHEN ul.message_id IS NOT NULL THEN TRUE ELSE FALSE END AS user_liked
        FROM 
            messages m
        LEFT JOIN 
            User u ON m.sender = u.USERNAME
        LEFT JOIN (
            SELECT message_id, COUNT(*) AS like_count
            FROM Msg_Liked
            GROUP BY message_id
        ) likes ON m.id = likes.message_id
        LEFT JOIN 
            Msg_Liked ul ON m.id = ul.message_id AND ul.username = ?
        WHERE 
            m.chat_name = ?
        ORDER BY 
            m.timestamp ASC;`

	rows, err := db.Query(query, username, chatName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var timestamp string
		if err := rows.Scan(&msg.MessageID, &msg.Sender, &msg.Message, &timestamp, &msg.ImgUser, &msg.NumberOfLikes, &msg.UserLiked); err != nil {
			log.Println("Error scanning message data:", err)
			continue
		}

		elapsedTime, err := formatElapsedTime(timestamp)
		if err != nil {
			log.Println("Error formatting timestamp:", err)
			continue
		}
		msg.TimeElapsed = elapsedTime

		messages = append(messages, msg)
	}
	return messages, nil
}

// récupère les messages d'un chat et les afficher sur la page web correspondante
func FilMessagesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	sessionValues, err := getSessionValues(session, "username", "chatname")
	if err != nil {
		// redirige l'utilisateur vers la page de connexion en cas d'erreur de session
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := openDatabase()
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// récupère les messages du chat correspondant
	messages, err := fetchMessages(db, sessionValues["username"], sessionValues["chatname"])
	if err != nil {
		log.Println("Error fetching messages:", err)
		http.Error(w, "Error retrieving messages", http.StatusInternalServerError)
		return
	}

	// initialise les données pour le rendu du template
	data := struct {
		ChatName string
		Messages []Message
	}{
		ChatName: sessionValues["chatname"],
		Messages: messages,
	}

	// Charge le template pour afficher les messages
	tmpl := template.Must(template.ParseFiles("templates/chat_messages.html"))
	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// pour envoyer un message dans une base de données
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	session, err := Store.Get(r, "session-name") // récupère la session
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	sessionValues, err := getSessionValues(session, "username", "chatname") // récupère les valeurs nécessaires de la session avec la fonction getSessionValues()
	if err != nil {
		http.Error(w, "Chat not selected or user not logged in. Please select a chat and log in.", http.StatusUnauthorized)
		return
	}

	message := r.FormValue("message") // récupère le contenu du message
	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		log.Println("Empty message received.")
		return
	}

	// récupère l'heure actuelle au format "YYYY-MM-DD HH:MM:SS"
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	db, err := openDatabase()
	if err != nil {
		http.Error(w, "Failed to connect to the database.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// insère le message dans la table messages
	_, err = db.Exec(
		"INSERT INTO messages (chat_name, sender, message, timestamp) VALUES (?, ?, ?, ?)",
		sessionValues["chatname"], sessionValues["username"], message, currentTime,
	)
	if err != nil {
		log.Println("Error saving message to database:", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// récupère les messages d'un chat et les retourner sous forme de réponse JSON
func FetchMessagesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	requiredKeys := []string{"username", "chatname"}
	sessionValues, err := validateSession(session, requiredKeys)
	if err != nil {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := openDatabase()
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// récupère les messages du chat spécifié dans la base de données
	messages, err := getMessages(db, sessionValues["username"], sessionValues["chatname"])
	if err != nil {
		log.Println("Error fetching messages:", err)
		http.Error(w, "Server error while fetching messages", http.StatusInternalServerError)
		return
	}

	// Retourne les messages sous forme de réponse JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		log.Println("Error encoding messages:", err)
		http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
	}
}

// formater l'heure en temps écoulé
func formatElapsedTime(input string) (string, error) {
	// définit le format du temps pour la conversion
	layout := "2006-01-02 15:04:05"
	inputTime, err := time.Parse(layout, input) // convertit l'entrée en objet time.Time
	if err != nil {
		return "", err
	}

	// récupère l'heure actuelle
	currentTime := time.Now()

	// calcule la différence
	duration := currentTime.Sub(inputTime)

	// formate en fonction de les années, les mois, les jours ou les heures
	years := int(duration.Hours() / (24 * 365))
	if years > 0 {
		return fmt.Sprintf("%d ans", years), nil
	}

	months := int(duration.Hours() / (24 * 30))
	if months > 0 {
		return fmt.Sprintf("%d mois", months), nil
	}

	days := int(duration.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dj", days), nil
	}

	hours := int(duration.Hours())
	if hours > 0 {
		return fmt.Sprintf("%dh", hours), nil
	}
	timeStr := inputTime.Format("15:04")
	// l'heure de défaut est celle de l'envoi du message
	return timeStr, nil
}

// //////////////////////////////////////// REGION ///////////////////////////////////////////////

// sauvegarde le nom de la région sélectionnée par l'utilisateur dans la session
func RegionHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("name") // récupère le nom de la région depuis les paramètres de requête HTTP
	if region == "" {
		http.Error(w, "Region not selected. Please choose a region.", http.StatusBadRequest)
		log.Println("No region selected.")
		return
	}

	session, err := Store.Get(r, "session-name") // récupère la session
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	session.Values["region"] = region // sauvegarde le nom de la région dans la session
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// redirige l'utilisateur vers la page welcome.html
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}
