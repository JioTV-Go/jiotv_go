const elements = safeGetElementsById([
  "portexe-search-button",
  "portexe-quality-select"
]);

const {
  "portexe-search-button": catLangApplyButton,
  "portexe-quality-select": qualityElement
} = elements;

catLangApplyButton.addEventListener("click", () => {
  const selectedCategories = [];
  const categoryCheckboxes = document.querySelectorAll(".category-checkbox");
  categoryCheckboxes.forEach(cb => {
    if (cb.checked) selectedCategories.push(cb.value);
  });

  const selectedLanguages = [];
  const languageCheckboxes = document.querySelectorAll(".language-checkbox");
  languageCheckboxes.forEach(cb => {
    if (cb.checked) selectedLanguages.push(cb.value);
  });

  // If "All Categories" (value "0") is checked, or if all/none are checked, clear the filter parameter
  const isAllCategoriesChecked = selectedCategories.includes("0") || selectedCategories.length === 0 || selectedCategories.length === categoryCheckboxes.length;
  const categoryParam = isAllCategoriesChecked ? "" : selectedCategories.filter(val => val !== "0").join(",");

  // If "All Languages" (value "0") is checked, or if all/none are checked, clear the filter parameter
  const isAllLanguagesChecked = selectedLanguages.includes("0") || selectedLanguages.length === 0 || selectedLanguages.length === languageCheckboxes.length;
  const languageParam = isAllLanguagesChecked ? "" : selectedLanguages.filter(val => val !== "0").join(",");

  // Apply URL parameters and reload
  updateUrlParameters({
    language: languageParam,
    category: categoryParam,
    q: qualityElement.value
  });

  // Reload the page
  document.location.href = window.location.href;
});

// On page load, set values from URL parameters
const urlParams = getCurrentUrlParams();
const language = urlParams.get("language");
const category = urlParams.get("category");

if (language) {
  const langs = language.split(",");
  document.querySelectorAll(".language-checkbox").forEach(cb => {
    if (cb.value === "0") {
      cb.checked = false;
    } else {
      cb.checked = langs.includes(cb.value);
    }
  });
} else {
  document.querySelectorAll(".language-checkbox").forEach(cb => {
    cb.checked = (cb.value === "0");
  });
}

if (category) {
  const cats = category.split(",");
  document.querySelectorAll(".category-checkbox").forEach(cb => {
    if (cb.value === "0") {
      cb.checked = false;
    } else {
      cb.checked = cats.includes(cb.value);
    }
  });
} else {
  document.querySelectorAll(".category-checkbox").forEach(cb => {
    cb.checked = (cb.value === "0");
  });
}

// Setup Select All toggle behavior
const setupSelectAll = (checkboxClass, allValue) => {
  const checkboxes = document.querySelectorAll(checkboxClass);
  const allCheckbox = Array.from(checkboxes).find(cb => cb.value === allValue);
  if (!allCheckbox) return;

  const otherCheckboxes = Array.from(checkboxes).filter(cb => cb.value !== allValue);

  allCheckbox.addEventListener("change", () => {
    if (allCheckbox.checked) {
      otherCheckboxes.forEach(cb => {
        cb.checked = false;
      });
    } else {
      const anyChecked = otherCheckboxes.some(item => item.checked);
      if (!anyChecked) {
        allCheckbox.checked = true;
      }
    }
  });

  otherCheckboxes.forEach(cb => {
    cb.addEventListener("change", () => {
      if (cb.checked) {
        allCheckbox.checked = false;
      } else {
        const anyChecked = otherCheckboxes.some(item => item.checked);
        if (!anyChecked) {
          allCheckbox.checked = true;
        }
      }
    });
  });
};

document.addEventListener('DOMContentLoaded', () => {
  // Run select-all wiring first so the category/language checkboxes behave
  // correctly even if favorites code below ever throws.
  setupSelectAll(".category-checkbox", "0");
  setupSelectAll(".language-checkbox", "0");
  updateFavoriteButtonStates();
  displayFavoriteChannels();
});

const onQualityChange = (elem) => {
  const quality = elem.value;
  
  if (quality === "auto") {
    updateUrlParameter("q", "");
    removeLocalStorageItem("quality");
  } else {
    updateUrlParameter("q", quality);
    setLocalStorageItem("quality", quality);
  }
  
  // Update all channel card href attributes with new query parameter.
  // Only target channel cards (a[href]); dropdown-content panels also carry
  // the .card class but are <div>s with no href, and touching them crashes.
  const currentParams = getCurrentUrlParams();
  document.querySelectorAll("a.card[data-channel-id]").forEach((cardElem) => {
    const href = cardElem.getAttribute("href");
    if (href) cardElem.setAttribute("href", href.split("?")[0] + "?" + currentParams.toString());
  });
};

const storedQuality = getLocalStorageItem("quality");
if (storedQuality && qualityElement) {
  qualityElement.value = storedQuality;
}

const urlParams2 = getCurrentUrlParams();
if (urlParams2.get("q") && qualityElement) {
  qualityElement.value = urlParams2.get("q");
  onQualityChange(qualityElement); 
}


const scrollToTop = () => {
  window.scrollTo({
    top: 0,
    behavior: "smooth",
  });
};

// Favorite Channels Functionality
const FAVORITES_STORAGE_KEY = "favoriteChannels";

function getFavoriteChannels() {
  return getLocalStorageItem(FAVORITES_STORAGE_KEY, []);
}

function saveFavoriteChannels(favoriteIds) {
  setLocalStorageItem(FAVORITES_STORAGE_KEY, favoriteIds);
}

function displayFavoriteChannels() {
  const favoriteIds = getFavoriteChannels();
  const elements = safeGetElementsById([
    "favorite-channels-section",
    "favorite-channels-container", 
    "original-channels-grid"
  ]);
  
  const { 
    "favorite-channels-section": favoriteChannelsSection,
    "favorite-channels-container": favoriteChannelsContainer,
    "original-channels-grid": originalChannelsGrid 
  } = elements;

  if (!favoriteChannelsSection || !favoriteChannelsContainer || !originalChannelsGrid) {
    console.error("One or more channel container elements not found.");
    return;
  }

  // Move all cards to a temporary fragment to prevent issues with live collections
  // or ensure they are detached before re-appending.
  // However, a simpler approach for now is to just re-append.
  // This might cause a brief flicker for a large number of cards.
  
  // Clear favorite container before potentially hiding it or re-populating
  // while (favoriteChannelsContainer.firstChild) {
  //   favoriteChannelsContainer.removeChild(favoriteChannelsContainer.firstChild);
  // }
  // The logic below of appending will move them, so explicit clearing is not strictly necessary
  // if we iterate over ALL cards and move them to correct container.

  if (favoriteIds.length > 0) {
    favoriteChannelsSection.style.display = 'block'; // Or 'flex' or 'grid' depending on layout
  } else {
    favoriteChannelsSection.style.display = 'none';
  }

  const allChannelCards = document.querySelectorAll('a.card[data-channel-id]');

  // Create DocumentFragments to batch DOM updates
  const favoriteFragment = document.createDocumentFragment();
  const originalFragment = document.createDocumentFragment();

  allChannelCards.forEach(card => {
    const cardChannelId = card.dataset.channelId;
    if (favoriteIds.includes(cardChannelId)) {
      favoriteFragment.appendChild(card);
    } else {
      originalFragment.appendChild(card);
    }
  });

  // Append fragments to their respective containers
  favoriteChannelsContainer.appendChild(favoriteFragment);
  originalChannelsGrid.appendChild(originalFragment);
}

function toggleFavorite(channelId) {
  const favoriteIds = getFavoriteChannels();
  const index = favoriteIds.indexOf(channelId);

  if (index > -1) { // Channel was a favorite, removing it
    favoriteIds.splice(index, 1);
    updateFavoriteButtonState(channelId, false);
  } else { // Channel was not a favorite, adding it
    favoriteIds.push(channelId);
    updateFavoriteButtonState(channelId, true);
  }
  
  saveFavoriteChannels(favoriteIds);
  displayFavoriteChannels(); // Refresh the channel lists
}

function updateFavoriteButtonStates() {
  const favoriteIds = getFavoriteChannels();
  const favoriteButtons = document.querySelectorAll(".favorite-btn");

  favoriteButtons.forEach(button => {
    const channelId = button.id.replace("favorite-btn-", "");
    updateFavoriteButtonState(channelId, favoriteIds.includes(channelId));
  });
}
