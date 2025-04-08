// Sélectionner les éléments
const burgerMenu = document.getElementById("burger"); // Le bouton du menu burger
const navMenu = document.querySelector(".meta-ul"); // Le menu de navigation

// Écouter les clics sur le bouton burger
burgerMenu.addEventListener("click", () => {
    navMenu.classList.toggle("active"); // Ajouter/retirer la classe 'active'
});

document.querySelectorAll('.coeur-container').forEach(coeur => {
    coeur.addEventListener('click', function () {
        const img = this.querySelector('.coeur');
        if (img.src.includes('coeur.png')) {
            img.src = 'static/img/coeur_rouge.png'; // Remplace par le cœur rouge
        } else {
            img.src = 'static/img/coeur.png'; // Reviens au cœur normal
        }
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

// Ouvrir le pop-up pour modifier la photo de profil
function openPopupProfil() {
    document.getElementById("overlay-profil").style.display = "block";
    document.getElementById("popup-profil").style.display = "block";
}

// Fermer le pop-up de la photo de profil
function closePopupProfil() {
    document.getElementById("overlay-profil").style.display = "none";
    document.getElementById("popup-profil").style.display = "none";
}

// Fermer le pop-up en cliquant sur l'overlay
document.getElementById("overlay-profil").onclick = closePopupProfil;




// Ouvrir le pop-up avec overlay sombre
function openPopupModif() {
    document.getElementById("overlay-modif").style.display = "block";
    document.getElementById("nouveauPseudo").value = document.getElementById("pseudo").innerText;
    document.getElementById("nouvelleBio").value = document.getElementById("bio").innerText;
    document.getElementById("popup-modif").style.display = "block";
}

// Fermer le pop-up et l'overlay
function closePopupModif() {
    document.getElementById("overlay-modif").style.display = "none";
    document.getElementById("popup-modif").style.display = "none";
}

// Sauvegarder les modifications
function sauverModifications() {
    const nouveauPseudo = document.getElementById("nouveauPseudo").value;
    const nouvelleBio = document.getElementById("nouvelleBio").value;

    document.getElementById("pseudo").innerText = nouveauPseudo;
    document.getElementById("bio").innerText = nouvelleBio;

    closePopupModif();
}

// Fermer le pop-up en cliquant sur l'overlay
document.getElementById("overlay-modif").onclick = closePopupModif;


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




/* it is not working
    // ------------------- SEARCH ---------------------------//

    var searchInput = document.getElementById("search-input");
    if (searchInput) {
        searchInput.addEventListener('input', function () {
            const query = this.value;
            const autocomBox = document.getElementById('autocom-box');
            
            if (autocomBox) {
                if (query.length < 2) {
                    autocomBox.innerHTML = '';
                    autocomBox.classList.remove('active');
                    return;
                }

                fetch(`/search-suggestions?q=${encodeURIComponent(query)}`)
                    .then(response => {
                        if (!response.ok) {
                            throw new Error('Network response was not ok');
                        }
                        return response.text(); // Get the response as text
                    })
                    .then(text => {
                        try {
                            return JSON.parse(text); // Attempt to parse the text as JSON
                        } catch (error) {
                            throw new Error('Failed to parse response as JSON: ' + text);
                        }
                    })
                    .then(data => {
                        autocomBox.innerHTML = '';
                        
                        data.forEach(suggestion => {
                            const li = document.createElement('li');
                            li.textContent = `${suggestion.department_name}, ${suggestion.region_name}`;
                            
                            li.addEventListener('click', () => {
                                document.getElementById('search-input').value = li.textContent;
                                autocomBox.innerHTML = '';
                                autocomBox.classList.remove('active');
                            });
                            
                            autocomBox.appendChild(li);
                        });

                        autocomBox.classList.add('active');
                    })
                    .catch(error => console.error('Error fetching suggestions:', error));
            }
        });
    } else {
        console.error("Element with ID 'search-input' not found.");
    }
    */

// ------------------- WebSocket ---------------------------//

let ws;

function connectWebSocket() {
    ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = () => {
        console.log('Connecté au serveur WebSocket');
    };

    ws.onmessage = (event) => {
        const messages = document.getElementById('messages');
        const message = document.createElement('li');
        message.textContent = event.data;
        messages.appendChild(message);
    };

    ws.onclose = (event) => {
        if (event.wasClean) {
            console.log(`Connection closed cleanly, code=${event.code}, reason=${event.reason}`);
        } else {
            console.error('Connection died');
        }
    };

    ws.onerror = (error) => {
        console.error(`WebSocket error: ${error}`);
    };
}

window.onload = () => {
    if (window.location.pathname === '/profil') {
        connectWebSocket();
    }
};