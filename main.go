package main

import (
	"database/sql"
	"fmt"

	// import functions from forum.go
	//forum "forum/Functions"

	"html/template"
	"log"
	"net/http"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//C:\Users\sarah\Documents\Forum\forum\forum.db
//C:\Users\sarah\Documents\Forum\forum\ProjectForum\main.go

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

func main() {
	//test to have acces to the functions from the forum.go file
	//forum.SayHello()

	// Serve static files from the "static" folder.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Route to serve the template.
	http.HandleFunc("/", forumHandler)

	// Launch the server on port 8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Erreur lors du lancement du serveur : ", err)
	}
}
