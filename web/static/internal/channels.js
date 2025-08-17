const elements = safeGetElementsById([
  "portexe-language-select",
  "portexe-category-select", 
  "portexe-search-button",
  "portexe-quality-select"
]);

const {
  "portexe-language-select": languageElement,
  "portexe-category-select": categoryElement,
  "portexe-search-button": catLangApplyButton,
  "portexe-quality-select": qualityElement
} = elements;

catLangApplyButton.addEventListener("click", () => {
  // Apply URL parameters and reload
  updateUrlParameters({
    language: languageElement.value,
    category: categoryElement.value,
    q: qualityElement.value
  });

  // Reload the page
  document.location.href = window.location.href;
});

// On page load, set values from URL parameters
const urlParams = getCurrentUrlParams();
const language = urlParams.get("language");
const category = urlParams.get("category");

if (language && languageElement) {
  languageElement.value = language;
}

if (category && categoryElement) {
  categoryElement.value = category;
}

const onQualityChange = (elem) => {
  const quality = elem.value;
  
  if (quality === "auto") {
    updateUrlParameter("q", "");
    removeLocalStorageItem("quality");
  } else {
    updateUrlParameter("q", quality);
    setLocalStorageItem("quality", quality);
  }
  
  // Update all card href attributes with new query parameter
  const playElems = document.getElementsByClassName("card");
  const currentParams = getCurrentUrlParams();
  
  for (let i = 0; i < playElems.length; i++) {
    const cardElem = playElems[i];
    const href = cardElem.getAttribute("href");
    cardElem.setAttribute("href", href.split("?")[0] + "?" + currentParams.toString());
  }
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

document.addEventListener('DOMContentLoaded', () => {
  updateFavoriteButtonStates();
  displayFavoriteChannels(); 
});
