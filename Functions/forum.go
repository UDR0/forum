package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Exporter le magasin de sessions
var (
	Store = sessions.NewCookieStore([]byte("votre-clé-secrète"))
	db    *sql.DB
)

// renvoie les message d'erreur
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

// //////////////////////////////////////// PAGE DE CREATION DE COMPTE //////////////////////////////////////////////////

// Regarde si l'utilisateur existe
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

// Vérifie que le mot de passe respecte les règles de sécurité
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

// Prend les informations données lors de la création de compte et les enregistre dans la base de données
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

	biographie := ""

	// Enregistre le tout dans la table "User"
	_, err = db.Exec("INSERT INTO User (USERNAME, PASSWORD, EMAIL, PHOTO_URL, BIOGRAPHY) VALUES (?, ?, ?, ?, ?)", pseudo, motDePasseChiffre, email, photoURL, biographie)
	if err != nil {
		renderError(w, "CreerCompte", "Erreur lors de la création du compte.")
		return
	}

	// Créer une nouvelle session et stocker le nom d'utilisateur
	session, _ := Store.Get(r, "session-name")
	session.Values["username"] = pseudo
	session.Save(r, w)

	// Rediriger vers la page mytripy-non après la création du compte
	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
}

// //////////////////////////////////////// PAGE DE CONNEXION //////////////////////////////////////////////////

// S'assure que les identifients sont correctent pour la connexion
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

	// Compare le mot de passe fournit avec celui crypté
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

// //////////////////////////////////////// PAGE DE PROFIL //////////////////////////////////////////////////

// Fonction qui renvoie à la page les différentes informations dont il a besoin
func ProfilPage(w http.ResponseWriter, r *http.Request) {
	type Region struct { // Les régions qu'il a aimé
		RegionName  string `json:"region_name"`
		RegionImg   string `json:"region_imgurl"`
		RegionDescr string `json:"region_description"`
	}

	type ChatInfo struct { // Les fils de discussions qu'il a créé
		Name         string `json:"name"`
		MessageCount int    `json:"message_count"`
		Description  string `json:"description"`
		PhotoURL     string `json:"photo_url"`
		Username     string `json:"username"`
	}

	type LikedChat struct { // Les fils de discussions qu'il a aimé
		Name         string `json:"name"`
		Description  string `json:"description"`
		MessageCount int    `json:"message_count"`
		PhotoURL     string `json:"photo_url"`
		Creator      string `json:"creator"`
	}

	var connected bool
	session, _ := Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string) // Prend le nom de l'utilisateur depuis la session

	if !ok {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prend les info de l'utilisateur
	var pseudo, urlPhoto, biography string
	err = db.QueryRow("SELECT USERNAME, PHOTO_URL, BIOGRAPHY FROM User WHERE USERNAME = ?", username).Scan(&pseudo, &urlPhoto, &biography)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prend les régions aimé par l'utilisateur
	queryRegions := `
        SELECT Region.REGION_NAME, Region.REGION_IMG_URL, Region.DESCRI
        FROM Region
        JOIN USER_LIKES ON Region.REGION_NAME = USER_LIKES.REGION_NAME
        WHERE USER_LIKES.USER_ID = ? AND USER_LIKES.LIKED = TRUE;
    `
	rowsRegions, err := db.Query(queryRegions, username)
	if err != nil {
		fmt.Println("Error executing query:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rowsRegions.Close()

	var regions []Region // sauvegarde les infos dans "regions"
	for rowsRegions.Next() {
		var region Region
		err := rowsRegions.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		regions = append(regions, region)
	}

	if err = rowsRegions.Err(); err != nil {
		fmt.Println("Error during row iteration:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prend les fils de discussions créés par l'utilisateur connecté
	queryChats := `
        SELECT c.name, COUNT(m.id) AS message_count, c.descri, u.PHOTO_URL, u.USERNAME
        FROM chats c
        LEFT JOIN messages m ON c.name = m.chat_name
        LEFT JOIN User u ON c.creator = u.USERNAME
        WHERE c.creator = ? 
        GROUP BY c.name, c.descri, u.PHOTO_URL, u.USERNAME;
    `
	rowsChats, err := db.Query(queryChats, username)
	if err != nil {
		fmt.Println("Error executing chat query:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rowsChats.Close()

	var chats []ChatInfo // Sauvegarde les fils de discussion dans "chats"
	for rowsChats.Next() {
		var chat ChatInfo
		err := rowsChats.Scan(&chat.Name, &chat.MessageCount, &chat.Description, &chat.PhotoURL, &chat.Username)
		if err != nil {
			fmt.Println("Error scanning chat row:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		chats = append(chats, chat)
	}

	if err = rowsChats.Err(); err != nil {
		fmt.Println("Error during chat row iteration:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prend les fils de discussions aimés
	queryLikedChats := `
        SELECT 
    	c.name, COUNT(m.id) AS message_count, c.descri, 
    	CASE 
        	WHEN c.principal = 1 THEN r.REGION_IMG_URL 
        	ELSE u.PHOTO_URL 
    	END AS PHOTO_URL, c.creator AS creator
		FROM chats c
		LEFT JOIN messages m ON c.name = m.chat_name
		LEFT JOIN Chat_Liked cl ON c.name = cl.chatID
		LEFT JOIN User u ON c.creator = u.USERNAME
		LEFT JOIN Region r ON c.region = r.REGION_NAME
		WHERE cl.Username = ? AND cl.liked = TRUE
		GROUP BY c.name, c.descri, c.creator, r.REGION_IMG_URL, u.PHOTO_URL;
    `
	rowsLikedChats, err := db.Query(queryLikedChats, username)
	if err != nil {
		fmt.Println("Error executing liked chat query:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rowsLikedChats.Close()

	var likedChats []LikedChat // Sauvegarde les fils de discussion dans "likedChats"
	for rowsLikedChats.Next() {
		var likedChat LikedChat
		err := rowsLikedChats.Scan(&likedChat.Name, &likedChat.MessageCount, &likedChat.Description, &likedChat.PhotoURL, &likedChat.Creator)
		if err != nil {
			fmt.Println("Error scanning liked chat row:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		likedChats = append(likedChats, likedChat)
	}

	if err = rowsLikedChats.Err(); err != nil {
		fmt.Println("Error during liked chat row iteration:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	connected = ok && username != ""

	// Prepare les données pour le template
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

	t, err := template.ParseFiles("templates/profil.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute le template avec les données
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

// modifie la biographie de l'utilisateur
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	type ProfileData struct { // prend les informations fournies par l'utilisateur
		Pseudo string `json:"pseudo"`
		Bio    string `json:"bio"`
	}

	var data ProfileData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Erreur de décodage JSON", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Modifie la base de données
	_, err = db.Exec(`UPDATE User SET BIOGRAPHY = ? WHERE USERNAME = ?;`,
		data.Bio, data.Pseudo)

	// Répond avec un message de succès
	w.WriteHeader(http.StatusOK)
}

// Permet de modifier la photo de profil depuis la page profil.html
func UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	type AvatarData struct { // Prend le nouveau URL
		Avatar string `json:"avatar"`
	}

	var data AvatarData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Erreur lors du décodage du corps de la requête", http.StatusBadRequest)
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string) // prend le nom de l'utilisateur depuis la session
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	// Valide l'avatar
	if data.Avatar == "" {
		http.Error(w, "URL d'avatar non valide ou vide", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Modifie la base de données avec le bon URL
	_, err = db.Exec(`UPDATE User SET PHOTO_URL = ? WHERE USERNAME = ?;`,
		data.Avatar, username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// Déconnexion de la session
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
}

// /////////////////////////////// LIKES //////////////////////////////////////

type LikeRequest struct {
	Region string `json:"region"`
	Liked  bool   `json:"liked"`
}

// Enregistre les régions aimées par l'utilisateur
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var likeRequest LikeRequest
	err := json.NewDecoder(r.Body).Decode(&likeRequest)
	if err != nil {
		http.Error(w, "Bad request: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string) // Prend le nom d'utilisateur depuis la session
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther) // Si l'utilisateur n'est pas connecté il est renvoyé sur la page de connexion
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Regarde si l'utilisateur a déjà aimé la region
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM USER_LIKES WHERE USER_ID = ? AND REGION_NAME = ?);`
	err = db.QueryRow(query, username, likeRequest.Region).Scan(&exists)
	if err != nil {
		fmt.Println("Error checking existing like:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		// Si oui alors il met à jour la table "USER_LIKES" pour deliké la région
		_, err = db.Exec(`UPDATE USER_LIKES SET LIKED = ? WHERE USER_ID = ? AND REGION_NAME = ?;`,
			likeRequest.Liked, username, likeRequest.Region)
	} else {
		// Sinon il l'insert dans la table "USER_LIKES"
		_, err = db.Exec(`INSERT INTO USER_LIKES (USER_ID, REGION_NAME, LIKED) VALUES (?, ?, ?);`,
			username, likeRequest.Region, likeRequest.Liked)
	}

	if err != nil {
		fmt.Println("Error executing SQL query:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Region '%s' liked status updated by user %d: %t", likeRequest.Region, username, likeRequest.Liked),
	})
}

type LikeChatRequest struct {
	Chat  string `json:"region"`
	Liked bool   `json:"liked"`
}

// Enregistre les fils de discussions aimés
func LikeChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var likeChatRequest LikeChatRequest
	err := json.NewDecoder(r.Body).Decode(&likeChatRequest)
	if err != nil {
		http.Error(w, "Bad request: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Regarde si le fils discussion est déjà aimé
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM Chat_Liked WHERE Username = ? AND chatID = ?);`
	err = db.QueryRow(query, username, likeChatRequest.Chat).Scan(&exists)
	if err != nil {
		fmt.Println("Error checking existing like:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		// Si oui il met à jour la table "Chat_Liked" pour deliké le fil de discussion
		_, err = db.Exec(`UPDATE Chat_Liked SET LIKED = ? WHERE Username = ? AND chatID = ?;`,
			likeChatRequest.Liked, username, likeChatRequest.Chat)
	} else {
		// Sinon le fil de discussion est enregistré dans la table "Chat_Liked"
		_, err = db.Exec(`INSERT INTO Chat_Liked (Username, chatID, LIKED) VALUES (?, ?, ?);`,
			username, likeChatRequest.Chat, likeChatRequest.Liked)
	}

	if err != nil {
		fmt.Println("Error executing SQL query:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Chat '%s' liked status updated by user %d: %t", likeChatRequest.Chat, username, likeChatRequest.Liked),
	})
}

// Regarde si le message est présent dans la table "Msg_Liked"
func recordExists(db *sql.DB, query string, args ...interface{}) (bool, error) {
	var exists bool
	err := db.QueryRow(query, args...).Scan(&exists) // exécute la requête SQL et scanne le résultat dans "exists" avec args... qui sépare les valeurs de type ...interface{}
	return exists, err
}

// Insère ou met à jour un enregistrement dans la base de données
func insertOrUpdateRecord(db *sql.DB, insertQuery string, updateQuery string, exists bool, args []interface{}) error {
	var err error
	if exists {
		_, err = db.Exec(updateQuery, args...) // Si le message existe, exécute une mise à jour avec args... qui sépare les valeurs de type ...interface{}
	} else {
		_, err = db.Exec(insertQuery, args...) // Sinon, insère un nouvel enregistrement avec args... qui sépare les valeurs de type ...interface{}
	}
	return err
}

// Sauvegarde les messages aimés par l'utilisateur
func LikeMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Définit la structure pour les messages
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

	username, ok := session.Values["username"].(string) // récupère le nom de l'utilisateur depuis la session
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

	// Vérifie si un like existe déjà pour le message avec la fonction recordExists()
	queryExists := `SELECT EXISTS(SELECT 1 FROM Msg_Liked WHERE Username = ? AND message_id = ?);`
	exists, err := recordExists(db, queryExists, username, likeData.MessageID)
	if err != nil {
		log.Println("Error checking existing like:", err)
		http.Error(w, `{"error": "Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	if likeData.Liked {
		// Si "like", insère ou met à jour le statut "LIKED" avec la fonction insertOrUpdateRecord()
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

///////////////////////////////////////// PAGE PRINCIPALE ///////////////////////////////////////////////////////////

// Prend les informations pour la page principal
func MyTripyNonHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, connected := session.Values["username"].(string)

	// Définit la structure RegionChat
	type RegionChat struct {
		RegionName  string
		ChatCount   int
		RegionImg   string
		RegionDescr string
		RegionLiked bool
	}

	// Prepare les informations pour le template
	data := struct {
		IsConnected bool
		Regions     []RegionChat
	}{
		IsConnected: connected, // Envoi si l'utilisateur est connecté ou pas
	}

	// Récupère les informations sur les régions
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Erreur d'ouverture de la base de données.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Requête qui prend les trois régions avec le plus de chats
	query := `
    	SELECT r.REGION_NAME, 
       	COUNT(c.name) AS CHAT_COUNT, r.REGION_IMG_URL, r.DESCRI, COALESCE(ul.LIKED, FALSE) AS LIKED
		FROM Region r
		JOIN chats c ON r.REGION_NAME = c.region  -- Directly join Region with chats using the region column.
		LEFT JOIN USER_LIKES ul ON r.REGION_NAME = ul.REGION_NAME AND ul.USER_ID = ?
		GROUP BY r.REGION_NAME, r.REGION_IMG_URL, r.DESCRI
		ORDER BY CHAT_COUNT DESC
		LIMIT 3;
    `

	rows, err := db.Query(query, username) // Appel la requête avec le nom de l'utilisateur
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Erreur lors de l'exécution de la requête.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() { // Sauvegarde le retour de la requête dans "region"
		var region RegionChat
		if err := rows.Scan(&region.RegionName, &region.ChatCount, &region.RegionImg, &region.RegionDescr, &region.RegionLiked); err != nil {
			http.Error(w, "Erreur lors du scan des résultats.", http.StatusInternalServerError)
			return
		}
		data.Regions = append(data.Regions, region)
	}

	tmpl, err := template.ParseFiles("templates/mytripy-non.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Charge le template avec les informations
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

// ///////////////////////////////////////// Destinations //////////////////////////////////////////////////

// Prend toutes les destinations et enregistre le status de connexion de l'utilisateur
func AllRegions(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	username, isConnected := session.Values["username"].(string)

	// Ctructure qui enregistre les régions
	type RegionChat struct {
		RegionName  string
		RegionImg   string
		RegionDescr string
		RegionLiked bool
	}

	data := struct {
		IsConnected bool
		Regions     []RegionChat
	}{
		IsConnected: isConnected, // Envoi le status de connexion
	}

	// Ouvre la base de données
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Requête qui envoi les régions
	query := `
        SELECT Region.REGION_NAME, Region.REGION_IMG_URL, Region.DESCRI, 
        COALESCE(USER_LIKES.LIKED, FALSE) AS LIKED
        FROM Region
        LEFT JOIN USER_LIKES 
        ON Region.REGION_NAME = USER_LIKES.REGION_NAME AND USER_LIKES.USER_ID = ?;
    `
	rows, err := db.Query(query, username)
	if err != nil {
		http.Error(w, "Error querying database.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var region RegionChat // Sauvegarde le résultat de la requête dans "region"
		if err := rows.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr, &region.RegionLiked); err != nil {
			http.Error(w, "Error scanning regions.", http.StatusInternalServerError)
			return
		}
		data.Regions = append(data.Regions, region)
	}

	tmpl, err := template.ParseFiles("templates/destinations.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Envoi les data dans le template
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

// ///////////////////////////////////////////////// SEARCHBAR //////////////////////////////////////////////////

// Envoi les suggestions de la searchbar
func SearchSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("q") // prend la recherche de l'utilisateur

	// S'assure que la recherche a plus de 2 caractères sinon on renvoi un tableau vide
	if len(query) < 2 {
		json.NewEncoder(w).Encode([]map[string]string{})
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Error opening database", http.StatusInternalServerError)
		fmt.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Prend la recherche et la met sous la forme de "%query%" pour la recherche dans la base de données
	searchPattern := "%" + query + "%"

	rows, err := db.Query(`
        SELECT D.DEPARTMENT_NAME, R.REGION_NAME
        FROM Department D
        JOIN Region R ON D.REGION_NAME = R.REGION_NAME
        WHERE D.DEPARTMENT_NAME LIKE ? OR R.REGION_NAME LIKE ?
        LIMIT 5;
    `, searchPattern, searchPattern)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Query execution error:", err)
		return
	}
	defer rows.Close()

	// Prend le résultat de la requête SQL
	var filtered []map[string]string
	for rows.Next() {
		var departmentName, regionName string
		if err := rows.Scan(&departmentName, &regionName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Row scan error:", err)
			return
		}
		// Prend la région et le département
		filtered = append(filtered, map[string]string{
			"departmentName": departmentName,
			"regionName":     regionName,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Rows iteration error:", err)
		return
	}

	// Envoi les suggestions filtrées
	json.NewEncoder(w).Encode(filtered)
}

// ///////////////////////////////////////////////// FIL DISCUSSION //////////////////////////////////////////////////

// Prend les fils de discussions filtrées par rapport à la région choisie
func FileDiscussion(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
	}

	username, _ := session.Values["username"].(string) // prend le nom d'utilisateur depuis la session
	region, ok := session.Values["region"].(string)
	if !ok || region == "" {
		http.Redirect(w, r, "/region-selection", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// prend le chat principal, son status de like et celui de la connection de l'utilisateur
	queryMain := `
        SELECT 
    c.name AS chat_name,
    COUNT(m.id) AS message_count,
    c.descri AS chat_description,
    r.REGION_IMG_URL AS region_image,
    COALESCE(like_counts.total_likes, 0) AS total_likes,
    COALESCE(cl.liked, FALSE) AS user_liked
FROM 
    chats c
LEFT JOIN 
    messages m ON c.name = m.chat_name
LEFT JOIN 
    Region r ON c.region = r.REGION_NAME
LEFT JOIN (
    SELECT 
        chatID, 
        COUNT(*) AS total_likes
    FROM 
        Chat_Liked
    WHERE 
        liked = 1
    GROUP BY 
        chatID
) like_counts ON c.name = like_counts.chatID
LEFT JOIN 
    Chat_Liked cl ON c.name = cl.chatID AND cl.Username = ?
WHERE 
    c.principal = TRUE
    AND c.region = ?
GROUP BY 
    c.name, c.descri, r.REGION_IMG_URL, like_counts.total_likes, cl.liked;
    `
	principal, err := db.Query(queryMain, username, region)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer principal.Close()

	type MainChat struct {
		Name         string
		MessageCount int
		Descri       string
		ImageURL     string
		TotalLikes   int
		UserLiked    bool
	}

	var mainChat MainChat // sauvegarde le résultat de la requête dans "mainChat"
	if principal.Next() {
		err := principal.Scan(&mainChat.Name, &mainChat.MessageCount, &mainChat.Descri, &mainChat.ImageURL, &mainChat.TotalLikes, &mainChat.UserLiked)
		if err != nil {
			http.Error(w, "Failed to scan main chat data", http.StatusInternalServerError)
			return
		}
	}

	// prend des chats utilisateur, leur status de like et celui de la connection de l'utilisateur
	queryChats := `
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
    SELECT 
        chatID, 
        COUNT(*) AS total_likes
    FROM 
        Chat_Liked
    WHERE 
        liked = 1
    GROUP BY 
        chatID
) like_counts ON c.name = like_counts.chatID
LEFT JOIN 
    Chat_Liked cl ON c.name = cl.chatID AND cl.Username = ?
WHERE 
    c.principal = FALSE 
    AND c.region = ?
GROUP BY 
    c.name, c.descri, u.PHOTO_URL, u.USERNAME, like_counts.total_likes, cl.liked;
    `
	rows, err := db.Query(queryChats, username, region)
	if err != nil {
		http.Error(w, "Server error while fetching chats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type UserChat struct {
		Name         string
		MessageCount int
		Descri       string
		PhotoURL     string
		Creator      string
		TotalLikes   int
		UserLiked    bool
	}

	var chats []UserChat // sauvegarde le résultat de la requête dans "chats"
	for rows.Next() {
		var chat UserChat
		if err := rows.Scan(&chat.Name, &chat.MessageCount, &chat.Descri, &chat.PhotoURL, &chat.Creator, &chat.TotalLikes, &chat.UserLiked); err != nil {
			log.Println("Error scanning chat data:", err)
			continue
		}
		chats = append(chats, chat)
	}

	// Structure final qui va être envoyé à la page
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

	// Envoi les informations à la page de fils de discussions
	if err := template.Must(template.ParseFiles("templates/welcome.html")).Execute(w, data); err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// Permet d'enregistrer un chat dans la base de données lorsqu'il a été créé par l'utilisateur
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

	creator, ok := session.Values["username"].(string)
	if !ok || creator == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	chatName := r.FormValue("chatname")           // récupère le nom du chat
	chatDescription := r.FormValue("description") // récupère la description
	region := r.FormValue("region")               // récupère la région
	if chatName == "" || region == "" {
		http.Error(w, "Chat name or region missing", http.StatusBadRequest)
		log.Println("Chat creation failed: missing chat name or region.")
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// Enregistre le chat et ses informations dans la table "chats"
	_, err = db.Exec("INSERT INTO chats (name, creator, region, descri, principal) VALUES (?, ?, ?, ?,?)", chatName, creator, region, chatDescription, 0)
	if err != nil {
		log.Printf("Chat creation failed: %v", err)
		http.Error(w, "Chat creation failed", http.StatusInternalServerError)
		return
	}

	// redirige l'utilisateur vers la page de fils de discussions
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}

// Lorsque l'utilisateur choisi un chat, son nom est enregistré dans la session
func SelectChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chatName := r.URL.Query().Get("chatname") // prend le nom du chat
	if chatName == "" {
		http.Error(w, "Chat name missing", http.StatusBadRequest)
		log.Println("Chat selection failed: missing chat name in request.")
		return
	}

	session, err := Store.Get(r, "session-name") // récupère la session
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	session.Values["chatname"] = chatName // enregistre le nom du fil de discussion dans la session
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ///////////////////////////////////////////// MESSAGES ///////////////////////////////////////////

// Récupère les messages de la base de données et les envoi au template
func FilMessagesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	username, usernameExists := session.Values["username"].(string)
	chatName, chatNameExists := session.Values["chatname"].(string)
	if !usernameExists || username == "" || !chatNameExists || chatName == "" {
		// Redirige l'utilisateur vers "/connexion" si l'utilisateur n'est pas connecté
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prend tous les messages et leur status de like
	rows, err := db.Query(
		`SELECT 
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
            m.timestamp ASC;`,
		username, chatName,
	)
	if err != nil {
		log.Println("Error fetching messages:", err)
		http.Error(w, "Error retrieving messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []struct {
		MessageID     int
		Sender        string
		Message       string
		TimeElapsed   string
		ImgUser       string
		NumberOfLikes int
		UserLiked     bool
	}

	for rows.Next() {
		var messageID int
		var sender, message, timestamp, imgUser string
		var numberOfLikes int
		var userLiked bool

		if err := rows.Scan(&messageID, &sender, &message, &timestamp, &imgUser, &numberOfLikes, &userLiked); err != nil {
			log.Println("Error scanning message row:", err)
			continue
		}

		// Calcule le temps d'écart entre le temps actuel et celui d'envoi du message
		elapsedTime, err := formatElapsedTime(timestamp)
		if err != nil {
			log.Println("Error parsing timestamp:", err)
			continue
		}

		messages = append(messages, struct { // enregistre le tout dans messages
			MessageID     int
			Sender        string
			Message       string
			TimeElapsed   string
			ImgUser       string
			NumberOfLikes int
			UserLiked     bool
		}{
			MessageID:     messageID,
			Sender:        sender,
			Message:       message,
			TimeElapsed:   elapsedTime,
			ImgUser:       imgUser,
			NumberOfLikes: numberOfLikes,
			UserLiked:     userLiked,
		})
	}

	data := struct { // Structure finale
		ChatName string
		Messages []struct {
			MessageID     int
			Sender        string
			Message       string
			TimeElapsed   string
			ImgUser       string
			NumberOfLikes int
			UserLiked     bool
		}
	}{
		ChatName: chatName,
		Messages: messages,
	}

	tmpl := template.Must(template.ParseFiles("templates/chat_messages.html"))
	err = tmpl.Execute(w, data) // Envoi les informations à la page
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// Les messages envoyé sont enregistrés dans la base de données
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	session, err := Store.Get(r, "session-name") // Récupère la session
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	chatName, ok := session.Values["chatname"].(string) // Récupère le nom du chat depuis la session
	if !ok || chatName == "" {
		http.Error(w, "Chat not selected. Please select a chat.", http.StatusBadRequest)
		log.Println("Chatname not found in session.")
		return
	}

	username, ok := session.Values["username"].(string) // Récupère le nom d'utilisateur depuis la session
	if !ok || username == "" {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	message := r.FormValue("message") // Récupère le message
	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		log.Println("Empty message received.")
		return
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05") // Récupère le temps actuelle

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	_, err = db.Exec( // Enregistre le message dans la table "messages"
		"INSERT INTO messages (chat_name, sender, message, timestamp) VALUES (?, ?, ?, ?)",
		chatName, username, message, currentTime,
	)
	if err != nil {
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		log.Println("Error saving message to database:", err)
		return
	}

}

// Sert à récupérer les messages dynamiquement
func FetchMessagesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name") // récupère la session
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	// récupère le nom d'utilisateur et de chat
	username, usernameExists := session.Values["username"].(string)
	chatName, chatNameExists := session.Values["chatname"].(string)
	if !usernameExists || username == "" || !chatNameExists || chatName == "" {
		//redirige vers "/SeConnecter" si l'utilisateur n'est pas connecté
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prend les messages pour un fil de discussions spécifique
	rows, err := db.Query(
		`SELECT 
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
            m.timestamp ASC;`,
		username, chatName,
	)
	if err != nil {
		log.Println("Error fetching messages:", err)
		http.Error(w, "Server error while fetching messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []struct {
		MessageID     int    `json:"message_id"`
		Sender        string `json:"sender"`
		Message       string `json:"message"`
		TimeElapsed   string `json:"time_elapsed"`
		ImgUser       string `json:"img_user"`
		NumberOfLikes int    `json:"number_of_likes"`
		UserLiked     bool   `json:"user_liked"`
	}

	for rows.Next() {
		var messageID int
		var sender, message, timestamp, imgUser string
		var numberOfLikes int
		var userLiked bool

		if err := rows.Scan(&messageID, &sender, &message, &timestamp, &imgUser, &numberOfLikes, &userLiked); err != nil {
			log.Println("Error scanning message data:", err)
			continue
		}

		// récupère l'écart entre l'envoi du message et le temps actuel
		elapsedTime, err := formatElapsedTime(timestamp)
		if err != nil {
			log.Println("Error formatting timestamp:", err)
			continue
		}

		messages = append(messages, struct {
			MessageID     int    `json:"message_id"`
			Sender        string `json:"sender"`
			Message       string `json:"message"`
			TimeElapsed   string `json:"time_elapsed"`
			ImgUser       string `json:"img_user"`
			NumberOfLikes int    `json:"number_of_likes"`
			UserLiked     bool   `json:"user_liked"`
		}{
			MessageID:     messageID,
			Sender:        sender,
			Message:       message,
			TimeElapsed:   elapsedTime,
			ImgUser:       imgUser,
			NumberOfLikes: numberOfLikes,
			UserLiked:     userLiked,
		})
	}

	// renvoi un message JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// Envoi l'écart entre le temps d'envoi du message et le temps actuel
func formatElapsedTime(input string) (string, error) {
	layout := "2006-01-02 15:04:05"
	inputTime, err := time.Parse(layout, input)
	if err != nil {
		return "", err
	}

	// récupère le temps actuel
	currentTime := time.Now()

	// Calcule la différence
	duration := currentTime.Sub(inputTime)

	// Gérer les différents cas
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
	// Si l'écart est plus petit que des heures alors on renvoi l'heure et les minutes d'envoi du message
	return timeStr, nil
}

// //////////////////////////////////////// REGION HANDLER ///////////////////////////////////////////////

// Sauvegarde la région où se trouve l'utilisateur
func RegionHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("name") // récupère le nom de la région
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

	session.Values["region"] = region // enregisre la région dans la session
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirige l'utilisateur vers la page de fils de discussion
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}

// /////////////////////////// OUBLIE DE MOT DE PASSE //////////////////////////////////////

// Permet à l'utilisateur de changer le mot de passe en cas d'oublie
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupère ces iformations depuis le formulaire du template
	email := r.FormValue("email")
	pseudo := r.FormValue("username")
	motDePasse := r.FormValue("password")
	confirmeMotDePasse := r.FormValue("confirmPassword")

	if motDePasse != confirmeMotDePasse { // s'assure que les mot de passe correspondent
		renderError(w, "mot-de-passe-oublie", "Les mots de passe ne correspondent pas.")
		return
	}

	if !isValidPassword(motDePasse) { // si le mot de passe respecte les contraintes de sécurité
		renderError(w, "mot-de-passe-oublie", "Le mot de passe doit contenir au minimum\nune majuscule, une minuscule, un caractère spécial, un chiffre, et au minimum 6 caractères.")
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "mot-de-passe-oublie", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	emailExists, pseudoExists, err := CheckUserExists(db, email, pseudo) // si le mail et le pseudo existent dans la base de données
	if err != nil {
		renderError(w, "mot-de-passe-oublie", "Erreur lors de la vérification des utilisateurs existants.")
		return
	}
	if !emailExists {
		renderError(w, "mot-de-passe-oublie", "L'email n'existe pas.")
		return
	}
	if !pseudoExists {
		renderError(w, "mot-de-passe-oublie", "Le pseudo n'existe pas.")
		return
	}

	// s'assure que le mail et le pseudo correspondent
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM User WHERE EMAIL = ? AND USERNAME = ?", email, pseudo).Scan(&count)
	if err != nil {
		renderError(w, "mot-de-passe-oublie", "Erreur lors de la vérification des identifiants.")
		return
	}
	if count == 0 {
		renderError(w, "mot-de-passe-oublie", "Les identifiants ne sont pas compatibles.")
		return
	}

	// chiffre le mot de passe
	motDePasseChiffre, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, "mot-de-passe-oublie", "Erreur lors du chiffrement du mot de passe.")
		return
	}

	// modifi le mot de passe dans la table "User"
	_, err = db.Exec("UPDATE User SET PASSWORD = ? WHERE EMAIL = ? AND USERNAME = ?", motDePasseChiffre, email, pseudo)
	if err != nil {
		renderError(w, "mot-de-passe-oublie", "Erreur lors de la création du compte.")
		return
	}

	// Rediriger vers la page mytripy-non après la création du compte
	http.Redirect(w, r, "/SeConnecter", http.StatusFound)
}
