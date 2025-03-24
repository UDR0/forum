package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {
	// Répertoire contenant les fichiers statiques
	fs := http.FileServer(http.Dir("./static"))

	// Serveur des fichiers statiques (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route pour afficher la page d'accueil
	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		// Charger le template HTML depuis le répertoire "template"
		tmpl, err := template.ParseFiles("./templates/index.html")
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

	// Route pour afficher la page se connecter
	http.HandleFunc("/SeConnecter", func(w http.ResponseWriter, r *http.Request) {
		// Charger le template HTML depuis le répertoire "template"
		tmpl, err := template.ParseFiles("templates/SeConnecter.html")
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
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}
