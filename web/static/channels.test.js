// Mock localStorage
let mockLocalStorageStore = {};

const localStorageMock = {
  getItem: jest.fn((key) => mockLocalStorageStore[key] || null),
  setItem: jest.fn((key, value) => {
    mockLocalStorageStore[key] = value.toString();
  }),
  removeItem: jest.fn((key) => {
    delete mockLocalStorageStore[key];
  }),
  clear: jest.fn(() => {
    mockLocalStorageStore = {};
  }),
};

// Assign the mock to the global window object if running in a JSDOM-like environment
global.localStorage = localStorageMock;

// Mock console.error to avoid cluttering test output
global.console.error = jest.fn();

// --- Start of functions from channels.js (or assume they are imported/available) ---
// For testing purposes, we'll redefine simplified versions or assume they are imported.
// In a real Jest environment, you'd import these from './channels.js'
// and potentially mock dependencies within that file if necessary.

const FAVORITES_STORAGE_KEY = "favoriteChannels";

function getFavoriteChannels() {
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

function saveFavoriteChannels(favoriteIds) {
  localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(favoriteIds));
}

// Mock displayFavoriteChannels for toggleFavorite tests initially
let displayFavoriteChannelsCalled = false;
function displayFavoriteChannels() {
  displayFavoriteChannelsCalled = true; // Simple spy for toggleFavorite

  // Actual DOM manipulation logic for displayFavoriteChannels will be tested separately
  const favoriteIds = getFavoriteChannels();
  const favoriteChannelsSection = document.getElementById("favorite-channels-section");
  const favoriteChannelsContainer = document.getElementById("favorite-channels-container");
  const originalChannelsGrid = document.getElementById("original-channels-grid");

  if (!favoriteChannelsSection || !favoriteChannelsContainer || !originalChannelsGrid) {
    // console.error("One or more channel container elements not found in mock DOM for displayFavoriteChannels.");
    return;
  }

  if (favoriteIds.length > 0) {
    favoriteChannelsSection.style.display = 'block';
  } else {
    favoriteChannelsSection.style.display = 'none';
  }

  const allChannelCards = document.querySelectorAll('a.card[data-channel-id]');
  allChannelCards.forEach(card => {
    const cardChannelId = card.dataset.channelId;
    if (favoriteIds.includes(cardChannelId)) {
      favoriteChannelsContainer.appendChild(card);
    } else {
      originalChannelsGrid.appendChild(card);
    }
  });
}

function toggleFavorite(channelId) {
  const favoriteIds = getFavoriteChannels();
  const button = document.getElementById(`favorite-btn-${channelId}`);
  const index = favoriteIds.indexOf(channelId);

  if (index > -1) {
    favoriteIds.splice(index, 1);
    if (button) {
      button.classList.remove("favorited");
    }
  } else {
    favoriteIds.push(channelId);
    if (button) {
      button.classList.add("favorited");
    }
  }
  saveFavoriteChannels(favoriteIds);
  displayFavoriteChannels(); // Call the actual or spied/mocked function
}

function updateFavoriteButtonStates() {
  const favoriteIds = getFavoriteChannels();
  const favoriteButtons = document.querySelectorAll(".favorite-btn");

  favoriteButtons.forEach(button => {
    const channelId = button.id.replace("favorite-btn-", "");
    if (favoriteIds.includes(channelId)) {
      button.classList.add("favorited");
    } else {
      button.classList.remove("favorited");
    }
  });
}
// --- End of functions from channels.js ---

describe('Favorite Channels Functionality', () => {
  beforeEach(() => {
    localStorageMock.clear();
    jest.clearAllMocks(); // Clears all Jest mocks, including localStorageMock calls and console.error
    displayFavoriteChannelsCalled = false; // Reset spy for toggleFavorite

    // Clear and set up basic DOM structure for each test
    document.body.innerHTML = `
      <div id="favorite-channels-section" style="display: none;">
        <h2>Favorite Channels</h2>
        <div id="favorite-channels-container"></div>
      </div>
      <div id="original-channels-grid"></div>
    `;
  });

  describe('getFavoriteChannels', () => {
    it('should return an empty array when localStorage is empty', () => {
      expect(getFavoriteChannels()).toEqual([]);
    });

    it('should return an empty array for invalid JSON', () => {
      localStorageMock.setItem(FAVORITES_STORAGE_KEY, 'invalid-json');
      expect(getFavoriteChannels()).toEqual([]);
      expect(console.error).toHaveBeenCalled();
    });

    it('should return an empty array if stored value is not an array', () => {
      localStorageMock.setItem(FAVORITES_STORAGE_KEY, JSON.stringify({ not: "an array" }));
      expect(getFavoriteChannels()).toEqual([]);
    });
    
    it('should return the parsed array of IDs for valid JSON', () => {
      const favorites = ['id1', 'id2'];
      localStorageMock.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(favorites));
      expect(getFavoriteChannels()).toEqual(favorites);
    });
  });

  describe('saveFavoriteChannels', () => {
    it('should stringify and save the array to localStorage', () => {
      const favorites = ['id1', 'id3'];
      saveFavoriteChannels(favorites);
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        FAVORITES_STORAGE_KEY,
        JSON.stringify(favorites)
      );
      expect(mockLocalStorageStore[FAVORITES_STORAGE_KEY]).toBe(JSON.stringify(favorites));
    });

    it('should overwrite existing values in localStorage', () => {
      saveFavoriteChannels(['id_old']);
      const newFavorites = ['id_new1', 'id_new2'];
      saveFavoriteChannels(newFavorites);
      expect(mockLocalStorageStore[FAVORITES_STORAGE_KEY]).toBe(JSON.stringify(newFavorites));
    });
  });

  describe('toggleFavorite', () => {
    const channelId = 'channel1';
    let favButton;

    beforeEach(() => {
      // Create a mock button for each toggleFavorite test
      favButton = document.createElement('button');
      favButton.id = `favorite-btn-${channelId}`;
      favButton.className = 'favorite-btn';
      document.body.appendChild(favButton);
      
      // Create a mock channel card
      const channelCard = document.createElement('a');
      channelCard.className = 'card';
      channelCard.dataset.channelId = channelId;
      document.getElementById('original-channels-grid').appendChild(channelCard);

      // Reset spy for displayFavoriteChannels
      displayFavoriteChannelsCalled = false; 
    });

    afterEach(() => {
      // Clean up the button and card
      if (favButton && favButton.parentNode) {
        favButton.parentNode.removeChild(favButton);
      }
      const card = document.querySelector(`a.card[data-channel-id="${channelId}"]`);
      if (card && card.parentNode) {
        card.parentNode.removeChild(card);
      }
    });

    it('should add a new favorite, update localStorage, class, and call displayFavoriteChannels', () => {
      toggleFavorite(channelId);
      expect(getFavoriteChannels()).toContain(channelId);
      expect(favButton.classList.contains('favorited')).toBe(true);
      expect(displayFavoriteChannelsCalled).toBe(true);
    });

    it('should remove an existing favorite, update localStorage, class, and call displayFavoriteChannels', () => {
      // Add first
      toggleFavorite(channelId); 
      expect(getFavoriteChannels()).toContain(channelId);
      expect(favButton.classList.contains('favorited')).toBe(true);
      
      displayFavoriteChannelsCalled = false; // Reset spy

      // Then remove
      toggleFavorite(channelId);
      expect(getFavoriteChannels()).not.toContain(channelId);
      expect(favButton.classList.contains('favorited')).toBe(false);
      expect(displayFavoriteChannelsCalled).toBe(true);
    });
  });

  describe('updateFavoriteButtonStates', () => {
    beforeEach(() => {
      // Create some mock buttons
      document.body.innerHTML += `
        <button id="favorite-btn-ch1" class="favorite-btn"></button>
        <button id="favorite-btn-ch2" class="favorite-btn"></button>
        <button id="favorite-btn-ch3" class="favorite-btn"></button>
      `;
    });

    it('should not add "favorited" class if no favorites in localStorage', () => {
      saveFavoriteChannels([]); // Ensure it's empty
      updateFavoriteButtonStates();
      expect(document.getElementById('favorite-btn-ch1').classList.contains('favorited')).toBe(false);
      expect(document.getElementById('favorite-btn-ch2').classList.contains('favorited')).toBe(false);
      expect(document.getElementById('favorite-btn-ch3').classList.contains('favorited')).toBe(false);
    });

    it('should add "favorited" class to correct buttons based on localStorage', () => {
      saveFavoriteChannels(['ch1', 'ch3']);
      updateFavoriteButtonStates();
      expect(document.getElementById('favorite-btn-ch1').classList.contains('favorited')).toBe(true);
      expect(document.getElementById('favorite-btn-ch2').classList.contains('favorited')).toBe(false);
      expect(document.getElementById('favorite-btn-ch3').classList.contains('favorited')).toBe(true);
    });
  });

  describe('displayFavoriteChannels', () => {
    let favSection, favContainer, originalGrid;
    let card1, card2, card3;

    beforeEach(() => {
      // DOM is reset by global beforeEach, grab references
      favSection = document.getElementById('favorite-channels-section');
      favContainer = document.getElementById('favorite-channels-container');
      originalGrid = document.getElementById('original-channels-grid');

      // Create mock channel cards
      card1 = document.createElement('a');
      card1.className = 'card';
      card1.dataset.channelId = 'c1';
      card1.textContent = 'Channel 1';
      originalGrid.appendChild(card1);

      card2 = document.createElement('a');
      card2.className = 'card';
      card2.dataset.channelId = 'c2';
      card2.textContent = 'Channel 2';
      originalGrid.appendChild(card2);
      
      card3 = document.createElement('a');
      card3.className = 'card';
      card3.dataset.channelId = 'c3';
      card3.textContent = 'Channel 3';
      originalGrid.appendChild(card3);
    });

    it('should hide favorites section and keep all cards in original grid if no favorites', () => {
      saveFavoriteChannels([]);
      displayFavoriteChannels();

      expect(favSection.style.display).toBe('none');
      expect(originalGrid.contains(card1)).toBe(true);
      expect(originalGrid.contains(card2)).toBe(true);
      expect(originalGrid.contains(card3)).toBe(true);
      expect(favContainer.children.length).toBe(0);
    });

    it('should show favorites section and move favorite cards to it', () => {
      saveFavoriteChannels(['c1', 'c3']);
      displayFavoriteChannels();

      expect(favSection.style.display).toBe('block');
      expect(favContainer.contains(card1)).toBe(true);
      expect(originalGrid.contains(card2)).toBe(true);
      expect(favContainer.contains(card3)).toBe(true);
      expect(originalGrid.children.length).toBe(1); // card2
      expect(favContainer.children.length).toBe(2); // card1, card3
    });

    it('should move a card from original to favorites when it becomes a favorite', () => {
      saveFavoriteChannels([]); // Start with no favorites
      displayFavoriteChannels();
      expect(originalGrid.contains(card1)).toBe(true);

      saveFavoriteChannels(['c1']); // Mark card1 as favorite
      displayFavoriteChannels(); // Re-render
      
      expect(favSection.style.display).toBe('block');
      expect(favContainer.contains(card1)).toBe(true);
      expect(originalGrid.contains(card1)).toBe(false);
    });

    it('should move a card from favorites to original when it is unfavorited', () => {
      saveFavoriteChannels(['c2']); // Start with card2 as favorite
      displayFavoriteChannels();
      expect(favContainer.contains(card2)).toBe(true);
      expect(favSection.style.display).toBe('block');

      saveFavoriteChannels([]); // Unfavorite card2
      displayFavoriteChannels(); // Re-render

      expect(originalGrid.contains(card2)).toBe(true);
      expect(favContainer.contains(card2)).toBe(false);
      expect(favSection.style.display).toBe('none'); // Assuming it's the only favorite
    });
  });
});
