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

