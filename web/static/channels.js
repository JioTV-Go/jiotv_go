const languageElement = document.getElementById("portexe-language-select");
const categoryElement = document.getElementById("portexe-category-select");
const catLangApplyButton = document.getElementById("portexe-search-button");
const qualityElement = document.getElementById("portexe-quality-select");

// Function to handle multi-select display behavior
function toggleMultiSelectDisplay(selectElement) {
  // No need to change size anymore, just update tooltip
  updateMultiSelectDisplay(selectElement);
}

// Function to handle touch-friendly multi-select behavior
function handleTouchFriendlySelect(selectElement) {
  // Use change event to handle "All" option logic without preventing dropdown opening
  selectElement.addEventListener('change', function(event) {
    const selectedOptions = Array.from(this.selectedOptions);
    const allOption = Array.from(this.options).find(opt => opt.value === "0");
    
    // Check if "All" option was just selected
    const allOptionSelected = allOption && allOption.selected;
    const nonAllOptionsSelected = selectedOptions.filter(opt => opt.value !== "0");
    
    if (allOptionSelected && nonAllOptionsSelected.length > 0) {
      // If "All" is selected along with other options, deselect everything else and keep only "All"
      Array.from(this.options).forEach(opt => {
        opt.selected = opt.value === "0";
      });
    } else if (nonAllOptionsSelected.length > 0 && allOptionSelected) {
      // If non-"All" options are selected and "All" is also selected, deselect "All"
      if (allOption) {
        allOption.selected = false;
      }
    }
    
    // Update display
    toggleMultiSelectDisplay(this);
  });
}

// Function to update the visual representation of multi-select
function updateMultiSelectDisplay(selectElement) {
  const selectedOptions = Array.from(selectElement.selectedOptions);
  const selectedCount = selectedOptions.length;
  
  if (selectedCount > 1) {
    // Create a visual indicator of multiple selections
    const firstSelectedText = selectedOptions[0].textContent;
    const displayText = selectedCount > 1 ? `${firstSelectedText} (+${selectedCount - 1} more)` : firstSelectedText;
    
    // Update the visual appearance by adjusting the title attribute
    selectElement.title = selectedOptions.map(opt => opt.textContent).join(', ');
  } else if (selectedCount === 1) {
    selectElement.title = selectedOptions[0].textContent;
  } else {
    selectElement.title = '';
  }
}

catLangApplyButton.addEventListener("click", () => {
  // apply to current url as query params and reload
  const url = new URL(window.location.href);
  
  // Handle multiple selected languages
  const selectedLanguages = Array.from(languageElement.selectedOptions).map(option => option.value);
  if (selectedLanguages.length > 0 && !selectedLanguages.includes("0")) {
    url.searchParams.set("language", selectedLanguages.join(","));
  } else {
    url.searchParams.delete("language");
  }
  
  // Handle multiple selected categories
  const selectedCategories = Array.from(categoryElement.selectedOptions).map(option => option.value);
  if (selectedCategories.length > 0 && !selectedCategories.includes("0")) {
    url.searchParams.set("category", selectedCategories.join(","));
  } else {
    url.searchParams.delete("category");
  }
  
  url.searchParams.set("q", qualityElement.value);

  // reload
  document.location.href = url.href;
});

// on page load, if either language or category is present in query params, set the value of the select element
const url = new URL(window.location.href);
const language = url.searchParams.get("language");
const category = url.searchParams.get("category");

if (language) {
  const languageValues = language.split(",");
  Array.from(languageElement.options).forEach(option => {
    if (languageValues.includes(option.value)) {
      option.selected = true;
    }
  });
  updateMultiSelectDisplay(languageElement);
}

if (category) {
  const categoryValues = category.split(",");
  Array.from(categoryElement.options).forEach(option => {
    if (categoryValues.includes(option.value)) {
      option.selected = true;
    }
  });
  updateMultiSelectDisplay(categoryElement);
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

// Initialize touch-friendly multi-select behavior on page load
document.addEventListener('DOMContentLoaded', () => {
  // Set up touch-friendly multi-select for both dropdowns
  handleTouchFriendlySelect(languageElement);
  handleTouchFriendlySelect(categoryElement);
  
  updateFavoriteButtonStates();
  displayFavoriteChannels(); 
});
