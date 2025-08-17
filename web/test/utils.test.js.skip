// Set up TextEncoder/TextDecoder for JSDOM compatibility
const { TextEncoder, TextDecoder } = require('util');
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Mock localStorage for testing
const localStorageMock = {
  data: {},
  getItem: jest.fn(key => localStorageMock.data[key] || null),
  setItem: jest.fn((key, value) => {
    localStorageMock.data[key] = value;
  }),
  removeItem: jest.fn(key => {
    delete localStorageMock.data[key];
  }),
  clear: jest.fn(() => {
    localStorageMock.data = {};
  }),
};

// Mock fetch for testing
global.fetch = jest.fn();

// Mock window and document
const mockElement = {
  textContent: 'Test Content',
  innerHTML: '',
  classList: {
    add: jest.fn(),
    remove: jest.fn(),
    contains: jest.fn(() => false),
    replace: jest.fn(),
  },
  setAttribute: jest.fn(),
  getAttribute: jest.fn(),
  appendChild: jest.fn(),
  id: '',
  className: '',
  tagName: 'DIV',
};

global.document = {
  getElementById: jest.fn((id) => {
    if (id === 'test-element' || id === 'element-1' || id === 'element-2' || 
        id.startsWith('favorite-btn-') || id.startsWith('star-icon-') || id.startsWith('x-icon-')) {
      const element = { ...mockElement };
      element.id = id;
      if (id === 'test-element') {
        element.textContent = 'Test Content';
      }
      if (id.includes('hidden')) {
        element.classList.contains = jest.fn(() => true);
      }
      return element;
    }
    return null;
  }),
  createElement: jest.fn((tagName) => ({
    ...mockElement,
    tagName: tagName.toUpperCase(),
  })),
};

global.window = {
  location: {
    href: 'http://localhost:3000/channels?search=test&category=sports',
    pathname: '/channels',
    search: '?search=test&category=sports'
  },
  history: {
    replaceState: jest.fn()
  },
  URLSearchParams: URLSearchParams,
};

global.localStorage = localStorageMock;

// Load the utility functions by evaluating the file content
const fs = require('fs');
const path = require('path');
const utilsScript = fs.readFileSync(path.join(__dirname, '..', 'static', 'internal', 'utils.js'), 'utf8');

// Remove the module.exports part and evaluate the functions
const scriptWithoutExports = utilsScript.replace(/if \(typeof module.*\{[\s\S]*\}/, '');
eval(scriptWithoutExports);

describe('Utility Functions', () => {
  beforeEach(() => {
    // Clear mocks
    jest.clearAllMocks();
    localStorageMock.clear();
  });

  afterEach(() => {
    // Reset mocks
    jest.resetAllMocks();
  });

  describe('DOM Utilities', () => {
    describe('safeGetElementById', () => {
      it('should return element when it exists', () => {
        const element = safeGetElementById('test-element');
        expect(element).toBeTruthy();
        expect(element.textContent).toBe('Test Content');
      });

      it('should return null when element does not exist', () => {
        const element = safeGetElementById('non-existent');
        expect(element).toBeNull();
      });

      it('should log warning when element not found and suppressError is false', () => {
        const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
        safeGetElementById('non-existent');
        expect(consoleSpy).toHaveBeenCalledWith("Element with ID 'non-existent' not found");
        consoleSpy.mockRestore();
      });

      it('should not log warning when suppressError is true', () => {
        const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
        safeGetElementById('non-existent', true);
        expect(consoleSpy).not.toHaveBeenCalled();
        consoleSpy.mockRestore();
      });
    });

    describe('safeGetElementsById', () => {
      it('should return object with all requested elements', () => {
        const elements = safeGetElementsById(['test-element', 'element-1', 'element-2']);
        expect(elements['test-element']).toBeTruthy();
        expect(elements['element-1']).toBeTruthy();
        expect(elements['element-2']).toBeTruthy();
        expect(elements['test-element'].textContent).toBe('Test Content');
      });

      it('should include null values for non-existent elements', () => {
        const elements = safeGetElementsById(['test-element', 'non-existent']);
        expect(elements['test-element']).toBeTruthy();
        expect(elements['non-existent']).toBeNull();
      });
    });

    describe('createElement', () => {
      it('should create element with basic attributes', () => {
        const element = createElement('div', { id: 'new-div', className: 'test-class' }, 'Test content');
        expect(element.tagName).toBe('DIV');
        expect(element.id).toBe('new-div');
        expect(element.className).toBe('test-class');
        expect(element.textContent).toBe('Test content');
      });

      it('should create element with innerHTML when provided', () => {
        const element = createElement('div', {}, 'Text content', '<span>HTML content</span>');
        expect(element.innerHTML).toBe('<span>HTML content</span>');
        expect(element.textContent).toBe('HTML content');
      });

      it('should create element with custom attributes', () => {
        const element = createElement('a', { 
          href: '/test', 
          'data-channel-id': '123',
          tabindex: '0'
        });
        expect(element.setAttribute).toHaveBeenCalledWith('href', '/test');
        expect(element.setAttribute).toHaveBeenCalledWith('data-channel-id', '123');
        expect(element.setAttribute).toHaveBeenCalledWith('tabindex', '0');
      });
    });
  });

  describe('CSS Class Utilities', () => {
    describe('toggleClasses', () => {
      it('should add and remove classes based on condition (true)', () => {
        const element = document.getElementById('test-element');
        toggleClasses(element, 'active', 'inactive', true);
        expect(element.classList.contains('active')).toBe(true);
        expect(element.classList.contains('inactive')).toBe(false);
      });

      it('should add and remove classes based on condition (false)', () => {
        const element = document.getElementById('test-element');
        element.classList.add('active');
        toggleClasses(element, 'active', 'inactive', false);
        expect(element.classList.contains('active')).toBe(false);
        expect(element.classList.contains('inactive')).toBe(true);
      });

      it('should handle null element gracefully', () => {
        expect(() => toggleClasses(null, 'class1', 'class2', true)).not.toThrow();
      });
    });

    describe('setElementVisibility', () => {
      it('should remove hidden class when visible is true', () => {
        const element = document.getElementById('test-element');
        element.classList.add('hidden');
        setElementVisibility(element, true);
        expect(element.classList.contains('hidden')).toBe(false);
      });

      it('should add hidden class when visible is false', () => {
        const element = document.getElementById('test-element');
        setElementVisibility(element, false);
        expect(element.classList.contains('hidden')).toBe(true);
      });
    });
  });

  describe('LocalStorage Utilities', () => {
    describe('getLocalStorageItem', () => {
      it('should return parsed JSON value when item exists', () => {
        localStorageMock.setItem('test-key', JSON.stringify({ name: 'test' }));
        const result = getLocalStorageItem('test-key');
        expect(result).toEqual({ name: 'test' });
      });

      it('should return default value when item does not exist', () => {
        const result = getLocalStorageItem('non-existent', 'default');
        expect(result).toBe('default');
      });

      it('should return default value when JSON parsing fails', () => {
        localStorageMock.setItem('invalid-json', 'invalid json');
        const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
        const result = getLocalStorageItem('invalid-json', 'default');
        expect(result).toBe('default');
        expect(consoleSpy).toHaveBeenCalled();
        consoleSpy.mockRestore();
      });
    });

    describe('setLocalStorageItem', () => {
      it('should store value as JSON string', () => {
        const result = setLocalStorageItem('test-key', { name: 'test' });
        expect(result).toBe(true);
        expect(localStorageMock.setItem).toHaveBeenCalledWith('test-key', '{"name":"test"}');
      });

      it('should handle storage errors gracefully', () => {
        localStorageMock.setItem.mockImplementation(() => {
          throw new Error('Storage quota exceeded');
        });
        const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
        const result = setLocalStorageItem('test-key', 'value');
        expect(result).toBe(false);
        expect(consoleSpy).toHaveBeenCalled();
        consoleSpy.mockRestore();
      });
    });

    describe('removeLocalStorageItem', () => {
      it('should remove item from localStorage', () => {
        const result = removeLocalStorageItem('test-key');
        expect(result).toBe(true);
        expect(localStorageMock.removeItem).toHaveBeenCalledWith('test-key');
      });
    });
  });

  describe('URL Utilities', () => {
    describe('getCurrentUrlParams', () => {
      it('should return URLSearchParams object', () => {
        const params = getCurrentUrlParams();
        expect(params).toBeInstanceOf(window.URLSearchParams);
        expect(params.get('search')).toBe('test');
        expect(params.get('category')).toBe('sports');
      });
    });

    describe('updateUrlParameter', () => {
      it('should update existing parameter', () => {
        updateUrlParameter('search', 'new-search');
        expect(global.history.replaceState).toHaveBeenCalledWith(
          {},
          '',
          '/channels?search=new-search&category=sports'
        );
      });

      it('should add new parameter', () => {
        updateUrlParameter('language', 'english');
        expect(global.history.replaceState).toHaveBeenCalledWith(
          {},
          '',
          '/channels?search=test&category=sports&language=english'
        );
      });

      it('should remove parameter when value is empty', () => {
        updateUrlParameter('search', '');
        expect(global.history.replaceState).toHaveBeenCalledWith(
          {},
          '',
          '/channels?category=sports'
        );
      });

      it('should use custom replaceState function when provided', () => {
        const customReplace = jest.fn();
        updateUrlParameter('search', 'custom', customReplace);
        expect(customReplace).toHaveBeenCalledWith(
          {},
          '',
          '/channels?search=custom&category=sports'
        );
      });
    });

    describe('updateUrlParameters', () => {
      it('should update multiple parameters', () => {
        updateUrlParameters({
          search: 'new-search',
          category: 'news',
          language: 'english'
        });
        expect(global.history.replaceState).toHaveBeenCalledWith(
          {},
          '',
          '/channels?search=new-search&category=news&language=english'
        );
      });
    });
  });

  describe('Fetch Utilities', () => {
    describe('postJSON', () => {
      it('should make POST request with JSON body', async () => {
        const mockResponse = { status: 'success' };
        fetch.mockResolvedValueOnce({
          json: () => Promise.resolve(mockResponse)
        });

        const result = await postJSON('/api/test', { name: 'test' });
        
        expect(fetch).toHaveBeenCalledWith('/api/test', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: '{"name":"test"}'
        });
        expect(result).toEqual(mockResponse);
      });

      it('should handle fetch errors', async () => {
        fetch.mockRejectedValueOnce(new Error('Network error'));
        const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

        await expect(postJSON('/api/test', {})).rejects.toThrow('Network error');
        expect(consoleSpy).toHaveBeenCalled();
        consoleSpy.mockRestore();
      });
    });

    describe('getJSON', () => {
      it('should make GET request and return JSON', async () => {
        const mockResponse = { data: 'test' };
        fetch.mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockResponse)
        });

        const result = await getJSON('/api/test');
        
        expect(fetch).toHaveBeenCalledWith('/api/test', {});
        expect(result).toEqual(mockResponse);
      });

      it('should handle HTTP errors', async () => {
        fetch.mockResolvedValueOnce({
          ok: false,
          status: 404
        });
        const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

        await expect(getJSON('/api/test')).rejects.toThrow('HTTP error! status: 404');
        expect(consoleSpy).toHaveBeenCalled();
        consoleSpy.mockRestore();
      });
    });
  });

  describe('Icon and Button Utilities', () => {
    describe('updateFavoriteButtonState', () => {
      it('should update button state when favorited', () => {
        updateFavoriteButtonState('123', true);
        
        const button = document.getElementById('favorite-btn-123');
        const starIcon = document.getElementById('star-icon-123');
        const xIcon = document.getElementById('x-icon-123');
        
        expect(button.classList.contains('favorited')).toBe(true);
        expect(starIcon.classList.contains('hidden')).toBe(true);
        expect(xIcon.classList.contains('hidden')).toBe(false);
      });

      it('should update button state when not favorited', () => {
        const button = document.getElementById('favorite-btn-123');
        const starIcon = document.getElementById('star-icon-123');
        const xIcon = document.getElementById('x-icon-123');
        
        // Set initial favorited state
        button.classList.add('favorited');
        starIcon.classList.add('hidden');
        xIcon.classList.remove('hidden');
        
        updateFavoriteButtonState('123', false);
        
        expect(button.classList.contains('favorited')).toBe(false);
        expect(starIcon.classList.contains('hidden')).toBe(false);
        expect(xIcon.classList.contains('hidden')).toBe(true);
      });

      it('should handle missing elements gracefully', () => {
        expect(() => updateFavoriteButtonState('non-existent', true)).not.toThrow();
      });
    });
  });
});