package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
)

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
	// Répertoire contenant les fichiers statiques
	fs := http.FileServer(http.Dir("./static"))

	// Serveur des fichiers statiques (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route pour afficher la page d'accueil
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Charger le template HTML depuis le répertoire "template"
		tmpl, err := template.ParseFiles("./templates/region.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
			return
		}

		// Exécuter le template et l'envoyer au client
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		}
	})

	// Lancer le serveur sur le port 8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	go openBrowser("http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}

}
