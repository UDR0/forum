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

// Regarde si l'utilisateur existe (ce l'ho già nell'altra pero non è uguale)
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

// guarda se il password dato va bene (con numer, maiuscule, minuscule, caratteri speciali)
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

// prende info date dalla creazione del account e le mette nella database (ce l'ho già ma non è uguale)
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

	// Créer une nouvelle session et stocker le nom d'utilisateur
	session, _ := Store.Get(r, "session-name")
	session.Values["username"] = pseudo
	session.Save(r, w)

	// Rediriger vers la page mytripy-non après la création du compte
	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
}

// guarda che le credentials sono giuste per connessione (c'è ma non è uguale)
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

// prende le info dal sql per mettere le info dello user nella pagina profilo
func ProfilPage(w http.ResponseWriter, r *http.Request) {
	type Region struct {
		RegionName  string `json:"region_name"`
		RegionImg   string `json:"region_imgurl"`
		RegionDescr string `json:"region_description"`
	}
	var connected bool
	session, _ := Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)

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

	var pseudo, urlPhoto, biography string
	err = db.QueryRow("SELECT USERNAME, PHOTO_URL, BIOGRAPHY FROM User WHERE USERNAME = ?", username).Scan(&pseudo, &urlPhoto, &biography)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `
        SELECT Region.REGION_NAME, Region.REGION_IMG_URL, Region.DESCRI
		FROM Region
		JOIN USER_LIKES ON Region.REGION_NAME = USER_LIKES.REGION_NAME
		WHERE USER_LIKES.USER_ID = ? AND USER_LIKES.LIKED = TRUE;
    `
	rows, err := db.Query(query, username)
	if err != nil {
		fmt.Println("Error executing query:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Slice to hold the regions
	var regions []Region

	for rows.Next() {
		var region Region
		err := rows.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		regions = append(regions, region)
	}

	// Check for errors after iteration
	if err = rows.Err(); err != nil {
		fmt.Println("Error during row iteration:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	username, isConnected := session.Values["username"].(string)
	if isConnected {
		connected = true
	} else {
		connected = false
	}

	data := struct {
		Pseudo      string
		PhotoURL    string
		Biography   string
		IsConnected bool
		Regions     []Region
	}{
		Pseudo:      pseudo,
		PhotoURL:    urlPhoto,
		Biography:   biography,
		IsConnected: connected,
		Regions:     regions,
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

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	type ProfileData struct {
		Pseudo string `json:"pseudo"`
		Bio    string `json:"bio"`
	}

	var data ProfileData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Erreur de décodage JSON", http.StatusBadRequest)
		return
	}

	// Simulez la mise à jour (par exemple, en sauvegardant dans une base de données)
	// Exemple : updateDatabase(data.Pseudo, data.Bio)

	// Répondre avec un message de succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Profil mis à jour avec succès !"))
}

// deconessione dalla sessione, questo c'è ma non è uguale
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/mytripy-non", http.StatusFound)
}

// ///////////////////////////////////////////////////////////////////////////////////////////:
type LikeRequest struct {
	Region string `json:"region"` // Region name from the client
	Liked  bool   `json:"liked"`  // Liked status from the client
}

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

	// Check if the user already liked this region
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM USER_LIKES WHERE USER_ID = ? AND REGION_NAME = ?);`
	err = db.QueryRow(query, username, likeRequest.Region).Scan(&exists)
	if err != nil {
		fmt.Println("Error checking existing like:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		// Update the existing like record
		_, err = db.Exec(`UPDATE USER_LIKES SET LIKED = ? WHERE USER_ID = ? AND REGION_NAME = ?;`,
			likeRequest.Liked, username, likeRequest.Region)
	} else {
		// Insert a new like record
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

////////////////////////////////////////////////////////////////////////////////////////////////////

// prende info di cui ha bisogno mytripy-non.html, da ridurre se possibile
func MyTripyNonHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session-name")
	_, connected := session.Values["username"].(string)

	type RegionChat struct {
		RegionName  string
		ChatCount   int
		RegionImg   string
		RegionDescr string
	}

	data := struct {
		IsConnected bool
		Regions     []RegionChat
	}{
		IsConnected: connected,
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

// prende tutte le regioni di cui ha bisognio la pagina destinations.html (troppo simile a quella dentro MyTripyNonHandler)
func AllRegions(w http.ResponseWriter, r *http.Request) {
	var connected bool
	session, _ := Store.Get(r, "session-name")
	_, isConnected := session.Values["username"].(string)
	connected = isConnected

	type RegionChat struct {
		RegionName  string
		RegionImg   string
		RegionDescr string
	}

	data := struct {
		IsConnected bool
		Regions     []RegionChat
	}{
		IsConnected: connected,
	}

	// Fetch regions from the database
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
        SELECT REGION_NAME, REGION_IMG_URL, DESCRI 
        FROM Region;
    `
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Error querying database.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var region RegionChat
		if err := rows.Scan(&region.RegionName, &region.RegionImg, &region.RegionDescr); err != nil {
			http.Error(w, "Error scanning regions.", http.StatusInternalServerError)
			return
		}
		data.Regions = append(data.Regions, region)
	}

	// Render the template
	tmpl, err := template.ParseFiles("templates/destinations.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func SearchSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("q")

	// Check if query length is at least 2 characters
	if len(query) < 2 {
		json.NewEncoder(w).Encode([]map[string]string{}) // Return an empty array if query is too short
		return
	}

	// Database connection (adjust with your own credentials)
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		http.Error(w, "Error opening database", http.StatusInternalServerError)
		fmt.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Append wildcards for the LIKE clause
	searchPattern := "%" + query + "%"

	// Prepare the SQL query
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

	// Process the query results
	var filtered []map[string]string
	for rows.Next() {
		var departmentName, regionName string
		if err := rows.Scan(&departmentName, &regionName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Row scan error:", err)
			return
		}
		// Add department and region name as separate fields in the response
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

	// Send the filtered options as JSON response
	json.NewEncoder(w).Encode(filtered)
}

// ///////////////////////////////////////////////// FIL DISCUSSION //////////////////////////////////////////////////
// prende tutte le chat dalla db
func FileDiscussion(w http.ResponseWriter, r *http.Request) {
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

	region, ok := session.Values["region"].(string)
	if !ok || region == "" {
		http.Redirect(w, r, "/region-selection", http.StatusSeeOther)
		return
	}

	// Fetch chats for the region...

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// Fetch main chat
	queryMain := `SELECT 
                      c.name AS chat_name, 
                      COUNT(m.id) AS message_count, 
                      c.descri, 
                      r.REGION_IMG_URL
                  FROM 
                      chats c
                  LEFT JOIN 
                      messages m ON c.name = m.chat_name
                  LEFT JOIN 
                      Region r ON c.region = r.REGION_NAME
                  WHERE 
                      c.principal = TRUE 
                      AND c.region = ?
                  GROUP BY 
                      c.name, c.descri, r.REGION_IMG_URL`
	principal, err := db.Query(queryMain, region)
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
	}

	var mainChat MainChat
	if principal.Next() {
		err := principal.Scan(&mainChat.Name, &mainChat.MessageCount, &mainChat.Descri, &mainChat.ImageURL)
		if err != nil {
			http.Error(w, "Failed to scan main chat data", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No main chat found", http.StatusNotFound)
		return
	}

	// Fetch user chats
	queryChats := `SELECT 
                       c.name, 
                       COUNT(m.id) AS message_count, 
                       c.descri, 
                       u.PHOTO_URL, 
                       u.USERNAME
                   FROM 
                       chats c
                   LEFT JOIN 
                       messages m ON c.name = m.chat_name
                   LEFT JOIN 
                       User u ON c.creator = u.USERNAME
                   WHERE 
                       c.principal = FALSE 
                       AND c.region = ?
                   GROUP BY 
                       c.name, c.descri, u.PHOTO_URL, u.USERNAME`
	rows, err := db.Query(queryChats, region)
	if err != nil {
		http.Error(w, "Server error while fetching chats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var chats []struct {
		Name         string
		MessageCount int
		Descri       string
		PhotoURL     string
		Creator      string
	}

	for rows.Next() {
		var chat struct {
			Name         string
			MessageCount int
			Descri       string
			PhotoURL     string
			Creator      string
		}
		if err := rows.Scan(&chat.Name, &chat.MessageCount, &chat.Descri, &chat.PhotoURL, &chat.Creator); err != nil {
			log.Println("Error scanning chat data:", err)
			continue
		}
		chats = append(chats, chat)
	}

	// Final data struct
	data := struct {
		Username string
		Region   string
		MainChat MainChat
		Chats    []struct {
			Name         string
			MessageCount int
			Descri       string
			PhotoURL     string
			Creator      string
		}
	}{
		Username: username,
		Region:   region,
		MainChat: mainChat,
		Chats:    chats,
	}

	if err := template.Must(template.ParseFiles("templates/welcome.html")).Execute(w, data); err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// chiamato quando vuoi creare un chat, e salva il nome del chat con nome del creatore, nome del chat e in che regione si trova, ti ridirige poi verso /welcome
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

	chatName := r.FormValue("chatname")
	chatDescription := r.FormValue("description") // Retrieve the description
	region := r.FormValue("region")
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

	// Adjust the SQL to include the description field
	_, err = db.Exec("INSERT INTO chats (name, creator, region, descri, principal) VALUES (?, ?, ?, ?,?)", chatName, creator, region, chatDescription, "FALSE")
	if err != nil {
		log.Printf("Chat creation failed: %v", err)
		http.Error(w, "Chat creation failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Chat created successfully: %s in region %s by %s with description: %s", chatName, region, creator, chatDescription)
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}

// chiamato quando scegli un chat, salva il nome del chat dove vuoi andare e ridirige verso /chat_messages
func SelectChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chatName := r.URL.Query().Get("chatname") // Get chatname from query parameter
	if chatName == "" {
		http.Error(w, "Chat name missing", http.StatusBadRequest)
		log.Println("Chat selection failed: missing chat name in request.")
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	session.Values["chatname"] = chatName
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Printf("Chatname stored in session: %s", chatName)
	w.WriteHeader(http.StatusOK)
}

// prende tutte le chat di una certa regione
func FetchChatsHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if region == "" {
		http.Error(w, "Region is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	// Fetch main chat
	queryMain := `SELECT 
                      c.name AS chat_name, 
                      COUNT(m.id) AS message_count, 
                      c.descri, 
                      r.REGION_IMG_URL
                  FROM 
                      chats c
                  LEFT JOIN 
                      messages m ON c.name = m.chat_name
                  LEFT JOIN 
                      Region r ON c.region = r.REGION_NAME
                  WHERE 
                      c.principal = TRUE 
                      AND c.region = ?
                  GROUP BY 
                      c.name, c.descri, r.REGION_IMG_URL`
	principal, err := db.Query(queryMain, region)
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
	}

	var mainChat MainChat
	if principal.Next() {
		err := principal.Scan(&mainChat.Name, &mainChat.MessageCount, &mainChat.Descri, &mainChat.ImageURL)
		if err != nil {
			http.Error(w, "Failed to scan main chat data", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No main chat found", http.StatusNotFound)
		return
	}

	// Fetch user chats
	queryChats := `SELECT 
                       c.name, 
                       COUNT(m.id) AS message_count, 
                       c.descri, 
                       u.PHOTO_URL, 
                       u.USERNAME
                   FROM 
                       chats c
                   LEFT JOIN 
                       messages m ON c.name = m.chat_name
                   LEFT JOIN 
                       User u ON c.creator = u.USERNAME
                   WHERE 
                       c.principal = FALSE 
                       AND c.region = ?
                   GROUP BY 
                       c.name, c.descri, u.PHOTO_URL, u.USERNAME`
	rows, err := db.Query(queryChats, region)
	if err != nil {
		http.Error(w, "Server error while fetching chats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var chats []struct {
		Name         string
		MessageCount int
		Descri       string
		PhotoURL     string
		Creator      string
	}

	for rows.Next() {
		var chat struct {
			Name         string
			MessageCount int
			Descri       string
			PhotoURL     string
			Creator      string
		}
		if err := rows.Scan(&chat.Name, &chat.MessageCount, &chat.Descri, &chat.PhotoURL, &chat.Creator); err != nil {
			log.Println("Error scanning chat data:", err)
			continue
		}
		chats = append(chats, chat)
	}

	// Final data struct
	data := struct {
		MainChat MainChat
		Chats    []struct {
			Name         string
			MessageCount int
			Descri       string
			PhotoURL     string
			Creator      string
		}
	}{
		MainChat: mainChat,
		Chats:    chats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding data", http.StatusInternalServerError)
	}
}

// ///////////////////////////////////////////// MESSAGES ///////////////////////////////////////////
// prende i messaggi dall db per darli alla pagina web
func FilMessagesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	chatName, ok := session.Values["chatname"].(string)
	if !ok || chatName == "" {
		http.Error(w, "Chat not selected. Please go back and select a chat.", http.StatusBadRequest)
		log.Println("Chatname not found in session.")
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/SeConnecter", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT m.sender, m.message, strftime('%Y-%m-%d %H:%M:%S', m.timestamp), u.PHOTO_URL FROM messages m LEFT JOIN User u ON m.sender = u.USERNAME WHERE m.chat_name = ? ORDER BY m.timestamp ASC;",
		chatName,
	)
	if err != nil {
		http.Error(w, "Server error while fetching messages", http.StatusInternalServerError)
		log.Println("Error fetching messages:", err)
		return
	}
	defer rows.Close()

	var messages []struct {
		Sender      string
		Message     string
		TimeElapsed string // Contains calculated elapsed time
		ImgUser     string
	}
	for rows.Next() {
		var sender, message, timestamp, imgUser string
		if err := rows.Scan(&sender, &message, &timestamp, &imgUser); err != nil {
			log.Println("Error scanning message data:", err)
			continue
		}

		// Calculate elapsed time
		elapsedTime, err := formatElapsedTime(timestamp)
		if err != nil {
			log.Println("Error parsing timestamp:", err)
			continue
		}

		messages = append(messages, struct {
			Sender      string
			Message     string
			TimeElapsed string
			ImgUser     string
		}{
			Sender:      sender,
			Message:     message,
			TimeElapsed: elapsedTime,
			ImgUser:     imgUser,
		})
	}

	data := struct {
		ChatName string
		Username string
		Messages []struct {
			Sender      string
			Message     string
			TimeElapsed string
			ImgUser     string
		}
	}{
		ChatName: chatName,
		Username: username,
		Messages: messages,
	}
	log.Printf("Messages passed to template: %+v", messages)

	err = template.Must(template.ParseFiles("templates/chat_messages.html")).Execute(w, data)
	if err != nil {
		log.Println("Error rendering chat messages template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// una volta mandato, il messaggio viene salvato nella db
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	chatName, ok := session.Values["chatname"].(string)
	if !ok || chatName == "" {
		http.Error(w, "Chat not selected. Please select a chat.", http.StatusBadRequest)
		log.Println("Chatname not found in session.")
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		log.Println("Empty message received.")
		return
	}

	// Save current timestamp with full date and time
	currentTime := time.Now().Format("2006-01-02 15:04:05") // Full timestamp format

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO messages (chat_name, sender, message, timestamp) VALUES (?, ?, ?, ?)",
		chatName, username, message, currentTime,
	)
	if err != nil {
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		log.Println("Error saving message to database:", err)
		return
	}

	log.Printf("Message saved successfully: %s - %s: %s", chatName, username, message)
}

// prende tutti il messaggi di una chat specifica per poi fare apparirli nella pagina
func FetchMessagesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	chatName, ok := session.Values["chatname"].(string)
	if !ok || chatName == "" {
		http.Error(w, "Chat not selected. Please select a chat.", http.StatusBadRequest)
		log.Println("Chatname not found in session.")
		return
	}

	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		renderError(w, "CreerCompte", "Erreur d'ouverture de la base de données.")
		return
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT m.sender, m.message, strftime('%Y-%m-%d %H:%M:%S', m.timestamp), u.PHOTO_URL FROM messages m LEFT JOIN User u ON m.sender = u.USERNAME WHERE m.chat_name = ? ORDER BY m.timestamp ASC;",
		chatName,
	)
	if err != nil {
		http.Error(w, "Server error while fetching messages", http.StatusInternalServerError)
		log.Println("Error fetching messages:", err)
		return
	}
	defer rows.Close()

	var messages []struct {
		Sender      string
		Message     string
		TimeElapsed string
		ImgUser     string
	}
	for rows.Next() {
		var sender, message, timestamp, imgUser string
		if err := rows.Scan(&sender, &message, &timestamp, &imgUser); err != nil {
			log.Println("Error scanning message data:", err)
			continue
		}

		// Calculate elapsed time
		elapsedTime, err := formatElapsedTime(timestamp)
		if err != nil {
			log.Println("Error parsing timestamp:", err)
			continue
		}

		messages = append(messages, struct {
			Sender      string
			Message     string
			TimeElapsed string
			ImgUser     string
		}{
			Sender:      sender,
			Message:     message,
			TimeElapsed: elapsedTime,
			ImgUser:     imgUser,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		log.Println("Error encoding messages to JSON:", err)
		http.Error(w, "Error encoding messages", http.StatusInternalServerError)
	}
}

func formatElapsedTime(input string) (string, error) {
	// Parse the input time string
	layout := "2006-01-02 15:04:05"
	inputTime, err := time.Parse(layout, input)
	if err != nil {
		return "", err
	}

	// Get the current time
	currentTime := time.Now()

	// Calculate the difference
	duration := currentTime.Sub(inputTime)

	// Handle specific cases
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
	// Default case if none of the above applies
	return timeStr, nil
}

// //////////////////////////////////////// REGION HANDLER ///////////////////////////////////////////////
// salva il nome della region in cui si trova lo user
func RegionHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("name")
	if region == "" {
		http.Error(w, "Region not selected. Please choose a region.", http.StatusBadRequest)
		log.Println("No region selected.")
		return
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error retrieving session:", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	session.Values["region"] = region
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}
