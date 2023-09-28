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

const onQualityChange = (elem) => {
  const quality = elem.value;
  if (quality === "auto") {
    // remove quality from url
    url.searchParams.delete("q");
  } else {
    url.searchParams.set("q", quality);
  }
  const playElems = document.getElementsByClassName("btn btn-outline btn-info btn-md");
  for (let i = 0; i < playElems.length; i++) {
    const elem = playElems[i];
    const href = elem.getAttribute("href");
    elem.setAttribute("href", href.split("?")[0] + url.search);
  }
  console.log(playElems);
};

const scrollToTop = () => {
  // make smooth scroll to top
  window.scrollTo({
    top: 0,
    behavior: "smooth",
  });
};
