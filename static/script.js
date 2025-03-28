document.addEventListener("DOMContentLoaded", function() {
    // Sélectionner les éléments
    const burgerMenu = document.getElementById("burger"); // Le bouton du menu burger
    const navMenu = document.querySelector(".meta-ul"); // Le menu de navigation
    
    // Écouter les clics sur le bouton burger
    if (burgerMenu && navMenu) {
        burgerMenu.addEventListener("click", () => {
            navMenu.classList.toggle("active"); // Ajouter/retirer la classe 'active'
        });
    }

    document.querySelectorAll('.coeur-container').forEach(coeur => {
        coeur.addEventListener('click', function () {
            const img = this.querySelector('.coeur');
            if (img) {
                if (img.src.includes('coeur.png')) {
                    img.src = 'static/img/coeur_rouge.png'; // Remplace par le cœur rouge
                } else {
                    img.src = 'static/img/coeur.png'; // Reviens au cœur normal
                }
            }
        });
    });

    // Sélectionne toutes les images du popup
    const avatars = document.querySelectorAll(".imgAvatar img");
    // Sélectionne l'image de profil
    const photoProfil = document.getElementById("photoProfil");
    // Champ caché pour l'URL de l'avatar
    const photoUrlInput = document.getElementById("photo_url");

    avatars.forEach(avatar => {
        avatar.addEventListener("click", function() {
            const popup = document.getElementById("popup");
            if (popup) {
                popup.style.display = "none";
            }
            if (photoProfil && photoUrlInput) {
                // Remplace la source de l'image de profil par celle de l'avatar cliqué
                photoProfil.src = this.src;
                // Met à jour le champ caché avec l'URL de l'avatar choisi
                photoUrlInput.value = this.src;
            }
        });
    });

    document.addEventListener("click", function(event) {
        const popup = document.getElementById("popup");
        const openPopupBtn = document.getElementById("photoProfil");
        if (popup && openPopupBtn && popup.style.display === "block" &&
            !popup.contains(event.target) &&
            event.target !== openPopupBtn
        ) {
            popup.style.display = "none";
        }
    });

    function openPopup() {
        const popup = document.getElementById("popup");
        if (popup) {
            popup.style.display = "block";
            event.stopPropagation(); // Empêche la propagation du clic pour éviter une fermeture immédiate
        }
    }

    // Ajoute un écouteur de clic sur le bouton pour ouvrir la popup
    if (photoProfil) {
        photoProfil.addEventListener("click", function(event) {
            openPopup();
        });
    }

    // ------------------- Profil ---------------------------//

    function openPopupProfil() {
        const popupProfil = document.getElementById("popupProfil");
        const floue = document.getElementById("floue");
        if (popupProfil && floue) {
            popupProfil.style.display = "block";
            floue.style.display = "block";
            event.stopPropagation(); // Empêche la propagation du clic pour éviter une fermeture immédiate
        }
    }

    function closePopupProfil() {
        const popupProfil = document.getElementById("popupProfil");
        const floue = document.getElementById("floue");
        if (popupProfil && floue) {
            popupProfil.style.display = "none";
            floue.style.display = "none";
            event.stopPropagation();
        }
    }

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
});