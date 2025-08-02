/**
 * Tests for Keyboard Navigation functionality
 */

// Mock localStorage
Object.defineProperty(window, 'localStorage', {
  value: {
    getItem: jest.fn(() => null),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn()
  },
  writable: true
});

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock screen
Object.defineProperty(window, 'screen', {
  writable: true,
  value: {
    width: 1920,
    height: 1080
  }
});

// Mock getBoundingClientRect for all elements
Element.prototype.getBoundingClientRect = jest.fn(() => ({
  top: 0,
  left: 0,
  bottom: 100,
  right: 100,
  width: 100,
  height: 100
}));

// Mock scrollIntoView
Element.prototype.scrollIntoView = jest.fn();

const KeyboardNavigation = require('./keyboard-navigation.js');

describe('KeyboardNavigation', () => {
  let keyboardNav;

  beforeEach(() => {
    jest.clearAllMocks();
    // Reset DOM
    document.body.innerHTML = `
      <div class="navbar">
        <div class="navbar-end">
          <button class="btn">Login</button>
        </div>
      </div>
      <div class="container">
        <input id="portexe-search-input" type="text" />
        <select id="portexe-quality-select">
          <option value="auto">Auto</option>
        </select>
        <div class="grid">
          <a href="/play/1" class="card" data-channel-id="1">Channel 1</a>
          <a href="/play/2" class="card" data-channel-id="2">Channel 2</a>
          <a href="/play/3" class="card" data-channel-id="3">Channel 3</a>
        </div>
      </div>
    `;
    keyboardNav = new KeyboardNavigation();
  });

  describe('initialization', () => {
    it('should create TV mode toggle button', () => {
      const toggleButton = document.querySelector('.navbar-end .btn:not(:last-child)');
      expect(toggleButton).toBeTruthy();
      expect(toggleButton.innerHTML).toContain('TV Mode');
    });

    it('should detect TV mode preference', () => {
      localStorage.getItem.mockReturnValue('true');
      const nav = new KeyboardNavigation();
      expect(nav.isTVMode).toBe(true);
    });
  });

  describe('TV mode toggle', () => {
    it('should enable TV mode when toggled', () => {
      keyboardNav.toggleTVMode();
      expect(document.body.classList.contains('tv-mode')).toBe(true);
      expect(keyboardNav.isTVMode).toBe(true);
      expect(keyboardNav.isEnabled).toBe(true);
    });

    it('should disable TV mode when toggled off', () => {
      keyboardNav.enableTVMode();
      keyboardNav.toggleTVMode();
      expect(document.body.classList.contains('tv-mode')).toBe(false);
      expect(keyboardNav.isTVMode).toBe(false);
      expect(keyboardNav.isEnabled).toBe(false);
    });

    it('should save TV mode preference', () => {
      keyboardNav.toggleTVMode();
      expect(localStorage.setItem).toHaveBeenCalledWith('jiotv-tv-mode', 'true');
    });
  });

  describe('focusable elements', () => {
    beforeEach(() => {
      keyboardNav.enableTVMode();
    });

    it('should identify focusable elements', () => {
      expect(keyboardNav.focusableElements.length).toBeGreaterThan(0);
      
      const hasSearchInput = keyboardNav.focusableElements.some(el => 
        el.id === 'portexe-search-input'
      );
      const hasChannelCards = keyboardNav.focusableElements.some(el => 
        el.classList.contains('card')
      );
      
      expect(hasSearchInput).toBe(true);
      expect(hasChannelCards).toBe(true);
    });

    it('should set initial focus', () => {
      expect(keyboardNav.currentFocusIndex).toBe(0);
      const firstElement = keyboardNav.focusableElements[0];
      expect(firstElement.classList.contains('tv-focused')).toBe(true);
    });
  });

  describe('keyboard navigation', () => {
    beforeEach(() => {
      keyboardNav.enableTVMode();
    });

    it('should handle arrow down navigation', () => {
      const initialIndex = keyboardNav.currentFocusIndex;
      
      const event = new KeyboardEvent('keydown', { key: 'ArrowDown' });
      document.dispatchEvent(event);
      
      expect(keyboardNav.currentFocusIndex).toBe(initialIndex + 1);
    });

    it('should handle arrow up navigation', () => {
      keyboardNav.currentFocusIndex = 1;
      keyboardNav.setFocus(1);
      
      const event = new KeyboardEvent('keydown', { key: 'ArrowUp' });
      document.dispatchEvent(event);
      
      expect(keyboardNav.currentFocusIndex).toBe(0);
    });

    it('should handle Enter key selection', () => {
      const mockClick = jest.fn();
      const currentElement = keyboardNav.focusableElements[keyboardNav.currentFocusIndex];
      currentElement.click = mockClick;
      
      const event = new KeyboardEvent('keydown', { key: 'Enter' });
      document.dispatchEvent(event);
      
      expect(mockClick).toHaveBeenCalled();
    });

    it('should ignore keyboard events when TV mode is disabled', () => {
      keyboardNav.disableTVMode();
      
      const initialIndex = keyboardNav.currentFocusIndex;
      const event = new KeyboardEvent('keydown', { key: 'ArrowDown' });
      document.dispatchEvent(event);
      
      expect(keyboardNav.currentFocusIndex).toBe(initialIndex);
    });
  });

  describe('focus management', () => {
    beforeEach(() => {
      keyboardNav.enableTVMode();
    });

    it('should add tv-focused class to current element', () => {
      keyboardNav.setFocus(0);
      const currentElement = keyboardNav.focusableElements[0];
      expect(currentElement.classList.contains('tv-focused')).toBe(true);
    });

    it('should remove tv-focused class from all elements when clearing focus', () => {
      keyboardNav.setFocus(0);
      keyboardNav.clearFocus();
      
      const focusedElements = document.querySelectorAll('.tv-focused');
      expect(focusedElements.length).toBe(0);
    });

    it('should scroll element into view when focused', () => {
      const element = keyboardNav.focusableElements[0];
      keyboardNav.setFocus(0);
      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: 'smooth',
        block: 'center', 
        inline: 'nearest'
      });
    });
  });

  describe('platform detection', () => {
    it('should detect TV user agents', () => {
      Object.defineProperty(window.navigator, 'userAgent', {
        writable: true,
        value: 'Mozilla/5.0 (SMART-TV; LINUX; Tizen 2.4.0) AppleWebKit/538.1'
      });
      
      const nav = new KeyboardNavigation();
      expect(nav.isTVMode).toBe(true);
    });
  });
});