/**
 * Keyboard/D-pad Navigation for JioTV Go
 * Enables TV-friendly navigation for Samsung Tizen, Android TV, etc.
 */

class KeyboardNavigation {
  constructor() {
    this.isEnabled = false;
    this.isTVMode = false;
    this.currentFocusIndex = 0;
    this.focusableElements = [];
    this.keyMappings = {
      'ArrowUp': 'up',
      'ArrowDown': 'down',
      'ArrowLeft': 'left',
      'ArrowRight': 'right',
      'Enter': 'select',
      'Escape': 'back',
      ' ': 'select', // Space bar
      'Backspace': 'back'
    };
    
    this.init();
  }

  init() {
    this.detectPlatform();
    this.setupEventListeners();
    this.createTVModeToggle();
    
    // Initialize on page load
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', () => this.initializeNavigation());
    } else {
      this.initializeNavigation();
    }
  }

  detectPlatform() {
    const userAgent = navigator.userAgent.toLowerCase();
    const isTV = /tizen|webos|roku|androidtv|googletv|smarttv|tv/i.test(userAgent) ||
                 /smart-tv|smarttv/i.test(userAgent) ||
                 window.matchMedia('(pointer: coarse)').matches && 
                 window.screen.width >= 1920;
    
    // Check for TV-like characteristics
    const isTVLike = window.screen.width >= 1920 && 
                     window.screen.height >= 1080 &&
                     !window.ontouchstart;
    
    this.isTVMode = isTV || isTVLike || this.getTVModePreference();
    
    if (this.isTVMode) {
      this.enableTVMode();
    }
  }

  getTVModePreference() {
    return localStorage.getItem('jiotv-tv-mode') === 'true';
  }

  setTVModePreference(enabled) {
    localStorage.setItem('jiotv-tv-mode', enabled.toString());
  }

  createTVModeToggle() {
    const navbar = document.querySelector('.navbar-end');
    if (!navbar) return;

    const toggleButton = document.createElement('button');
    toggleButton.className = 'btn btn-outline btn-sm';
    toggleButton.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 20.25h12m-7.5-3v3m3-3v3m-10.125-3h17.25c.621 0 1.125-.504 1.125-1.125V4.875c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125z" />
      </svg>
      <span class="hidden sm:inline ml-1">TV Mode</span>
    `;
    toggleButton.title = 'Toggle TV Mode';
    toggleButton.onclick = () => this.toggleTVMode();
    
    // Insert before the last child (login/logout button)
    navbar.insertBefore(toggleButton, navbar.lastElementChild);
  }

  toggleTVMode() {
    this.isTVMode = !this.isTVMode;
    this.setTVModePreference(this.isTVMode);
    
    if (this.isTVMode) {
      this.enableTVMode();
    } else {
      this.disableTVMode();
    }
  }

  enableTVMode() {
    this.isEnabled = true;
    this.isTVMode = true;
    document.body.classList.add('tv-mode');
    this.updateFocusableElements();
    this.setInitialFocus();
  }

  disableTVMode() {
    this.isEnabled = false;
    this.isTVMode = false;
    document.body.classList.remove('tv-mode');
    this.clearFocus();
  }

  setupEventListeners() {
    document.addEventListener('keydown', (e) => this.handleKeyDown(e));
    
    // Handle mouse usage to temporarily disable keyboard navigation
    document.addEventListener('mousemove', (e) => {
      if (this.isTVMode && (Math.abs(e.movementX) > 0 || Math.abs(e.movementY) > 0)) {
        this.temporarilyDisable();
      }
    });

    // Re-enable on subsequent keyboard usage
    document.addEventListener('keydown', () => {
      if (this.isTVMode && !this.isEnabled) {
        this.isEnabled = true;
        this.updateFocusableElements();
        this.setInitialFocus();
      }
    });
  }

  temporarilyDisable() {
    this.isEnabled = false;
    this.clearFocus();
  }

  handleKeyDown(e) {
    if (!this.isEnabled || !this.isTVMode) return;

    const action = this.keyMappings[e.key];
    if (!action) return;

    e.preventDefault();
    
    switch (action) {
      case 'up':
        this.navigateUp();
        break;
      case 'down':
        this.navigateDown();
        break;
      case 'left':
        this.navigateLeft();
        break;
      case 'right':
        this.navigateRight();
        break;
      case 'select':
        this.selectCurrent();
        break;
      case 'back':
        this.goBack();
        break;
    }
  }

  updateFocusableElements() {
    // Get all focusable elements in order
    const selectors = [
      '#portexe-search-input',
      '#portexe-quality-select',
      '#portexe-category-select', 
      '#portexe-language-select',
      '#portexe-search-button',
      '.card[href]', // Channel cards
      '.btn', // Other buttons
      'button:not(.favorite-btn)', // Regular buttons but not favorite buttons
      'a[href]', // Links
      'input',
      'select'
    ];

    this.focusableElements = [];
    
    selectors.forEach(selector => {
      const elements = document.querySelectorAll(selector);
      elements.forEach(el => {
        if (this.isElementVisible(el) && !el.hasAttribute('disabled')) {
          this.focusableElements.push(el);
        }
      });
    });

    // Remove duplicates while preserving order (O(n) using Set)
    this.focusableElements = [...new Set(this.focusableElements)];
  }

  isElementVisible(el) {
    const style = window.getComputedStyle(el);
    return style.display !== 'none' && 
           style.visibility !== 'hidden' && 
           el.offsetParent !== null;
  }

  setInitialFocus() {
    if (this.focusableElements.length > 0) {
      this.currentFocusIndex = 0;
      this.setFocus(this.currentFocusIndex);
    }
  }

  setFocus(index) {
    if (index < 0 || index >= this.focusableElements.length) return;
    
    // Remove focus from all elements
    this.clearFocus();
    
    // Set focus on current element
    const element = this.focusableElements[index];
    if (element) {
      element.classList.add('tv-focused');
      element.focus();
      
      // Scroll into view
      element.scrollIntoView({ 
        behavior: 'smooth', 
        block: 'center',
        inline: 'nearest'
      });
    }
  }

  clearFocus() {
    document.querySelectorAll('.tv-focused').forEach(el => {
      el.classList.remove('tv-focused');
    });
  }

  navigateUp() {
    const currentElement = this.focusableElements[this.currentFocusIndex];
    if (!currentElement) return;

    // Special handling for channel grid
    if (currentElement.classList.contains('card')) {
      const newIndex = this.findElementInDirection('up');
      if (newIndex !== -1) {
        this.currentFocusIndex = newIndex;
        this.setFocus(this.currentFocusIndex);
        return;
      }
    }

    // Default navigation - previous element
    if (this.currentFocusIndex > 0) {
      this.currentFocusIndex--;
      this.setFocus(this.currentFocusIndex);
    }
  }

  navigateDown() {
    const currentElement = this.focusableElements[this.currentFocusIndex];
    if (!currentElement) return;

    // Special handling for channel grid
    if (currentElement.classList.contains('card')) {
      const newIndex = this.findElementInDirection('down');
      if (newIndex !== -1) {
        this.currentFocusIndex = newIndex;
        this.setFocus(this.currentFocusIndex);
        return;
      }
    }

    // Default navigation - next element
    if (this.currentFocusIndex < this.focusableElements.length - 1) {
      this.currentFocusIndex++;
      this.setFocus(this.currentFocusIndex);
    }
  }

  navigateLeft() {
    const currentElement = this.focusableElements[this.currentFocusIndex];
    if (!currentElement) return;

    // Special handling for channel grid
    if (currentElement.classList.contains('card')) {
      const newIndex = this.findElementInDirection('left');
      if (newIndex !== -1) {
        this.currentFocusIndex = newIndex;
        this.setFocus(this.currentFocusIndex);
        return;
      }
    }

    // Default navigation - previous element
    if (this.currentFocusIndex > 0) {
      this.currentFocusIndex--;
      this.setFocus(this.currentFocusIndex);
    }
  }

  navigateRight() {
    const currentElement = this.focusableElements[this.currentFocusIndex];
    if (!currentElement) return;

    // Special handling for channel grid
    if (currentElement.classList.contains('card')) {
      const newIndex = this.findElementInDirection('right');
      if (newIndex !== -1) {
        this.currentFocusIndex = newIndex;
        this.setFocus(this.currentFocusIndex);
        return;
      }
    }

    // Default navigation - next element
    if (this.currentFocusIndex < this.focusableElements.length - 1) {
      this.currentFocusIndex++;
      this.setFocus(this.currentFocusIndex);
    }
  }

  findElementInDirection(direction) {
    const currentElement = this.focusableElements[this.currentFocusIndex];
    if (!currentElement) return -1;

    const currentRect = currentElement.getBoundingClientRect();
    const channelCards = this.focusableElements.filter(el => el.classList.contains('card'));
    
    let bestElement = null;
    let bestDistance = Infinity;

    for (let i = 0; i < channelCards.length; i++) {
      const card = channelCards[i];
      if (card === currentElement) continue;
      
      const rect = card.getBoundingClientRect();
      const actualIndex = this.focusableElements.findIndex(el => el === card);
      
      if (actualIndex === -1) continue; // Element not found in focusable elements
      
      let isValidDirection = false;
      let distance = 0;

      switch (direction) {
        case 'up':
          isValidDirection = rect.bottom <= currentRect.top;
          distance = currentRect.top - rect.bottom + Math.abs(rect.left - currentRect.left);
          break;
        case 'down':
          isValidDirection = rect.top >= currentRect.bottom;
          distance = rect.top - currentRect.bottom + Math.abs(rect.left - currentRect.left);
          break;
        case 'left':
          isValidDirection = rect.right <= currentRect.left;
          distance = currentRect.left - rect.right + Math.abs(rect.top - currentRect.top);
          break;
        case 'right':
          isValidDirection = rect.left >= currentRect.right;
          distance = rect.left - currentRect.right + Math.abs(rect.top - currentRect.top);
          break;
      }

      if (isValidDirection && distance < bestDistance) {
        bestDistance = distance;
        bestElement = actualIndex;
      }
    }

    return bestElement;
  }

  selectCurrent() {
    const currentElement = this.focusableElements[this.currentFocusIndex];
    if (!currentElement) return;

    // Handle different element types
    if (currentElement.tagName === 'A') {
      currentElement.click();
    } else if (currentElement.tagName === 'BUTTON') {
      currentElement.click();
    } else if (currentElement.tagName === 'INPUT') {
      currentElement.focus();
    } else if (currentElement.tagName === 'SELECT') {
      currentElement.focus();
    }
  }

  goBack() {
    // Try browser back first
    if (window.history.length > 1) {
      window.history.back();
    } else if (window.location.pathname !== '/') {
      // Go to home page
      window.location.href = '/';
    }
  }

  initializeNavigation() {
    if (this.isTVMode) {
      this.enableTVMode();
    }
  }
}

// Initialize keyboard navigation when script loads
if (typeof module === 'undefined' || !module.exports) {
  const keyboardNavigation = new KeyboardNavigation();
}

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = KeyboardNavigation;
}