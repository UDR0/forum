package main

import (
	"fmt"
	forum "forum/Functions"
	"html/template"
	"net/http"
)

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
		http.Redirect(w, r, "/mytripy-non", http.StatusFound)
	})

	// Routes pour les pages HTML
	http.HandleFunc("/mytripy-non", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "mytripy-non")
	})

	http.HandleFunc("/SeConnecter", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			forum.CheckCredentialsForConnection(w, r)
		} else {
			renderTemplate(w, "SeConnecter")
		}
	})

	/*http.HandleFunc("/CreerCompte", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "CreerCompte")
	})*/

	http.HandleFunc("/CreerCompte", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			forum.CreateUser(w, r)
		} else {
			renderTemplate(w, "CreerCompte")
		}
	})

	http.HandleFunc("/mot-de-passe-oublie", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "mot-de-passe-oublie")
	})

	http.HandleFunc("/profil", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "profil")
	})

	// Démarrer le serveur
	fmt.Println("Serveur lancé sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}
