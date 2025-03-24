package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func getImageURLFromDB() string {
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

func renderTemplate(w http.ResponseWriter, tmpl string) {
	tmplPath := fmt.Sprintf("templates/%s.html", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Servir les fichiers statiques (CSS, JS, images)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route pour la page d'accueil
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mytripy_non", http.StatusFound)
	})

	// Routes pour les pages HTML
	http.HandleFunc("/mytripy_non", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "mytripy_non")
	})

	http.HandleFunc("/SeConnecter", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "SeConnecter")
	})

	// Démarrer le serveur
	fmt.Println("Serveur lancé sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}
