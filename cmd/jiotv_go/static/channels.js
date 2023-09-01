const languageElement = document.getElementById("portexe-language-select");
const categoryElement = document.getElementById("portexe-category-select");
const catLangApplyButton = document.getElementById("portexe-search-button");

catLangApplyButton.addEventListener("click", () => {
    // apply to current url as query params and reload
    const url = new URL(window.location.href);
    url.searchParams.set("language", languageElement.value);
    url.searchParams.set("category", categoryElement.value);

    // reload
    document.location.href = url.href;
});


// on page load, if either language or category is present in query params, set the value of the select element
const url = new URL(window.location.href);
const language = url.searchParams.get("language");
const category = url.searchParams.get("category");

if (language) {
    languageElement.value = language;
}

if (category) {
    categoryElement.value = category;
}

const scrollToTop = () => {
    // make smooth scroll to top
    window.scrollTo({
        top: 0,
        behavior: "smooth"
    });
};
