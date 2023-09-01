// const languageOptions = [
//     { value: 1, label: "Hindi" },
//     { value: 2, label: "Marathi" },
//     { value: 3, label: "Punjabi" },
//     { value: 4, label: "Urdu" },
//     { value: 5, label: "Bengali" },
//     { value: 6, label: "English" },
//     { value: 7, label: "Malayalam" },
//     { value: 8, label: "Tamil" },
//     { value: 9, label: "Gujarati" },
//     { value: 10, label: "Odia" },
//     { value: 11, label: "Telugu" },
//     { value: 12, label: "Bhojpuri" },
//     { value: 13, label: "Kannada" },
//     { value: 14, label: "Assamese" },
//     { value: 15, label: "Nepali" },
//     { value: 16, label: "French" },
//     { value: 18, label: "Other" }
// ];

// const categoryOptions = [
//     { value: 5, label: "Entertainment" },
//     { value: 6, label: "Movies" },
//     { value: 7, label: "Kids" },
//     { value: 8, label: "Sports" },
//     { value: 9, label: "Lifestyle" },
//     { value: 10, label: "Infotainment" },
//     { value: 12, label: "News" },
//     { value: 13, label: "Music" },
//     { value: 15, label: "Devotional" },
//     { value: 16, label: "Business" },
//     { value: 17, label: "Educational" },
//     { value: 18, label: "Shopping" },
//     { value: 19, label: "JioDarshan" }
// ];

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
