/**
 * Utility functions to reduce code repetition across the application
 */

// DOM Utilities
/**
 * Safely get DOM element by ID with optional error handling
 * @param {string} id - Element ID
 * @param {boolean} suppressError - Whether to suppress console errors
 * @returns {Element|null} - DOM element or null if not found
 */
function safeGetElementById(id, suppressError = false) {
  const element = document.getElementById(id);
  if (!element && !suppressError) {
    console.warn(`Element with ID '${id}' not found`);
  }
  return element;
}

/**
 * Get multiple DOM elements by their IDs
 * @param {string[]} ids - Array of element IDs
 * @param {boolean} suppressError - Whether to suppress console errors
 * @returns {Object} - Object with ID as key and element as value
 */
function safeGetElementsById(ids, suppressError = false) {
  const elements = {};
  ids.forEach(id => {
    elements[id] = safeGetElementById(id, suppressError);
  });
  return elements;
}

/**
 * Create DOM element with attributes and content
 * @param {string} tagName - Element tag name
 * @param {Object} attributes - Object with attribute key-value pairs
 * @param {string} content - Inner text content
 * @param {string} innerHTML - Inner HTML content (takes precedence over content)
 * @returns {Element} - Created DOM element
 */
function createElement(tagName, attributes = {}, content = '', innerHTML = '') {
  const element = document.createElement(tagName);
  
  // Set attributes
  Object.entries(attributes).forEach(([key, value]) => {
    if (key === 'className') {
      element.className = value;
    } else {
      element.setAttribute(key, value);
    }
  });
  
  // Set content
  if (innerHTML) {
    element.innerHTML = innerHTML;
  } else if (content) {
    element.innerText = content;
  }
  
  return element;
}

// CSS Class Utilities
/**
 * Toggle CSS classes on element with conditional logic
 * @param {Element} element - DOM element
 * @param {string} addClass - Class to add
 * @param {string} removeClass - Class to remove
 * @param {boolean} condition - Condition to determine action
 */
function toggleClasses(element, addClass, removeClass, condition) {
  if (!element) return;
  
  if (condition) {
    if (removeClass) element.classList.remove(removeClass);
    if (addClass) element.classList.add(addClass);
  } else {
    if (addClass) element.classList.remove(addClass);
    if (removeClass) element.classList.add(removeClass);
  }
}

/**
 * Set element visibility by toggling hidden class
 * @param {Element} element - DOM element
 * @param {boolean} visible - Whether element should be visible
 */
function setElementVisibility(element, visible) {
  if (!element) return;
  toggleClasses(element, null, 'hidden', visible);
}

// LocalStorage Utilities
/**
 * Safely get item from localStorage with JSON parsing
 * @param {string} key - Storage key
 * @param {*} defaultValue - Default value if key doesn't exist or parsing fails
 * @returns {*} - Parsed value or default
 */
function getLocalStorageItem(key, defaultValue = null) {
  try {
    const item = localStorage.getItem(key);
    if (item === null) return defaultValue;
    return JSON.parse(item);
  } catch (error) {
    console.error(`Error parsing localStorage item '${key}':`, error);
    return defaultValue;
  }
}

/**
 * Safely set item in localStorage with JSON stringification
 * @param {string} key - Storage key
 * @param {*} value - Value to store
 * @returns {boolean} - Whether operation succeeded
 */
function setLocalStorageItem(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch (error) {
    console.error(`Error setting localStorage item '${key}':`, error);
    return false;
  }
}

/**
 * Remove item from localStorage
 * @param {string} key - Storage key
 * @returns {boolean} - Whether operation succeeded
 */
function removeLocalStorageItem(key) {
  try {
    localStorage.removeItem(key);
    return true;
  } catch (error) {
    console.error(`Error removing localStorage item '${key}':`, error);
    return false;
  }
}

// URL Utilities
/**
 * Get current URL search parameters
 * @returns {URLSearchParams} - URL search parameters object
 */
function getCurrentUrlParams() {
  return new URLSearchParams(window.location.search);
}

/**
 * Update URL search parameter without page reload
 * @param {string} key - Parameter key
 * @param {string} value - Parameter value (empty string to remove)
 * @param {Function} replaceStateFunc - Optional custom history.replaceState function
 */
function updateUrlParameter(key, value, replaceStateFunc = null) {
  const url = new URL(window.location.href);
  
  if (value && value.trim() !== '') {
    url.searchParams.set(key, value);
  } else {
    url.searchParams.delete(key);
  }
  
  const newUrl = `${url.pathname}?${url.searchParams}`;
  
  if (replaceStateFunc) {
    replaceStateFunc({}, '', newUrl);
  } else {
    window.history.replaceState({}, '', newUrl);
  }
}

/**
 * Update multiple URL search parameters
 * @param {Object} params - Object with key-value pairs to update
 * @param {Function} replaceStateFunc - Optional custom history.replaceState function
 */
function updateUrlParameters(params, replaceStateFunc = null) {
  const url = new URL(window.location.href);
  
  Object.entries(params).forEach(([key, value]) => {
    if (value && value.trim() !== '') {
      url.searchParams.set(key, value);
    } else {
      url.searchParams.delete(key);
    }
  });
  
  const newUrl = `${url.pathname}?${url.searchParams}`;
  
  if (replaceStateFunc) {
    replaceStateFunc({}, '', newUrl);
  } else {
    window.history.replaceState({}, '', newUrl);
  }
}

// Fetch Utilities
/**
 * Make POST request with JSON body and common error handling
 * @param {string} url - Request URL
 * @param {Object} data - Data to send in request body
 * @param {Object} options - Additional fetch options
 * @returns {Promise} - Fetch response promise
 */
async function postJSON(url, data, options = {}) {
  const defaultOptions = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  };
  
  const fetchOptions = { ...defaultOptions, ...options };
  
  try {
    const response = await fetch(url, fetchOptions);
    return await response.json();
  } catch (error) {
    console.error(`Error making POST request to ${url}:`, error);
    throw error;
  }
}

/**
 * Make GET request with common error handling
 * @param {string} url - Request URL
 * @param {Object} options - Additional fetch options
 * @returns {Promise} - Fetch response promise
 */
async function getJSON(url, options = {}) {
  try {
    const response = await fetch(url, options);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error(`Error making GET request to ${url}:`, error);
    throw error;
  }
}

// Icon and Button Utilities
/**
 * Update favorite button state with star/x icons
 * @param {string} channelId - Channel ID
 * @param {boolean} isFavorite - Whether channel is favorited
 */
function updateFavoriteButtonState(channelId, isFavorite) {
  const elements = safeGetElementsById([
    `favorite-btn-${channelId}`,
    `star-icon-${channelId}`,
    `x-icon-${channelId}`
  ], true);
  
  const { [`favorite-btn-${channelId}`]: button, [`star-icon-${channelId}`]: starIcon, [`x-icon-${channelId}`]: xIcon } = elements;
  
  if (button) {
    toggleClasses(button, 'favorited', null, isFavorite);
  }
  
  if (starIcon && xIcon) {
    setElementVisibility(starIcon, !isFavorite);
    setElementVisibility(xIcon, isFavorite);
  }
}

// Export functions for use in other files (if module system is available)
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    safeGetElementById,
    safeGetElementsById,
    createElement,
    toggleClasses,
    setElementVisibility,
    getLocalStorageItem,
    setLocalStorageItem,
    removeLocalStorageItem,
    getCurrentUrlParams,
    updateUrlParameter,
    updateUrlParameters,
    postJSON,
    getJSON,
    updateFavoriteButtonState,
  };
}