package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// Ouvre la base de données SQLite
func openDB() (*sql.DB, error) {
	// Ouvre la base de données dans le répertoire /view/forum.db
	db, err := sql.Open("sqlite3", "./view/forum.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Route pour afficher l'image d'une région
func regionHandler(w http.ResponseWriter, r *http.Request) {
	// Ouvre la base de données
	db, err := openDB()
	if err != nil {
		http.Error(w, "Erreur lors de l'ouverture de la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Récupère l'URL de l'image pour la région (par exemple id_region = 1)
	var imageURL string
	err = db.QueryRow("SELECT image_url FROM regions WHERE id_region = 1").Scan(&imageURL)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'image de la région", http.StatusInternalServerError)
		return
	}

	// Charge le template HTML
	tmpl, err := template.ParseFiles("./templates/region.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	// Exécute le template avec l'URL de l'image comme donnée
	err = tmpl.Execute(w, imageURL)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Serveur des fichiers statiques pour les images et autres ressources (CSS, JS)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route pour afficher la page de la région
	http.HandleFunc("/region", regionHandler)

	// Lancer le serveur sur le port 8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Erreur lors du lancement du serveur : ", err)
	}
}
