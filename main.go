// file con l'immagine inizializzata direttamente all'inizio del file
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"

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

	// Print the retrieved image URL to the console
	//fmt.Printf("Image URL retrieved from database: %s\n", imageURL)

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

func openBrowser(url string) {
	var err error
	switch os := runtime.GOOS; os {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}

func main() {
	// Serve static files from the "static" folder.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Route to serve the template.
	http.HandleFunc("/", forumHandler)

	// Lancer le serveur sur le port 8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	go openBrowser("http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
