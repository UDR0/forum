package main

import (
	"fmt"
	forum "forum/Functions"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"  // go get github.com/gorilla/sessions
	"github.com/gorilla/websocket" // go get github.com/gorilla/websocket
)

var (
	store    = sessions.NewCookieStore([]byte("something-very-secret"))
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := fmt.Sprintf("templates/%s.html", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Servir les fichiers statiques (CSS, JS, images)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route pour la page d'accueil
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/mytripy-non", http.StatusFound)
	})

	// Routes pour les pages HTML
	http.HandleFunc("/mytripy-non", forum.MyTripyNonHandler)

	http.HandleFunc("/SeConnecter", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			forum.CheckCredentialsForConnection(w, r)
			session, _ := store.Get(r, "session")
			session.Values["user"] = r.FormValue("username")
			session.Save(r, w)
			http.Redirect(w, r, "/profil", http.StatusFound)
		} else {
			renderTemplate(w, "SeConnecter", nil)
		}
	})

	http.HandleFunc("/CreerCompte", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			forum.CreateUser(w, r)
		} else {
			renderTemplate(w, "CreerCompte", nil)
		}
	})

	http.HandleFunc("/mot-de-passe-oublie", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "mot-de-passe-oublie", nil)
	})

	http.HandleFunc("/profil", forum.ProfilPage)

	http.HandleFunc("/destinations", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "destinations", nil)
	})

	http.HandleFunc("/filsDiscussion", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "filsDiscussion", nil)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		session.Options.MaxAge = -1
		session.Save(r, w)
		forum.Logout(w, r)
	})

	http.HandleFunc("/search-suggestions", forum.SearchSuggestionsHandler)

	// Route pour ajouter un chat
	http.Handle("/addChat", forum.AuthMiddleware(http.HandlerFunc(forum.AddChat)))

	// Route pour la gestion des WebSockets
	http.HandleFunc("/ws", handleConnections)
	/*   CONTACT FILE
	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "contact")
	})
	*/

	// Démarrer le serveur
	fmt.Println("Serveur lancé sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	user := session.Values["user"]
	if user == nil {
		http.Error(w, "Utilisateur non connecté", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur lors de la mise à niveau de la connexion :", err)
		return
	}
	defer ws.Close()

	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Erreur lors de la lecture du message :", err)
			break
		}
		fmt.Printf("Message reçu de %v : %v\n", user, msg)

		err = ws.WriteJSON(msg)
		if err != nil {
			fmt.Println("Erreur lors de l'écriture du message :", err)
			break
		}
	}
}
