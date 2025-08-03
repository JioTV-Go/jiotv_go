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

// Assign the mock to the global window object
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true
});

// Mock window.matchMedia for system theme detection
const mockMatchMedia = jest.fn();
Object.defineProperty(window, 'matchMedia', {
  value: mockMatchMedia,
  writable: true
});

// Define the functions from common.js for testing
const getCurrentTheme = () => {
  if (localStorage.getItem("theme")) {
    // return local storage theme
    return localStorage.getItem("theme");
  } else if (document.getElementsByTagName("html")[0].hasAttribute("data-theme")) {
    // return data-theme attribute
    const theme = document.getElementsByTagName("html")[0].getAttribute("data-theme");
    localStorage.setItem("theme", theme);
    return theme;
  } else {
    // return system theme
    if (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    ) {
      localStorage.setItem("theme", "dark");
      return "dark";
    }
    localStorage.setItem("theme", "light");
    return "light";
  }
};

const toggleTheme = () => {
  // toggle or add attribute "data-theme" to html tag
  const htmlTag = document.getElementsByTagName("html")[0];
  if (getCurrentTheme() == "dark") {
    localStorage.setItem("theme", "light");
    htmlTag.setAttribute("data-theme", "light");
  } else {
    localStorage.setItem("theme", "dark");
    htmlTag.setAttribute("data-theme", "dark");
  }
};

const initializeTheme = () => {
  const sunIcon = document.getElementById("sunIcon");
  const moonIcon = document.getElementById("moonIcon");
  const htmlTag = document.getElementsByTagName("html")[0];

  if (getCurrentTheme() == "light") {
    if (sunIcon && moonIcon) {
      sunIcon.classList.replace("swap-on", "swap-off");
      moonIcon.classList.replace("swap-off", "swap-on");
    }
    htmlTag.setAttribute("data-theme", "light");
  }
};

describe('Theme Management Functionality', () => {
  beforeEach(() => {
    localStorageMock.clear();
    jest.clearAllMocks();
    
    // Reset HTML structure
    document.body.innerHTML = '';
    
    // Reset HTML tag attributes
    const htmlTag = document.getElementsByTagName("html")[0];
    if (htmlTag) {
      htmlTag.removeAttribute('data-theme');
    }
  });

  describe('getCurrentTheme', () => {
    it('should return theme from localStorage if available', () => {
      localStorageMock.setItem('theme', 'dark');
      expect(getCurrentTheme()).toBe('dark');
    });

    it('should return theme from data-theme attribute if localStorage is empty', () => {
      const htmlTag = document.getElementsByTagName("html")[0];
      htmlTag.setAttribute('data-theme', 'light');
      expect(getCurrentTheme()).toBe('light');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light');
    });

    it('should return system theme preference when no stored theme', () => {
      mockMatchMedia.mockReturnValue({ matches: true });
      expect(getCurrentTheme()).toBe('dark');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark');
    });

    it('should default to light theme when system preference is not dark', () => {
      mockMatchMedia.mockReturnValue({ matches: false });
      expect(getCurrentTheme()).toBe('light');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light');
    });

    it('should default to light theme when matchMedia is not available', () => {
      Object.defineProperty(window, 'matchMedia', { value: null });
      expect(getCurrentTheme()).toBe('light');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light');
    });
  });

  describe('toggleTheme', () => {
    it('should toggle from dark to light theme', () => {
      localStorageMock.setItem('theme', 'dark');
      const htmlTag = document.getElementsByTagName("html")[0];
      
      toggleTheme();
      
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light');
      expect(htmlTag.getAttribute('data-theme')).toBe('light');
    });

    it('should toggle from light to dark theme', () => {
      localStorageMock.setItem('theme', 'light');
      const htmlTag = document.getElementsByTagName("html")[0];
      
      toggleTheme();
      
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark');
      expect(htmlTag.getAttribute('data-theme')).toBe('dark');
    });

    it('should toggle to dark when no theme is set initially', () => {
      mockMatchMedia.mockReturnValue({ matches: false }); // System prefers light
      const htmlTag = document.getElementsByTagName("html")[0];
      
      toggleTheme();
      
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark');
      expect(htmlTag.getAttribute('data-theme')).toBe('dark');
    });
  });

  describe('initializeTheme', () => {
    beforeEach(() => {
      // Create mock theme toggle icons
      const sunIcon = document.createElement('div');
      sunIcon.id = 'sunIcon';
      sunIcon.classList.add('swap-on');
      
      const moonIcon = document.createElement('div');
      moonIcon.id = 'moonIcon';
      moonIcon.classList.add('swap-off');
      
      document.body.appendChild(sunIcon);
      document.body.appendChild(moonIcon);
    });

    it('should initialize light theme UI when current theme is light', () => {
      localStorageMock.setItem('theme', 'light');
      const htmlTag = document.getElementsByTagName("html")[0];
      const sunIcon = document.getElementById("sunIcon");
      const moonIcon = document.getElementById("moonIcon");
      
      initializeTheme();
      
      expect(sunIcon.classList.contains('swap-off')).toBe(true);
      expect(sunIcon.classList.contains('swap-on')).toBe(false);
      expect(moonIcon.classList.contains('swap-on')).toBe(true);
      expect(moonIcon.classList.contains('swap-off')).toBe(false);
      expect(htmlTag.getAttribute('data-theme')).toBe('light');
    });

    it('should not modify UI when current theme is dark', () => {
      localStorageMock.setItem('theme', 'dark');
      const htmlTag = document.getElementsByTagName("html")[0];
      const sunIcon = document.getElementById("sunIcon");
      const moonIcon = document.getElementById("moonIcon");
      
      // Set initial classes
      sunIcon.classList.add('swap-on');
      moonIcon.classList.add('swap-off');
      
      initializeTheme();
      
      // Should remain unchanged for dark theme
      expect(sunIcon.classList.contains('swap-on')).toBe(true);
      expect(moonIcon.classList.contains('swap-off')).toBe(true);
      expect(htmlTag.getAttribute('data-theme')).toBeNull();
    });

    it('should handle missing theme icons gracefully', () => {
      localStorageMock.setItem('theme', 'light');
      
      // Remove icons from DOM
      document.getElementById("sunIcon").remove();
      document.getElementById("moonIcon").remove();
      
      expect(() => initializeTheme()).not.toThrow();
    });
  });
});