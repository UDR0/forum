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
        event.stopPropagation();
        event.preventDefault();
        const img = this.querySelector('.destination-coeur');
        const regionName = this.getAttribute('data-region'); // Get region name
        
        // Determine liked status
        const liked = !img.src.includes('coeur_rouge.png'); // True if red heart is being added
        
        // Toggle heart icon
        if (liked) {
            img.src = 'static/img/coeur_rouge.png'; // Change to red heart
            console.log(`Liked region: ${regionName}`);
        } else {
            img.src = 'static/img/coeur.png'; // Change back to normal heart
            console.log(`Unliked region: ${regionName}`);
        }

        // Send data to the server
        fetch('/like', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ region: regionName, liked }),
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

/*
document.querySelectorAll('.destination-coeur-container').forEach(container => {
    container.addEventListener('click', function () {
        const img = this.querySelector('.destination-coeur');
        const regionName = this.getAttribute('data-region'); // Get region name
        
        // Determine liked status
        const liked = !img.src.includes('coeur_rouge.png'); // True if red heart is being added
        
        // Toggle heart icon
        if (liked) {
            img.src = 'static/img/coeur_rouge.png'; // Change to red heart
            console.log(`Liked region: ${regionName}`);
        } else {
            img.src = 'static/img/coeur.png'; // Change back to normal heart
            console.log(`Unliked region: ${regionName}`);
        }

        // Send data to the server
        fetch('/like', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ region: regionName, liked }),
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
    const avatars = document.querySelectorAll(".imgAvatar img"); // Select all avatar images
    const photoProfil = document.getElementById("photoProfil"); // Profile photo element
    const photoUrlElement = document.getElementById("photo_url"); // Hidden input for avatar URL

    avatars.forEach(avatar => {
        avatar.addEventListener("click", function () {
            // Update the profile photo
            photoProfil.src = this.src;

            // Update the hidden input value
            if (photoUrlElement) {
                photoUrlElement.value = this.src;
                console.log(`photo_url updated: ${photoUrlElement.value}`);
            } else {
                console.error("Element 'photo_url' not found!");
            }

            // Close the pop-up after avatar selection
            closePopupProfil();
        });
    });
});
*/

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

function PopupFils() {
    const popupFils = document.getElementById("popupAjouterFil");
    const imageBtnFils = document.getElementById("btnAjouterFil");

    if (popupFils.style.display === "flex") {
        document.getElementById("popupAjouterFil").style.display = "none";
        document.getElementById("btnAjouterFil").src = "static/img/ajouter.png";
    } else {
        document.getElementById("popupAjouterFil").style.display = "flex";
        document.getElementById("btnAjouterFil").src = "static/img/moin.png";
    }
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


document.addEventListener("DOMContentLoaded", function () {
    const searchBar = document.getElementById('searchBar');
    const searchButton = document.querySelector('.search-icon');

    searchBar.addEventListener('keydown',(event) => {
        if (event.key === 'Enter') {

        redirectToRegion();
        }

        
    });

    searchButton.addEventListener('click', () => {
        redirectToRegion();
    });
});

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