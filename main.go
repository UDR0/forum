package main

import (
	"fmt"
	forum "forum/Functions" // import du packet qui contient les fonctions du forum
	"html/template"
	"net/http"

	"github.com/gorilla/sessions" // go get github.com/gorilla/sessions
)

// définit cookie store pour gérer les sessions utilisateur
var (
	store = sessions.NewCookieStore([]byte("something-very-secret"))
)

// fonction qui charger et rend un template HTML
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := fmt.Sprintf("templates/%s.html", tmpl)
	t, err := template.ParseFiles(tmplPath) // analyse le fichier template
	if err != nil {                         // message d'erreur si le template ne peut pas être chargé
		http.Error(w, "Erreur lors du chargement du template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil { // exécutez le template avec les données fournies en paramètre
		http.Error(w, "Erreur lors de l'exécution du template : "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// servir les fichiers statiques (CSS, JS, images)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// route pour la page d'accueil
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// si l'URL est "/" alors on redirige vers une autre page
		http.Redirect(w, r, "/mytripy-non", http.StatusFound)
	})

	// routes pour les pages "mytripy-non.HTML"
	http.HandleFunc("/mytripy-non", forum.MyTripyNonHandler)

	// route pour afficher la page "À propos"
	http.HandleFunc("/apropos", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "apropos", nil)
	})

	// route pour la connexion des utilisateurs
	http.HandleFunc("/SeConnecter", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { // vérifie les identifiants de connexion
			forum.CheckCredentialsForConnection(w, r)
			session, _ := store.Get(r, "session")
			session.Values["user"] = r.FormValue("username") // crée une session utilisateur
			session.Save(r, w)
			http.Redirect(w, r, "/profil", http.StatusFound) // redirige vers la page de profil
		} else {
			// rend le template de connexion
			renderTemplate(w, "SeConnecter", nil)
		}
	})

	// route pour la création de compte utilisateur
	http.HandleFunc("/CreerCompte", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			forum.CreateUser(w, r) // appelle la fonction pour créer un compte
		} else {
			renderTemplate(w, "CreerCompte", nil) // rend le template de création de compte
		}
	})

	// route pour afficher la page "Mot de passe oublié"
	http.HandleFunc("/mot-de-passe-oublie", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "mot-de-passe-oublie", nil)
	})

	// routes pour afficher la page de profil et modifier son contenu
	http.HandleFunc("/profil", forum.ProfilPage)
	http.HandleFunc("/updateProfile", forum.UpdateProfile)
	http.HandleFunc("/updateAvatar", forum.UpdateAvatar)
	http.HandleFunc("/destinations", forum.AllRegions)

	// routes pour gérer les "likes"
	http.HandleFunc("/like", forum.LikeHandler)
	http.HandleFunc("/likechat", forum.LikeChatHandler)
	http.HandleFunc("/like-message", forum.LikeMessageHandler)

	// route pour la déconnexion de l'utilisateur
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		// invalide la session utilisateur
		session, _ := store.Get(r, "session")
		session.Options.MaxAge = -1
		session.Save(r, w)
		// exécute la fonction de déconnexion
		forum.Logout(w, r)
	})

	// route pour gérer les suggestions de recherche
	http.HandleFunc("/search", forum.SearchSuggestionsHandler)

	// Routes pour gérer les discussions et messages
	// discussions
	http.HandleFunc("/welcome", forum.FileDiscussion)
	http.HandleFunc("/create-chat", forum.CreateChatHandler)
	http.HandleFunc("/select-chat", forum.SelectChatHandler)
	http.HandleFunc("/fetch-chats", forum.FetchChatsHandler)
	// messages
	http.HandleFunc("/chat_messages", forum.FilMessagesHandler)
	http.HandleFunc("/send-message", forum.SendMessageHandler)
	http.HandleFunc("/fetch-messages", forum.FetchMessagesHandler)

	// route pour les régions spécifiques
	http.HandleFunc("/region", forum.RegionHandler)

	// Démarrer le serveur sur le port :8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}
