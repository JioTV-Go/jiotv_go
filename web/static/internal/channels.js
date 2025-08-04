const languageElement = document.getElementById("portexe-language-select");
const categoryElement = document.getElementById("portexe-category-select");
const catLangApplyButton = document.getElementById("portexe-search-button");
const qualityElement = document.getElementById("portexe-quality-select");

catLangApplyButton.addEventListener("click", () => {
  // apply to current url as query params and reload
  const url = new URL(window.location.href);
  url.searchParams.set("language", languageElement.value);
  url.searchParams.set("category", categoryElement.value);
  url.searchParams.set("q", qualityElement.value);

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
  const currentUrl = new URL(window.location.href); // Use a fresh URL object
  if (quality === "auto") {
    // remove quality from url
    currentUrl.searchParams.delete("q");
    // remove quality from local storage
    localStorage.removeItem("quality");
  } else {
    // set quality in url
    currentUrl.searchParams.set("q", quality);
    // set quality in local storage
    localStorage.setItem("quality", quality);
  }
  history.pushState({}, "", currentUrl.href); // Update history with the modified URL
  const playElems = document.getElementsByClassName("card");
  for (let i = 0; i < playElems.length; i++) {
    const cardElem = playElems[i]; // Renamed to avoid confusion with the 'elem' parameter
    const href = cardElem.getAttribute("href");
    cardElem.setAttribute("href", href.split("?")[0] + currentUrl.search);
  }
};

const storedQuality = localStorage.getItem("quality"); // Renamed to avoid conflict
if (storedQuality) {
  qualityElement.value = storedQuality;
}

if (url.searchParams.get("q")) {
  qualityElement.value = url.searchParams.get("q");
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

export function getFavoriteChannels() {
  const storedFavorites = localStorage.getItem(FAVORITES_STORAGE_KEY);
  if (!storedFavorites) {
    return [];
  }
  try {
    const parsedFavorites = JSON.parse(storedFavorites);
    return Array.isArray(parsedFavorites) ? parsedFavorites : [];
  } catch (e) {
    console.error("Error parsing favorite channels from localStorage:", e);
    return [];
  }
}

export function saveFavoriteChannels(favoriteIds) {
  localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(favoriteIds));
}

export function displayFavoriteChannels() {
  const favoriteIds = getFavoriteChannels();
  const favoriteChannelsSection = document.getElementById("favorite-channels-section");
  const favoriteChannelsContainer = document.getElementById("favorite-channels-container");
  const originalChannelsGrid = document.getElementById("original-channels-grid");

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

export function toggleFavorite(channelId) {
  const favoriteIds = getFavoriteChannels();
  const button = document.getElementById(`favorite-btn-${channelId}`);
  const starIcon = document.getElementById(`star-icon-${channelId}`);
  const xIcon = document.getElementById(`x-icon-${channelId}`);
  const index = favoriteIds.indexOf(channelId);

  if (index > -1) { // Channel was a favorite, removing it
    favoriteIds.splice(index, 1);
    if (button) {
      button.classList.remove("favorited"); // Existing class toggle
      if (starIcon && xIcon) {
        starIcon.classList.remove('hidden');
        xIcon.classList.add('hidden');
      }
    }
  } else { // Channel was not a favorite, adding it
    favoriteIds.push(channelId);
    if (button) {
      button.classList.add("favorited"); // Existing class toggle
      if (starIcon && xIcon) {
        starIcon.classList.add('hidden');
        xIcon.classList.remove('hidden');
      }
    }
  }
  saveFavoriteChannels(favoriteIds);
  displayFavoriteChannels(); // Refresh the channel lists
}

export function updateFavoriteButtonStates() {
  const favoriteIds = getFavoriteChannels();
  const favoriteButtons = document.querySelectorAll(".favorite-btn");

  favoriteButtons.forEach(button => {
    const channelId = button.id.replace("favorite-btn-", "");
    const starIcon = document.getElementById(`star-icon-${channelId}`);
    const xIcon = document.getElementById(`x-icon-${channelId}`);

    if (starIcon && xIcon) { // Ensure icons exist
      if (favoriteIds.includes(channelId)) {
        button.classList.add("favorited"); // Existing class toggle
        starIcon.classList.add('hidden');
        xIcon.classList.remove('hidden');
      } else {
        button.classList.remove("favorited"); // Existing class toggle
        starIcon.classList.remove('hidden');
        xIcon.classList.add('hidden');
      }
    }
  });
}

document.addEventListener('DOMContentLoaded', () => {
  updateFavoriteButtonStates();
  displayFavoriteChannels(); 
});
