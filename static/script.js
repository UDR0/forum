document.addEventListener("DOMContentLoaded", function () {
    const burger = document.getElementById("burger");
    const nav = document.getElementById("nav");

    burger.addEventListener("click", function () {
        nav.classList.toggle("active");
        burger.classList.toggle("active");
    });
});
