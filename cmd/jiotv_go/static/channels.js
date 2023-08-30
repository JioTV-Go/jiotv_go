const languageOptions = [
    { value: 1, label: "Hindi" },
    { value: 2, label: "Marathi" },
    { value: 3, label: "Punjabi" },
    { value: 4, label: "Urdu" },
    { value: 5, label: "Bengali" },
    { value: 6, label: "English" },
    { value: 7, label: "Malayalam" },
    { value: 8, label: "Tamil" },
    { value: 9, label: "Gujarati" },
    { value: 10, label: "Odia" },
    { value: 11, label: "Telugu" },
    { value: 12, label: "Bhojpuri" },
    { value: 13, label: "Kannada" },
    { value: 14, label: "Assamese" },
    { value: 15, label: "Nepali" },
    { value: 16, label: "French" },
    { value: 18, label: "Other" }
];

// Create select element
const languageElement = document.createElement("select");
languageElement.classList.add("select", "select-primary", "select-sm", "sm:select-md", "w-full", "max-w-auto", "sm:max-w-xs", "sm:w-auto", "rounded-xl");

// add empty first disabled option
const languageDummyOption = document.createElement("option");
languageDummyOption.value = "0";
languageDummyOption.textContent = "Select Language";
languageDummyOption.selected = true;
languageElement.appendChild(languageDummyOption);

// Create and add option elements
languageOptions.forEach(option => {
    const optionElement = document.createElement("option");
    optionElement.value = option.value;
    optionElement.textContent = option.label;
    languageElement.appendChild(optionElement);
});

// Append select element to a container (e.g., a div with an id of "language-select-container")
const container = document.getElementById("portexe-search-root");
container.appendChild(languageElement);

const categoryOptions = [
    { value: 5, label: "Entertainment" },
    { value: 6, label: "Movies" },
    { value: 7, label: "Kids" },
    { value: 8, label: "Sports" },
    { value: 9, label: "Lifestyle" },
    { value: 10, label: "Infotainment" },
    { value: 12, label: "News" },
    { value: 13, label: "Music" },
    { value: 15, label: "Devotional" },
    { value: 16, label: "Business" },
    { value: 17, label: "Educational" },
    { value: 18, label: "Shopping" },
    { value: 19, label: "JioDarshan" }
];

// Create select element
const categoryElement = document.createElement("select");
categoryElement.classList.add("select", "select-primary", "select-sm", "sm:select-md","w-full", "max-w-auto", "sm:max-w-xs", "sm:w-auto", "rounded-xl");

// add empty first disabled option
const categoryDummyOption = document.createElement("option");
categoryDummyOption.value = "0";
categoryDummyOption.textContent = "Select Category";
categoryDummyOption.selected = true;
categoryElement.appendChild(categoryDummyOption);

// Create and add option elements
categoryOptions.forEach(option => {
    const optionElement = document.createElement("option");
    optionElement.value = option.value;
    optionElement.textContent = option.label;
    categoryElement.appendChild(optionElement);
});

// Append select element to a container (e.g., a div with an id of "category-select-container")
container.appendChild(categoryElement);

const applyButton = document.createElement("button");
applyButton.classList.add("btn", "btn-primary", "btn-sm", "sm:btn-md", "w-full", "sm:w-auto","rounded-xl");
applyButton.textContent = "Apply";
applyButton.addEventListener("click", () => {
    // apply to current url as query params and reload
    const url = new URL(window.location.href);
    url.searchParams.set("language", languageElement.value);
    url.searchParams.set("category", categoryElement.value);

    // reload
    document.location.href = url.href;
});

container.appendChild(applyButton);

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
