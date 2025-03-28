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

// ------------------- SEARCH ---------------------------//

document.getElementById('search-input').addEventListener('input', function () {
    
    const query = this.value;  // recupère la valeur de la requete
    const autocomBox = document.getElementById('autocom-box'); // recupère la boite de suggestions avec l'id 'autocom-box'
    
    
    if (query.length < 2) { // permet de donner les suggestions seulement si la requete a plus de 2 lettres 
        autocomBox.innerHTML = ''; 
        autocomBox.classList.remove('active'); 
        return; 
    }

    
    fetch(`/search-suggestions?q=${encodeURIComponent(query)}`) // Encode la requête pour l'utiliser dans l'URL
        .then(response => response.json()) // Convertit la réponse en JSON
        .then(data => { 
            autocomBox.innerHTML = ''; // Réinitialise la boite de suggestions avant d'ajouter les nouvelles
            
            data.forEach(suggestion => {
                // Crée un nouvel élément <li> pour chaque suggestion
                const li = document.createElement('li'); 
                // Met les texte des options sous la forme de => "Department, Region"
                li.textContent = `${suggestion.department_name}, ${suggestion.region_name}`;
                
                li.addEventListener('click', () => {
                    // quand une options est choisie
                    document.getElementById('search-input').value = li.textContent; // met l'option choisie dans la bar de recherche
                    autocomBox.innerHTML = ''; // efface toutes les options après qu'une d'entre elles est choisie
                    autocomBox.classList.remove('active'); 
                });
                
                // ajoute a la liste des options
                autocomBox.appendChild(li);
            });

            // rend visible les options
            autocomBox.classList.add('active');
        })
        
        .catch(error => console.error('Error fetching suggestions:', error)); // message d'erreur dans la console en cas d'erreur 
});