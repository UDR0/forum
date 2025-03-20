package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Ouvrir (ou créer) la base SQLite
	db, err := sql.Open("sqlite3", "./forum.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Vérifier la connexion
	err = db.Ping()
	if err != nil {
		log.Fatal("Erreur de connexion à la base :", err)
	}

	fmt.Println("Connexion à SQLite réussie !")
}
