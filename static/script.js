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

document.addEventListener("DOMContentLoaded", function() {
    // Sélectionne toutes les images du popup
    const avatars = document.querySelectorAll(".imgAvatar img");
    // Sélectionne l'image de profil
    const photoProfil = document.getElementById("photoProfil");
    // Champ caché pour l'URL de l'avatar
    const photoUrlInput = document.getElementById("photo_url");

    avatars.forEach(avatar => {
        avatar.addEventListener("click", function() {
            document.getElementById("popup").style.display = "none";
            // Remplace la source de l'image de profil par celle de l'avatar cliqué
            photoProfil.src = this.src;
            // Met à jour le champ caché avec l'URL de l'avatar choisi
            photoUrlInput.value = this.src;
        });
    });
});

document.addEventListener("click", function(event) {
    const popup = document.getElementById("popup");
    const openPopupBtn = document.getElementById("photoProfil");
    if (popup.style.display === "block" &&
        !popup.contains(event.target) &&
        event.target !== openPopupBtn
    ) {
        popup.style.display = "none";
    }
});

function openPopup() {
    document.getElementById("popup").style.display = "block";
    event.stopPropagation(); // Empêche la propagation du clic pour éviter une fermeture immédiate
}

// Ajoute un écouteur de clic sur le bouton pour ouvrir la popup
document.getElementById("photoProfil").addEventListener("click", function(event) {
    openPopup();
});

// ------------------- Profil ---------------------------//

function openPopupProfil() {
    document.getElementById("popupProfil").style.display = "block";
    document.getElementById("floue").style.display = "block";
    event.stopPropagation(); // Empêche la propagation du clic pour éviter une fermeture immédiate
}

function closePopupProfil() {
    document.getElementById("popupProfil").style.display = "none";
    document.getElementById("floue").style.display = "none";
    event.stopPropagation();
}