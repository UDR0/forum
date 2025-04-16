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
        img.src = liked ? 'static/img/coeur_rouge.png' : 'static/img/coeur.png';

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
            img.src = liked ? 'static/img/coeur.png' : 'static/img/coeur_rouge.png';
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
            img.src = 'static/img/coeur_rouge.png'; // Change to red heart
            console.log(`Liked region: ${chatName}`);
        } else {
            img.src = 'static/img/coeur.png'; // Change back to normal heart
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
        return fetch('/fetch-messages?chatname=ChatNamePlaceholder')
            .then(response => response.json())
            .then(messages => {
                messageContainer.innerHTML = ""; // Clear current messages

                messages.forEach(msg => {
                    const postDiv = document.createElement("div");
                    postDiv.className = "post";

                    // Dynamically add data-message-id in msg-coeur-container
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
                            <img src="static/img/coeur.png" alt="Like" class="msg-like">
                        </div>
                    `;
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
            heartIcon.src = isLiked ? "static/img/coeur.png" : "static/img/coeur_rouge.png";

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
});

function redirectToRegion() {
    const searchValue = document.getElementById('searchBar')?.value.trim(); // Safely get the value
    if (searchValue.includes(',')) {
        const regionName = searchValue.split(',')[1].trim(); // Extract region name after the comma
        window.location.href = `/region?name=${encodeURIComponent(regionName)}`; // Navigate to the desired URL
    } else {
        alert('Veuillez sélectionner une option valide !'); // Feedback for invalid input
    }
}


function redirectToRegion() {
    const searchValue = searchBar.value.trim(); // Get the value from the search bar
    
        // Check if the input follows the expected format "DepartmentName, RegionName"
        if (searchValue.includes(',')) {
            const regionName = searchValue.split(',')[1].trim(); // Extract region name after the comma
            window.location.href = `/region?name=${encodeURIComponent(regionName)}`; // Navigate to the desired URL
        } else {
            alert('Veuillez sélectionner une option valide !'); // User feedback for invalid input
        }
}