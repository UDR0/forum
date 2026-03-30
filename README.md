# MyTripi

Un forum de voyage permettant aux utilisateurs de partager leurs expériences, poser des questions et découvrir de nouvelles destinations.

---

## Aperçu

<p align="center">
  <img src="static/MyTripy.pdf" width="500">
</p>

---

## Objectif du projet

Créer une plateforme interactive où les utilisateurs peuvent :

- Partager leurs voyages
- Poser des questions
- Discuter avec la communauté
- Découvrir de nouvelles destinations

---

## Architecture du projet

Le projet est structuré de manière claire :


├── docker/ # Configuration Docker
├── static/ # Fichiers statiques (CSS, images…)
├── templates/ # Pages HTML
├── view/ # Logique côté vue / handlers
├── main.go # Point d’entrée du serveur
├── SeConnecter.html
├── styles.css


---

## Technologies utilisées

- Go (Golang) → Backend
- HTML5 → Structure
- CSS3 → Design
- Docker → Environnement

---

## Lancer le projet

### 🔹 Méthode 1 : sans Docker

1. Installer Go
2. Lancer le serveur :
go run main.go
Ouvrir dans le navigateur :
http://localhost:8080

---

### 🔹 Méthode 1 : sans Docker
docker build -t travel-forum .
docker run -p 8080:8080 travel-forum
