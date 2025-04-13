# MyTripy

Un forum collaboratif dédié à l'exploration et le partage des expériences sur les magnifiques régions et départements de France. MyTripy permet aux utilisateurs de se connecter, d'échanger des conseils et de découvrir la culture, la géographie et le charme des destinations françaises.

## Description 

MyTripy est une plateforme web conçue pour inspirer les passionnés de voyage à partager leurs expériences sur les régions et départements français. Développée avec Go (Golang), HTML, CSS et une base de données (SQLite), MyTripy est une application intuitive et interactive. Grâce à l’intégration de Docker, l’accès à notre site est simplifié quel que soit le système d’exploitation utilisé. Que vous soyez un voyageur expérimenté ou un explorateur novice, MyTripy vous propose une communauté conviviale dédiée au partage et à la découverte.

## Prise en main

### Dépendances 

Avant de configurer MyTripy, assurez-vous d'avoir les éléments suivants installés :
* **Go** (version 1.24.1)
* **Docker** (version 28.0.1)
* **SQLite** (version 3.38.2)
* Un navigateur web (Chrome, Firefox, etc.)

### Installation 

1. Clonez ce dépôt :
   ```bash
   git clone https://github.com/UDR0/forum.git
   cd forum

2. Construction de l'image :
    ```bash
    docker build -t forum .

3. Lancement de l'image docker :
    ```
    docker run -v ${pwd}\forum.db:/app/forum.db -p 8080:8080 forum

### Aide  

Si vous rencontrez des problèmes lors de l'installation ou de l'exécution de MyTripy, voici quelques étapes de dépannage :

1. **Vérifiez la base de données** : 
   - Assurez-vous que le fichier `forum.db` est présent dans l'image docker après le lancement et qu'il contient les bonnes données.

2. **Conflits de port** : 
   - Si le port **8080** est déjà utilisé sur votre machine, modifiez-le lors de l'exécution de l'image Docker :
     ```bash
     docker run -v ${pwd}/forum.db:/app/forum.db -p <nouveau_port>:8080 forum
     ```

3. **Problèmes avec Docker** :
   - Vérifiez que Docker est correctement installé et configuré en utilisant la commande suivante :
     ```bash
     docker --version
     ```
   - Assurez-vous que Docker est en cours d'exécution sur votre machine.


## Auteurs

* **Sara SMITH** 
* **Mathias BOUCHENOIRE** 
* **Samuel BOUHNIK-LOURY** 
* **William PONS**
