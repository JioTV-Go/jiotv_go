// Function to get favorite channels from localStorage
function getFavoriteChannels() {
  const favorites = localStorage.getItem('favoriteChannels');
  if (favorites) {
    try {
      const parsedFavorites = JSON.parse(favorites);
      // Ensure it's an array; otherwise, return an empty array
      return Array.isArray(parsedFavorites) ? parsedFavorites : [];
    } catch (e) {
      console.error("Error parsing favoriteChannels from localStorage:", e);
      return [];
    }
  }
  return [];
}

// Function to save favorite channels to localStorage
function saveFavoriteChannels(channelIdsArray) {
  if (!Array.isArray(channelIdsArray)) {
    console.error("saveFavoriteChannels expects an array.");
    return;
  }
  localStorage.setItem('favoriteChannels', JSON.stringify(channelIdsArray));
}

// Function to check if a channel is a favorite
function isChannelFavorite(channelId) {
  const favorites = getFavoriteChannels();
  // Ensure comparison is consistent (e.g., by converting channelId to string)
  return favorites.includes(String(channelId));
}

// Function to add a channel to favorites
function addFavoriteChannel(channelId) {
  let favorites = getFavoriteChannels();
  const channelIdStr = String(channelId);
  if (!favorites.includes(channelIdStr)) {
    favorites.push(channelIdStr);
    saveFavoriteChannels(favorites);
  }
}

// Function to remove a channel from favorites
function removeFavoriteChannel(channelId) {
  let favorites = getFavoriteChannels();
  const channelIdStr = String(channelId);
  favorites = favorites.filter(id => id !== channelIdStr);
  saveFavoriteChannels(favorites);
}

// Function to toggle the favorite status of a channel
function toggleFavoriteStatus(channelId) {
  const channelIdStr = String(channelId);
  if (isChannelFavorite(channelIdStr)) {
    removeFavoriteChannel(channelIdStr);
    return false; // Channel was removed
  } else {
    addFavoriteChannel(channelIdStr);
    return true; // Channel was added
  }
}

// Function to update the visual state of a favorite button
function updateFavoriteButtonVisualState(buttonElement, isFavorite) {
  if (!buttonElement) return;
  const svgIcon = buttonElement.querySelector('svg');
  if (!svgIcon) return;

  if (isFavorite) {
    svgIcon.setAttribute('fill', 'currentColor');
    // Ensure the class that provides the yellow color is present
    svgIcon.classList.add('text-yellow-400'); 
  } else {
    svgIcon.setAttribute('fill', 'none');
    // We can keep text-yellow-400 for the stroke, or manage it if an unfilled state needs a different stroke color
  }
}

// Function to set up favorite buttons: initial state and event listeners
function setupFavoriteButtons() {
  const buttons = document.querySelectorAll('.favorite-btn');
  buttons.forEach(button => {
    const channelId = button.dataset.channelId;
    if (!channelId) return;

    let isFav = isChannelFavorite(channelId);
    updateFavoriteButtonVisualState(button, isFav);

    button.addEventListener('click', function(event) {
      event.preventDefault();
      event.stopPropagation();
      
      const currentChannelId = this.dataset.channelId; // 'this' refers to the button
      if (!currentChannelId) return;

      const newFavStatus = toggleFavoriteStatus(currentChannelId);
      updateFavoriteButtonVisualState(this, newFavStatus);

      // Optional: Dispatch a custom event if other parts of the app need to react
      // document.dispatchEvent(new CustomEvent('favoritesUpdated', { detail: { channelId: currentChannelId, isFavorite: newFavStatus } }));
    });
  });
}

// Initialize favorite buttons when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', setupFavoriteButtons);

// Function to filter and display only favorite channels
function filterChannelsByFavorites() {
  const channelCards = document.querySelectorAll('.card[data-channel-id]');
  channelCards.forEach(card => {
    const channelId = card.dataset.channelId;
    if (isChannelFavorite(channelId)) {
      card.style.display = 'block'; // Or appropriate display style for grid
    } else {
      card.style.display = 'none';
    }
  });
  if (typeof updateFocusableChannels === 'function' && typeof setNewFocus === 'function') {
    updateFocusableChannels();
    setNewFocus(0);
  }
}

// Function to show all channels
function showAllChannels() {
  const channelCards = document.querySelectorAll('.card[data-channel-id]');
  channelCards.forEach(card => {
    card.style.display = 'block'; // Or appropriate display style for grid
  });
  if (typeof updateFocusableChannels === 'function' && typeof setNewFocus === 'function') {
    updateFocusableChannels();
    setNewFocus(0);
  }
  const favoritesTab = document.getElementById('favorites-tab');
  if (favoritesTab) {
    favoritesTab.classList.remove('btn-active'); // Example class for active state
  }
}

// Add event listener for the Favorites tab
document.addEventListener('DOMContentLoaded', () => {
  // Ensure setupFavoriteButtons runs first, then these listeners are added.
  // Alternatively, can combine DOMContentLoaded listeners or ensure order if dependencies exist.
  // For now, assuming setupFavoriteButtons has no direct dependency on these tab clicks.

  const favoritesTab = document.getElementById('favorites-tab');
  if (favoritesTab) {
    favoritesTab.addEventListener('click', function(event) {
      event.preventDefault();
      filterChannelsByFavorites();
      this.classList.add('btn-active'); // Example class for active state
      // Potentially remove 'btn-active' from other tabs if needed
    });
  }

  const homeTab = document.getElementById('home-tab');
  if (homeTab) {
    // No preventDefault needed if it's a real navigation link,
    // but if it were to be handled purely client-side in the future:
    // homeTab.addEventListener('click', function(event) {
    //   event.preventDefault(); 
    //   showAllChannels();
    // });
    // For now, just ensure showAllChannels is called to reset view if it was filtered.
    // The page reload will also handle this, but this makes it explicit for client-side state.
    homeTab.addEventListener('click', function() {
        // If the home button is clicked, we want to ensure the channel list is reset.
        // showAllChannels() will remove the 'btn-active' from favorites and display all cards.
        // This is beneficial if we ever move to a single-page application model more heavily.
        showAllChannels(); 
    });
  }
});
