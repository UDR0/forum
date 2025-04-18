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
            location.reload(); 
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


function changeavatar(newSrc){
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


////////////////////////////////// SEARCHBAR ///////////////////////////////////////// Define functions globally
/*
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
}*/
function filterOptions() {
    const dropdown = document.getElementById('dropdown');
    const searchBar = document.getElementById('searchBar');
    const input = searchBar.value.toLowerCase();
    dropdown.innerHTML = ''; // Clear previous results

    if (input.length >= 2) {
        fetch(`/search?q=${input}`)
            .then(response => response.json())
            .then(filteredOptions => {
                // Ensure filteredOptions is valid before calling .forEach
                if (filteredOptions && Array.isArray(filteredOptions)) {
                    filteredOptions.forEach(option => {
                        const displayText = `${option.departmentName}, ${option.regionName}`;
                        const item = document.createElement('div');
                        item.textContent = displayText;
                        item.className = 'dropdown-item';
                        item.onclick = () => selectOption(displayText, option.regionName);
                        dropdown.appendChild(item);
                    });
                }

                // Only show the dropdown if there are valid options
                dropdown.style.display = filteredOptions?.length > 0 ? 'block' : 'none';
            })
            .catch(error => console.error("Error fetching options:", error));
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
        setInterval(() => { fetchMessages() }, 4000);

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

    
    function fetchMessages() {
        return fetch('/fetch-messages?chatname=ChatNamePlaceholder') // Replace with actual chat name
            .then(response => response.json())
            .then(messages => {
                const messageContainer = document.getElementById("message-container");
    
                // Only proceed if messageContainer is not null
                if (messageContainer) {
                    messageContainer.innerHTML = ""; // Clear current messages
    
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
                                    ? `<img src="static/img/coeur_rouge.png" alt="Liked" class="msg-like">`
                                    : `<img src="static/img/coeur.png" alt="Like" class="msg-like">`
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
    

  /*  // Function to fetch messages and update the container
    function fetchMessages() {
        return fetch('/fetch-messages?chatname=ChatNamePlaceholder') // Replace "ChatNamePlaceholder" with the actual chat name
            .then(response => response.json())
            .then(messages => {
                const messageContainer = document.getElementById("message-container");
                messageContainer.innerHTML = ""; // Clear current messages

                if (messageContainer) {
    
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
                                ? `<img src="static/img/coeur_rouge.png" alt="Liked" class="msg-like">`
                                : `<img src="static/img/coeur.png" alt="Like" class="msg-like">`
                            }
                            <p>${msg.number_of_likes}</p>
                        </div>
                    `;
    
                    // Append the new message to the container
                    messageContainer.appendChild(postDiv);
                });
            })

            .catch(error => console.error("Erreur lors de la récupération des messages :", error));
    }*/
    

        function heartMsg(heartIcon) {
            const msgContainer = heartIcon.closest(".msg-coeur-container");
            const messageId = msgContainer?.getAttribute("data-message-id"); 
            const likeCountElement = msgContainer?.querySelector("p"); 
        
            if (messageId && likeCountElement) {
                const isLiked = heartIcon.src.includes("coeur_rouge.png");
                heartIcon.src = isLiked ? "static/img/coeur.png" : "static/img/coeur_rouge.png";
        
                // Adjust the like count
                likeCountElement.textContent = parseInt(likeCountElement.textContent, 10) + (isLiked ? -1 : 1);
        
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
    /*searchBar.addEventListener("keydown", (event) => {
        if (event.key === "Enter") { // Check for "Enter" key press
            event.preventDefault(); // Prevent form submission or default behavior
            redirectToRegion();
        }
    });

    // Add event listener for search icon click
    searchIcon.addEventListener("click", redirectToRegion);*/

    if (searchBar) {
        searchBar.addEventListener("keydown", (event) => {
            if (event.key === "Enter") { // Check for "Enter" key press
                event.preventDefault(); // Prevent form submission or default behavior
                redirectToRegion();
            }
        });

        searchIcon.addEventListener("click", redirectToRegion);
    }
});

function selectRegion(regionName) {
    fetch(`/region?name=${regionName}`, { method: 'GET' })
        .then(response => {
            if (response.ok) {
                window.location.href = "/welcome";
            } else {
                console.error("Failed to select region:", response.statusText);
            }
        })
        .catch(error => console.error("Error selecting region:", error));
}

function selectChat(chatName) {
    fetch(`/select-chat?chatname=${chatName}`)
        .then(response => {
            if (response.ok) {
                window.location.href = "/chat_messages"; // Reindirizza alla pagina dei messaggi
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
        document.getElementById("btnAjouterFil").src = "static/img/ajouter.png"
    } else {
        document.getElementById("popupAjouterFil").style.display = "flex";
        document.getElementById("btnAjouterFil").src = "static/img/moin.png"
    }
}
/*
function fetchChats(region) {
fetch(`/fetch-chats?region=${region}`)
.then(response => response.json())
.then(data => {
    const chatList = document.getElementById("chat-list");
    chatList.innerHTML = ""; // Clear current chats

    // Handle case when no chats are available
    if (!data.Chats || data.Chats.length === 0) {
        const noChatsMessage = document.createElement("p");
        noChatsMessage.textContent = `No chats available in ${region}. Create one below!`;
        chatList.appendChild(noChatsMessage);
        return;
    }

    // Display Principal Chat
    if (data.MainChat) {
        const principalChat = document.createElement("div");
        principalChat.className = "principal-chat";
        principalChat.innerHTML = `
            <h3>Principal Chat</h3>
            <p>Name: ${data.MainChat.Name}</p>
            <p>Description: ${data.MainChat.Descri}</p>
            <p>Messages: ${data.MainChat.MessageCount}</p>
            <p>Total Likes: ${data.MainChat.TotalLikes}</p>
            <img src="${data.MainChat.ImageURL}" alt="Region Image">
        `;
        chatList.appendChild(principalChat);
    }

    // Display User Chats
    const userChatsTitle = document.createElement("h3");
    userChatsTitle.textContent = "User Chats";
    chatList.appendChild(userChatsTitle);

    data.Chats.forEach(chat => {
        const listItem = document.createElement("li");
        listItem.className = "chat-item";

        listItem.innerHTML = `
            <button>${chat.Name} (${chat.Creator})</button>
            <p>Description: ${chat.Descri}</p>
            <p>Messages: ${chat.MessageCount}</p>
            <p>Total Likes: ${chat.TotalLikes}</p>
            <img src="${chat.PhotoURL}" alt="Creator Image">
        `;

        // Event listener for chat selection
        const button = listItem.querySelector("button");
        button.addEventListener("click", () => {
            selectChat(chat.Name); // Function for handling chat selection
        });

        chatList.appendChild(listItem);
    });
})
.catch(error => console.error("Error fetching chats:", error));
}*/

function fetchChats(region) {
    fetch(`/fetch-chats?region=${region}`)
        .then(response => response.json())
        .then(data => {
            const chatList = document.getElementById("chat-list");

            // Only proceed if chatList is not null
            if (chatList) {
                chatList.innerHTML = ""; // Clear current chats

                // Handle case when no chats are available
                if (!data.Chats || data.Chats.length === 0) {
                    const noChatsMessage = document.createElement("p");
                    noChatsMessage.textContent = `No chats available in ${region}. Create one below!`;
                    chatList.appendChild(noChatsMessage);
                    return;
                }

                // Display Principal Chat
                if (data.MainChat) {
                    const principalChat = document.createElement("div");
                    principalChat.className = "principal-chat";
                    principalChat.innerHTML = `
                        <h3>Principal Chat</h3>
                        <p>Name: ${data.MainChat.Name}</p>
                        <p>Description: ${data.MainChat.Descri}</p>
                        <p>Messages: ${data.MainChat.MessageCount}</p>
                        <p>Total Likes: ${data.MainChat.TotalLikes}</p>
                        <img src="${data.MainChat.ImageURL}" alt="Region Image">
                    `;
                    chatList.appendChild(principalChat);
                }

                // Display User Chats
                const userChatsTitle = document.createElement("h3");
                userChatsTitle.textContent = "User Chats";
                chatList.appendChild(userChatsTitle);

                data.Chats.forEach(chat => {
                    const listItem = document.createElement("li");
                    listItem.className = "chat-item";

                    listItem.innerHTML = `
                        <button>${chat.Name} (${chat.Creator})</button>
                        <p>Description: ${chat.Descri}</p>
                        <p>Messages: ${chat.MessageCount}</p>
                        <p>Total Likes: ${chat.TotalLikes}</p>
                        <img src="${chat.PhotoURL}" alt="Creator Image">
                    `;

                    // Event listener for chat selection
                    const button = listItem.querySelector("button");
                    button.addEventListener("click", () => {
                        selectChat(chat.Name); // Function for handling chat selection
                    });

                    chatList.appendChild(listItem);
                });
            }
        })
        .catch(error => console.error("Error fetching chats:", error));
}


function createChat() {
    const chatTitle = document.getElementById("chatTitle").value.trim();
    const chatDescription = document.getElementById("chatDescription").value.trim(); // Fetch description
    const region = document.getElementById("regionField").value;

    if (!chatTitle || !region) {
        alert("Le titre et la région sont obligatoires !");
        return;
    }

    // Send data via POST
    fetch("/create-chat", {
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded",
        },
        body: `chatname=${encodeURIComponent(chatTitle)}&description=${encodeURIComponent(chatDescription)}&region=${encodeURIComponent(region)}`
    })
        .then(response => {
            if (response.ok) {
                window.location.href = "/welcome"; // Redirect to welcome page
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


document.addEventListener("DOMContentLoaded", function () {
    fetchChats("{{.Region}}");
});