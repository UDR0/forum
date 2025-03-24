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


// Ecouter les evenements sur le bputton pour choisir sa PP
document.getElementById("uploadBtn").addEventListener("click", function() {
    document.getElementById("fileInput").click();
});

// Remplace la PP
document.getElementById("fileInput").addEventListener("change", function(event) {
    const file = event.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            document.getElementById("photoProfil").src = e.target.result;
        };
        reader.readAsDataURL(file);
    }
});
