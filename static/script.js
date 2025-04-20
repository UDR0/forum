<<<<<<< HEAD
// Sélectionner les éléments
const burgerMenu = document.getElementById("burger"); // Le bouton du menu burger
const navMenu = document.querySelector(".meta-ul"); // Le menu de navigation

// Écouter les clics sur le bouton burger
burgerMenu.addEventListener("click", () => {
    navMenu.classList.toggle("active"); // Ajouter/retirer la classe 'active'
});


// Sélectionne tous les éléments qui ont les classes "destination-region-popular" et "filPrincipal-region"
document.querySelectorAll(".destination-region-popular", ".filPrincipal-region").forEach((card) => {
    card.addEventListener("click", function () { //  Écoute les événements pour le clic sur chaque élément
        const targetUrl = card.getAttribute("data-link");  // Récupère l'URL cible à partir de l'attribut "data-link" de l'élément
        if (targetUrl) {
            window.location.href = targetUrl; // Redirige la fenêtre actuelle vers l'URL
        }
    });
});

document.querySelectorAll('.destination-coeur-container').forEach(container => {
    container.addEventListener("click", function (event) {
        event.stopPropagation(); // Prevent parent element click
        event.preventDefault(); // Prevent default action

        const img = this.querySelector('.destination-coeur');
        const regionName = this.getAttribute('data-region'); // Get region name

        if (!img || !regionName) {
            console.error('Error: Missing heart icon or region name.');
            return;
        }

        // Determine liked status
        const liked = !img.src.includes('coeur_rouge.png'); // True if switching to red heart

        // Optimistically toggle the heart icon
        img.src = liked ? 'static/img/icon/coeur_rouge.png' : 'static/img/icon/coeur.png';

        // Send the like status to the server
        fetch('/like', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ region: regionName, liked }),
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to update like status on the server.');
                }
                return response.json();
            })
            .then(data => {
                console.log(`Server response:`, data); // Log successful server response
            })
            .catch(error => {
                console.error('Error communicating with the server:', error);

                // Revert the heart icon if the request fails
                img.src = liked ? 'static/img/icon/coeur.png' : 'static/img/icon/coeur_rouge.png';
            });
    });
});


document.querySelectorAll('.chat-coeur-container').forEach(container => {
    container.addEventListener("click", function (event) {
        event.stopPropagation();
        event.preventDefault();
        const img = this.querySelector('.chat-coeur');
        const chatName = this.getAttribute('data-chat'); // Get region name

        // Determine liked status
        const liked = !img.src.includes('coeur_rouge.png'); // True if red heart is being added

        // Toggle heart icon
        if (liked) {
            img.src = 'static/img/icon/coeur_rouge.png'; // Change to red heart
            console.log(`Liked region: ${chatName}`);
        } else {
            img.src = 'static/img/icon/coeur.png'; // Change back to normal heart
            console.log(`Unliked region: ${chatName}`);
        }

        // Send data to the server
        fetch('/likechat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ region: chatName, liked }),
        })
            .then(response => response.json())
            .then(data => {
                console.log(`Server response:`, data); // Log server's response
            })
            .catch(error => {
                console.error('Error communicating with the server:', error);
            });
    });
});


// Gestion des avatars dans le pop-up de profil
document.addEventListener("DOMContentLoaded", function () {
    const avatars = document.querySelectorAll(".imgAvatar img");
    const photoProfil = document.getElementById("photoProfil");

    avatars.forEach(avatar => {
        avatar.addEventListener("click", function () {
            // Change l'image de profil
            photoProfil.src = this.src;
            closePopupProfil(); // Ferme le pop-up de profil après sélection
        });
    });
});


// ------------------- Profil ---------------------------//


// Function to open the pop-up to modify the profile photo
function openPopupProfil() {
    document.getElementById("overlay-profil").style.display = "block";
    document.getElementById("popup-profil").style.display = "block";
}

// Function to close the pop-up for the profile photo
function closePopupProfil() {
    document.getElementById("overlay-profil").style.display = "none";
    document.getElementById("popup-profil").style.display = "none";
}

// ------------------- Modifications Pop-up ---------------------------//

// Function to open the pop-up with overlay for modifying the profile
function openPopupModif() {
    document.getElementById("overlay-modif").style.display = "block";
    document.getElementById("nouveauPseudo").value = document.getElementById("pseudo").innerText;
    document.getElementById("nouvelleBio").value = document.getElementById("bio").innerText;
    document.getElementById("popup-modif").style.display = "block";
}

// Function to close the pop-up and overlay
function closePopupModif() {
    // CHANGED: Fixed invalid "null" to "none"
    document.getElementById("overlay-modif").style.display = "none";
    document.getElementById("popup-modif").style.display = "none";
}

// Function to save modifications and update the content
function sauverModifications() {
    const nouveauPseudo = document.getElementById("nouveauPseudo").value;
    const nouvelleBio = document.getElementById("nouvelleBio").value;

    // Mettre à jour l'affichage local
    document.getElementById("pseudo").innerText = nouveauPseudo;
    document.getElementById("bio").innerText = nouvelleBio;

    // Envoyer les données au serveur
    fetch('/updateProfile', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            pseudo: nouveauPseudo,
            bio: nouvelleBio
        })
    })
        .then(response => {
            if (response.ok) {
            } else {
                console.error("Erreur lors de la mise à jour.");
            }
        })
        .catch(error => {
            console.error("Erreur de connexion :", error);
        });

    closePopupModif();
}


function changeavatar(newSrc) {
    document.getElementById("photo_url").value = newSrc;
}

// ------------------- Pop-up for Additional Features ---------------------------//
function updateAvatar(avatarURL) {
    fetch('/updateAvatar', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            avatar: avatarURL
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Errore durante l\'aggiornamento dell\'avatar.');
            }
            return response.json();
        })
        .then(data => {
            // Ricarica la pagina
            window.location.reload();
        })
        .catch(error => console.error("Errore:", error));
}



// Fonction pour filtrer les suggestions dans le menu déroulant en fonction de la recherche de l'utilisateur
function filterOptions() {
    const dropdown = document.getElementById('dropdown');   // Récupère l'élément du menu déroulant
    const searchBar = document.getElementById('searchBar');
    const input = searchBar.value.toLowerCase();
    dropdown.innerHTML = '';  // Efface les résultats précédents du menu déroulant avant d'afficher les nouveaux

    if (input.length >= 2) { // Effectue la recherche une fois que la saisie de l'utilisateur contient au moins 2 caractères 
        fetch(`/search?q=${input}`)    // Envoie une requête au serveur pour récupérer les options filtrées
            .then(response => response.json())
            .then(filteredOptions => {
                // Vérifie que filteredOptions est bien un tableau avant d'appliquer la boucle
                if (filteredOptions && Array.isArray(filteredOptions)) {
                    filteredOptions.forEach(option => {  // Parcourt chaque option filtrée et l'ajoute au menu déroulant
                        const displayText = `${option.departmentName}, ${option.regionName}`;
                        const item = document.createElement('div');
                        item.textContent = displayText;
                        item.className = 'dropdown-item';
                        item.onclick = () => selectOption(displayText, option.regionName);
                        dropdown.appendChild(item);
                    });
                }

                // Affiche ou masque le menu déroulant en fonction de la présence d'options valides
                dropdown.style.display = filteredOptions?.length > 0 ? 'block' : 'none';
            })
            .catch(error => console.error("Error fetching options:", error));
    } else {
        // Masque le menu déroulant si la recherche contient moins de 2 caractères
        dropdown.style.display = 'none';
    }
}


function selectOption(displayText, regionName) {
    searchBar.value = displayText; // Met à jour la barre de recherche une fois que l'utilisateur a choisi une suggestion
    dropdown.style.display = 'none';
}



// Attend que la page html soit entièrement chargé avant d'exécuter le script
document.addEventListener("DOMContentLoaded", () => {
    // Récupère les éléments liés aux messages
    const messageInput = document.getElementById("messageInput");
    const sendButton = document.getElementById("sendButton");
    const messageContainer = document.getElementById("message-container");

    // Récupère les éléments liés à la searchbar
    const searchBar = document.getElementById("searchBar");
    const searchIcon = document.getElementById("search-icon");



    if (messageContainer) {
        scrollToBottom(); // Fait défiler automatiquement le conteneur des messages vers le bas après le rechargement de la page

        setInterval(() => { fetchMessages() }, 4000); // Récupère les messages automatiquement toutes les 4 secondes

        messageContainer.addEventListener("click", (event) => { // Ajoute un écouteur d'événements pour détecter le clic sur le coeur
            if (event.target.classList.contains("msg-like")) {
                heartMsg(event.target); // Appel la fonction heartMsg 
            }
        });



        // Fonction qui envoi un message
        function sendMessage() {
            const message = messageInput.value.trim();

            if (message === "") {
                alert("Le message ne peut pas être vide.");
                return;
            }

            // Envoi du message au serveur via une requête POST
            fetch("/send-message", {
                method: "POST",
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded",
                },
                body: `message=${encodeURIComponent(message)}`,
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error("Erreur lors de l'envoi du message.");
                    }
                    messageInput.value = ""; // Efface la zone de texte après l'envoi

                    // Recharge les messages juste apres en avoir envoyé un
                    return fetchMessages();
                })
                .then(() => {
                    scrollToBottom(); // Assure le défilement après l'envoi du nouveau message
                })
                .catch(error => {
                    console.error("Erreur :", error);
                });
        }

        // Ajoute un évènnement d'écoute sur la touche Entrer
        messageInput?.addEventListener("keydown", (event) => {
            if (event.key === "Enter" && !event.shiftKey) {
                event.preventDefault();
                sendMessage();
            }
        });

        // Ajoute un évènnement d'écoute sur le bouton
        sendButton?.addEventListener("click", sendMessage);
    }

    function fetchMessages() {
        return fetch('/fetch-messages?chatname=ChatNamePlaceholder') // Remplace par le nom réel du chat
            .then(response => response.json())
            .then(messages => {
                const messageContainer = document.getElementById("message-container");

                // Vérifie que messageContainer existe et que messages n'est pas null ou undefined
                if (messageContainer && Array.isArray(messages)) {
                    messageContainer.innerHTML = ""; // Efface les messages actuels avant d'afficher les nouveaux

                    messages.forEach(msg => {
                        const postDiv = document.createElement("div");
                        postDiv.className = "post";

                        postDiv.innerHTML = `
                        <div class="infoPost">
                            <img src="${msg.img_user}" class="photoProfil" alt="Photo de profil">
                            <div class="txtInfoPost">
                                <h3>${msg.sender}</h3>
                                <h4>${msg.time_elapsed}</h4>
                            </div>
                        </div>
                        <div class="message">
                            <p>${msg.message}</p>
                        </div>
                        <div class="msg-coeur-container" data-message-id="${msg.message_id}">
                            ${msg.user_liked
                                ? `<img src="static/img/icon/coeur_rouge.png" alt="Liked" class="msg-like">`
                                : `<img src="static/img/icon/coeur.png" alt="Like" class="msg-like">`
                            }
                            <p>${msg.number_of_likes}</p>
                        </div>
                    `;

                        messageContainer.appendChild(postDiv);
                    });
                }
            })
            .catch(error => console.error("Erreur lors de la récupération des messages :", error));
    }


    // Fonction pour gérer le like/unlike d'un message
    function heartMsg(heartIcon) {
        const msgContainer = heartIcon.closest(".msg-coeur-container");
        const messageId = msgContainer?.getAttribute("data-message-id");
        const likeCountElement = msgContainer?.querySelector("p");

        if (messageId && likeCountElement) {
            const isLiked = heartIcon.src.includes("coeur_rouge.png");
            heartIcon.src = isLiked ? "static/img/icon/coeur.png" : "static/img/icon/coeur_rouge.png";

            // Met à jour visuellement le nombre de likes
            likeCountElement.textContent = parseInt(likeCountElement.textContent, 10) + (isLiked ? -1 : 1);

            // Envoie le statut du like au serveur
            fetch("/like-message", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ message_id: parseInt(messageId, 10), liked: !isLiked }),
            })
                .then(response => response.json())
                .catch(error => {
                    console.error("Error updating like status:", error);
                });
        }
    }

    // Fonction pour faire défiler la boîte de messages vers le bas automatiquement
    function scrollToBottom() {
        messageContainer.scrollTop = messageContainer.scrollHeight;
    }

    // Function qui redirige vers une région
    function redirectToRegion() {
        const searchValue = searchBar.value.trim(); // Prend les valeurs de la searchbar

        // Regarde si la recherche est sous la forme "DepartmentName, RegionName"
        if (searchValue.includes(',')) {
            const regionName = searchValue.split(',')[1].trim(); // Prend le nom de la région après la virgule
            window.location.href = `/region?name=${encodeURIComponent(regionName)}`; // Cherche la région
        } else {
            alert('Veuillez sélectionner une option valide !');
        }
    }

    if (searchBar) { // La recherche va se faire losque l'utilisateur utilise la touche Entrée ou appuie sur le bouton
        searchBar.addEventListener("keydown", (event) => {
            if (event.key === "Enter") {
                event.preventDefault();
                redirectToRegion();
            }
        });

        searchIcon.addEventListener("click", redirectToRegion);
    }

});

// Fonction pour sélectionner une région spécifique via une requête GET
function selectRegion(regionName) {
    fetch(`/region?name=${regionName}`, { method: 'GET' })
        .then(response => {
            if (response.ok) {
                window.location.href = "/welcome"; // Redirige vers la page de fils de discussions après sélection de la région
            } else {
                console.error("Failed to select region:", response.statusText);
            }
        })
        .catch(error => console.error("Error selecting region:", error));
}

// Fonction pour sélectionner un chat spécifique
function selectChat(chatName) {
    fetch(`/select-chat?chatname=${chatName}`)
        .then(response => {
            if (response.ok) {
                window.location.href = "/chat_messages"; // Redirige vers la page des messages du chat après sélection
            } else {
                console.error("Failed to select chat:", response.statusText);
            }
        })
        .catch(error => console.error("Error selecting chat:", error));
}

function PopupFils() {
    const popupFils = document.getElementById("popupAjouterFil");
    const imageBtnFils = document.getElementById("btnAjouterFil");

    if (popupFils.style.display === "flex") {
        document.getElementById("popupAjouterFil").style.display = "none";
        document.getElementById("btnAjouterFil").src = "static/img/icon/ajouter.png"
    } else {
        document.getElementById("popupAjouterFil").style.display = "flex";
        document.getElementById("btnAjouterFil").src = "static/img/icon/moins.png"
    }
}



// Fonction pour créer un nouveau chat
function createChat() {
    const chatTitle = document.getElementById("chatTitle").value.trim(); // Récupère le nom
    const chatDescription = document.getElementById("chatDescription").value.trim(); // Récupère la description
    const region = document.getElementById("regionField").value; // Récupère la région

    if (!chatTitle || !region) {  // Vérifie que le titre et la région sont bien renseignés
        alert("Le titre et la région sont obligatoires !");
        return;
    }

    fetch("/create-chat", { // Envoie les données du chat au serveur 
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded",
        },
        body: `chatname=${encodeURIComponent(chatTitle)}&description=${encodeURIComponent(chatDescription)}&region=${encodeURIComponent(region)}`
    })
        .then(response => {
            if (response.ok) {
                window.location.href = "/welcome"; // Redirige vers la page de fils de discussion après la création du chat
            } else {
                console.error("Failed to create chat:", response.statusText);
                alert("Échec de la création du chat. Veuillez réessayer !");
            }
        })
        .catch(error => {
            console.error("Error creating chat:", error);
            alert("Une erreur s'est produite lors de la création du chat.");
        });
}



/////////////////////////// Chat //////////////////////////

function adjustHeight(textarea) {
    // Calculer la hauteur du texte en fonction du contenu
    const scrollHeight = textarea.scrollHeight;
    const minHeight = 30;  // Hauteur minimale (en pixels, équivalent à environ 1 ligne)
    const maxHeight = 50;  // Hauteur maximale (en pixels, équivalent à environ 3 lignes)

    // Appliquer la hauteur du textarea en fonction du contenu
    if (scrollHeight <= maxHeight) {
        // Si la hauteur du texte est inférieure à la hauteur maximale, ajuster la hauteur
        textarea.style.height = scrollHeight + 'px';
    } else {
        // Si la hauteur dépasse la hauteur maximale, on applique un scroll
        textarea.style.height = maxHeight + 'px';
        textarea.style.overflowY = 'auto'; // Activer le scroll vertical
    }

    // Ne pas descendre sous la hauteur minimale
    if (textarea.value == "") {
        textarea.style.height = 1.5 + 'em';
    }
}
=======
// Sélectionner les éléments
const burgerMenu = document.getElementById("burger"); // Le bouton du menu burger
const navMenu = document.querySelector(".meta-ul"); // Le menu de navigation

// Écouter les clics sur le bouton burger
burgerMenu.addEventListener("click", () => {
    navMenu.classList.toggle("active"); // Ajouter/retirer la classe 'active'
});

document.querySelectorAll(".destination-region-popular", ".filPrincipal-region").forEach((card) => {
    card.addEventListener("click", function () {
        const targetUrl = card.getAttribute("data-link");
        if (targetUrl) {
            window.location.href = targetUrl;
        }
    });
});

document.querySelectorAll('.destination-coeur-container').forEach(container => {
    container.addEventListener("click", function(event) {
        event.stopPropagation(); // Prevent parent element click
        event.preventDefault(); // Prevent default action

        const img = this.querySelector('.destination-coeur');
        const regionName = this.getAttribute('data-region'); // Get region name

        if (!img || !regionName) {
            console.error('Error: Missing heart icon or region name.');
            return;
        }

        // Determine liked status
        const liked = !img.src.includes('coeur_rouge.png'); // True if switching to red heart

        // Optimistically toggle the heart icon
        img.src = liked ? 'static/img/icon/coeur_rouge.png' : 'static/img/icon/coeur.png';

        // Send the like status to the server
        fetch('/like', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ region: regionName, liked }),
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to update like status on the server.');
            }
            return response.json();
        })
        .then(data => {
            console.log(`Server response:`, data); // Log successful server response
        })
        .catch(error => {
            console.error('Error communicating with the server:', error);

            // Revert the heart icon if the request fails
            img.src = liked ? 'static/img/icon/coeur.png' : 'static/img/icon/coeur_rouge.png';
        });
    });
});


document.querySelectorAll('.chat-coeur-container').forEach(container => {
    container.addEventListener("click", function(event) {
        event.stopPropagation();
        event.preventDefault();
        const img = this.querySelector('.chat-coeur');
        const chatName = this.getAttribute('data-chat'); // Get region name
        
        // Determine liked status
        const liked = !img.src.includes('coeur_rouge.png'); // True if red heart is being added
        
        // Toggle heart icon
        if (liked) {
            img.src = 'static/img/icon/coeur_rouge.png'; // Change to red heart
            console.log(`Liked region: ${chatName}`);
        } else {
            img.src = 'static/img/icon/coeur.png'; // Change back to normal heart
            console.log(`Unliked region: ${chatName}`);
        }

        // Send data to the server
        fetch('/likechat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ region: chatName, liked }),
        })
        .then(response => response.json())
        .then(data => {
            console.log(`Server response:`, data); // Log server's response
        })
        .catch(error => {
            console.error('Error communicating with the server:', error);
        });
    });
});


// Gestion des avatars dans le pop-up de profil
document.addEventListener("DOMContentLoaded", function () {
    const avatars = document.querySelectorAll(".imgAvatar img");
    const photoProfil = document.getElementById("photoProfil");

    avatars.forEach(avatar => {
        avatar.addEventListener("click", function () {
            // Change l'image de profil
            photoProfil.src = this.src;
            closePopupProfil(); // Ferme le pop-up de profil après sélection
        });
    });
});


// ------------------- Profil ---------------------------//


// Function to open the pop-up to modify the profile photo
function openPopupProfil() {
    document.getElementById("overlay-profil").style.display = "block";
    document.getElementById("popup-profil").style.display = "block";
}

// Function to close the pop-up for the profile photo
function closePopupProfil() {
    document.getElementById("overlay-profil").style.display = "none";
    document.getElementById("popup-profil").style.display = "none";
}

// ------------------- Modifications Pop-up ---------------------------//

// Function to open the pop-up with overlay for modifying the profile
function openPopupModif() {
    document.getElementById("overlay-modif").style.display = "block";
    document.getElementById("nouveauPseudo").value = document.getElementById("pseudo").innerText;
    document.getElementById("nouvelleBio").value = document.getElementById("bio").innerText;
    document.getElementById("popup-modif").style.display = "block";
}

// Function to close the pop-up and overlay
function closePopupModif() {
    // CHANGED: Fixed invalid "null" to "none"
    document.getElementById("overlay-modif").style.display = "none";
    document.getElementById("popup-modif").style.display = "none";
}

// Function to save modifications and update the content
function sauverModifications() {
    const nouveauPseudo = document.getElementById("nouveauPseudo").value;
    const nouvelleBio = document.getElementById("nouvelleBio").value;

    // Mettre à jour l'affichage local
    document.getElementById("pseudo").innerText = nouveauPseudo;
    document.getElementById("bio").innerText = nouvelleBio;

    // Envoyer les données au serveur
    fetch('/updateProfile', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            pseudo: nouveauPseudo,
            bio: nouvelleBio
        })
    })
    .then(response => {
        if (response.ok) {
            console.log("Mise à jour réussie !");
        } else {
            console.error("Erreur lors de la mise à jour.");
        }
    })
    .catch(error => {
        console.error("Erreur de connexion :", error);
    });

    closePopupModif();
}


// ------------------- Pop-up for Additional Features ---------------------------//
function updateAvatar(avatarURL) {
    fetch('/updateAvatar', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            avatar: avatarURL
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Errore durante l\'aggiornamento dell\'avatar.');
        }
        return response.json();
    })
    .then(data => {
        console.log(data.message);

        // Ricarica la pagina
        window.location.reload();
    })
    .catch(error => console.error("Errore:", error));
}


////////////////////////////////// SEARCHBAR ///////////////////////////////////////// Define functions globally

function filterOptions() {
    const dropdown = document.getElementById('dropdown');
    const searchBar = document.getElementById('searchBar');
    const input = searchBar.value.toLowerCase();
    dropdown.innerHTML = ''; // Clear previous results
    if (input.length >= 2) {
        fetch(`/search?q=${input}`)
            .then(response => response.json())
            .then(filteredOptions => {
                filteredOptions.forEach(option => {
                    const displayText = `${option.departmentName}, ${option.regionName}`;
                    const item = document.createElement('div');
                    item.textContent = displayText;
                    item.className = 'dropdown-item';
                    item.onclick = () => selectOption(displayText, option.regionName);
                    dropdown.appendChild(item);
                });
                dropdown.style.display = filteredOptions.length > 0 ? 'block' : 'none';
            });
    } else {
        dropdown.style.display = 'none';
    }
}

function selectOption(displayText, regionName) {
    searchBar.value = displayText; // Update search bar with selected option
    dropdown.style.display = 'none';
}

// messages, searchbar
document.addEventListener("DOMContentLoaded", () => {
    // Message-related elements
    const messageInput = document.getElementById("messageInput");
    const sendButton = document.getElementById("sendButton");
    const messageContainer = document.getElementById("message-container");

    if (messageContainer) {
        // Automatically scroll to the bottom after the page reloads
        scrollToBottom();

        // Automatically fetch messages every 4 seconds
        setInterval(() => {
            fetchMessages().then(() => {
                scrollToBottom(); // Ensure the view scrolls after messages are fetched
            });
        }, 4000);

        // Use event delegation for heart icon clicks
        messageContainer.addEventListener("click", (event) => {
            if (event.target.classList.contains("msg-like")) {
                heartMsg(event.target); // Call the heartMsg function
            }
        });

        // Function to send a message
        function sendMessage() {
            const message = messageInput.value.trim();

            if (message === "") {
                alert("Le message ne peut pas être vide.");
                return;
            }

            fetch("/send-message", {
                method: "POST",
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded",
                },
                body: `message=${encodeURIComponent(message)}`,
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error("Erreur lors de l'envoi du message.");
                    }
                    messageInput.value = ""; // Clear the textarea after sending

                    // Reload messages immediately after sending a new one
                    return fetchMessages();
                })
                .then(() => {
                    scrollToBottom(); // Ensure scroll happens after new messages are appended
                })
                .catch(error => {
                    console.error("Erreur :", error);
                });
        }

        // Add event listener for Enter key
        messageInput?.addEventListener("keydown", (event) => {
            if (event.key === "Enter" && !event.shiftKey) {
                event.preventDefault(); // Prevent adding a new line
                sendMessage();
            }
        });

        // Add event listener for Send button
        sendButton?.addEventListener("click", sendMessage);
    }

    // Function to fetch messages and update the container
    function fetchMessages() {
        return fetch('/fetch-messages?chatname=ChatNamePlaceholder') // Replace "ChatNamePlaceholder" with the actual chat name
            .then(response => response.json())
            .then(messages => {
                const messageContainer = document.getElementById("message-container");
                messageContainer.innerHTML = ""; // Clear current messages
    
                messages.forEach(msg => {
                    const postDiv = document.createElement("div");
                    postDiv.className = "post";
    
                    // Dynamically create the HTML structure for each message
                    postDiv.innerHTML = `
                        <div class="infoPost">
                            <img src="${msg.img_user}" class="photoProfil" alt="Photo de profil">
                            <div class="txtInfoPost">
                                <h3>${msg.sender}</h3>
                                <h4>${msg.time_elapsed}</h4>
                            </div>
                        </div>
                        <div class="message">
                            <p>${msg.message}</p>
                        </div>
                        <div class="msg-coeur-container" data-message-id="${msg.message_id}">
                            ${msg.user_liked
                                ? `<img src="static/img/icon/coeur_rouge.png" alt="Liked" class="msg-like">`
                                : `<img src="static/img/icon/coeur.png" alt="Like" class="msg-like">`
                            }
                            <p>${msg.number_of_likes}</p>
                        </div>
                    `;
    
                    // Append the new message to the container
                    messageContainer.appendChild(postDiv);
                });
            })
            .catch(error => console.error("Erreur lors de la récupération des messages :", error));
    }
    

    // Function to handle heart icon clicks (likes)
    function heartMsg(heartIcon) {
        console.log("Heart icon clicked!");

        // Locate the container with the data-message-id attribute
        const msgContainer = heartIcon.closest(".msg-coeur-container");
        const messageId = msgContainer?.getAttribute("data-message-id"); // Get the message ID from the container

        if (messageId) {
            // Log the message ID for debugging
            console.log(`Message ID: ${messageId}`);

            // Optional: Toggle heart icon state
            const isLiked = heartIcon.src.includes("coeur_rouge.png");
            heartIcon.src = isLiked ? "static/img/icon/coeur.png" : "static/img/icon/coeur_rouge.png";

            // Send the like status to the server
            fetch("/like-message", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ message_id: parseInt(messageId, 10), liked: !isLiked }),
            })
                .then((response) => {
                    if (!response.ok) {
                        throw new Error("Failed to update like status on the server.");
                    }
                    return response.json();
                })
                .then((data) => {
                    console.log(`Like status updated for message ${messageId}:`, data);
                })
                .catch((error) => {
                    console.error("Error updating like status:", error);
                });
        } else {
            console.error("Message ID not found for the clicked heart icon.");
        }
    }

    // Function to scroll to the bottom of the message container
    function scrollToBottom() {
        messageContainer.scrollTop = messageContainer.scrollHeight;
    }

    // Fetch messages on page load
    fetchMessages();

    // Handle search actions
    const searchBar = document.getElementById("searchBar");
    const searchIcon = document.getElementById("search-icon");

    // Function to redirect to region
    function redirectToRegion() {
        const searchValue = searchBar.value.trim(); // Safely get the value from the search bar
        
        // Check if the input follows the expected format "DepartmentName, RegionName"
        if (searchValue.includes(',')) {
            const regionName = searchValue.split(',')[1].trim(); // Extract region name after the comma
            window.location.href = `/region?name=${encodeURIComponent(regionName)}`; // Navigate to the desired URL
        } else {
            alert('Veuillez sélectionner une option valide !'); // Feedback for invalid input
        }
    }

    // Add event listener for Enter key in the search bar
    searchBar.addEventListener("keydown", (event) => {
        if (event.key === "Enter") { // Check for "Enter" key press
            event.preventDefault(); // Prevent form submission or default behavior
            redirectToRegion();
        }
    });

    // Add event listener for search icon click
    searchIcon.addEventListener("click", redirectToRegion);
});

/////////////////////////// Chat //////////////////////////

function adjustHeight(textarea) {

    // Calculer la hauteur du texte en fonction du contenu
    const scrollHeight = textarea.scrollHeight;
    const minHeight = 30;  // Hauteur minimale (en pixels, équivalent à environ 1 ligne)
    const maxHeight = 50;  // Hauteur maximale (en pixels, équivalent à environ 3 lignes)

    // Appliquer la hauteur du textarea en fonction du contenu
    if (scrollHeight <= maxHeight) {
        // Si la hauteur du texte est inférieure à la hauteur maximale, ajuster la hauteur
        textarea.style.height = scrollHeight + 'px';
    } else {
        // Si la hauteur dépasse la hauteur maximale, on applique un scroll
        textarea.style.height = maxHeight + 'px';
        textarea.style.overflowY = 'auto'; // Activer le scroll vertical
    }

    // Ne pas descendre sous la hauteur minimale
    if (textarea.value == "") {
        textarea.style.height = 1.5 + 'em';
    }
}
>>>>>>> 6c0ac1947bce57893d165fee60e3fe15f92a36db
