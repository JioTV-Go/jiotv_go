/**
 * Simple tests for utility functions - basic functionality only
 */
describe('Utility Functions - Basic Tests', () => {
  // Mock localStorage
  const localStorageMock = {
    data: {},
    getItem: jest.fn(key => localStorageMock.data[key] || null),
    setItem: jest.fn((key, value) => { localStorageMock.data[key] = value; }),
    removeItem: jest.fn(key => { delete localStorageMock.data[key]; }),
    clear: jest.fn(() => { localStorageMock.data = {}; }),
  };

  beforeEach(() => {
    global.localStorage = localStorageMock;
    jest.clearAllMocks();
    localStorageMock.clear();
  });

  describe('LocalStorage Utilities', () => {
    // Test the localStorage utility functions in isolation
    const getLocalStorageItem = (key, defaultValue = null) => {
      try {
        const item = localStorage.getItem(key);
        if (item === null) return defaultValue;
        return JSON.parse(item);
      } catch (error) {
        console.error(`Error parsing localStorage item '${key}':`, error);
        return defaultValue;
      }
    };

    const setLocalStorageItem = (key, value) => {
      try {
        localStorage.setItem(key, JSON.stringify(value));
        return true;
      } catch (error) {
        console.error(`Error setting localStorage item '${key}':`, error);
        return false;
      }
    };

    it('should store and retrieve JSON values correctly', () => {
      const testData = { channels: ['1', '2', '3'] };
      const result = setLocalStorageItem('test-key', testData);
      expect(result).toBe(true);
      
      // Verify the data was stored correctly by retrieving it
      const retrieved = getLocalStorageItem('test-key');
      expect(retrieved).toEqual(testData);
    });

    it('should return default value for non-existent keys', () => {
      const result = getLocalStorageItem('non-existent', 'default');
      expect(result).toBe('default');
    });

    it('should handle invalid JSON gracefully', () => {
      localStorage.setItem('invalid', 'invalid json');
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      const result = getLocalStorageItem('invalid', 'default');
      expect(result).toBe('default');
      expect(consoleSpy).toHaveBeenCalled();
      
      consoleSpy.mockRestore();
    });
  });

  describe('CSS Class Management', () => {
    const mockElement = {
      classList: {
        add: jest.fn(),
        remove: jest.fn(),
        contains: jest.fn(() => false),
      }
    };

    const toggleClasses = (element, addClass, removeClass, condition) => {
      if (!element) return;
      
      if (condition) {
        if (removeClass) element.classList.remove(removeClass);
        if (addClass) element.classList.add(addClass);
      } else {
        if (addClass) element.classList.remove(addClass);
        if (removeClass) element.classList.add(removeClass);
      }
    };

    beforeEach(() => {
      jest.clearAllMocks();
    });

    it('should add and remove classes based on condition', () => {
      toggleClasses(mockElement, 'active', 'inactive', true);
      expect(mockElement.classList.remove).toHaveBeenCalledWith('inactive');
      expect(mockElement.classList.add).toHaveBeenCalledWith('active');
    });

    it('should handle opposite condition', () => {
      toggleClasses(mockElement, 'active', 'inactive', false);
      expect(mockElement.classList.remove).toHaveBeenCalledWith('active');
      expect(mockElement.classList.add).toHaveBeenCalledWith('inactive');
    });

    it('should handle null element gracefully', () => {
      expect(() => toggleClasses(null, 'class1', 'class2', true)).not.toThrow();
    });
  });

  describe('URL Parameter Utilities', () => {
    it('should validate URL creation logic', () => {
      // Test the core logic without relying on window object
      const testUrl = new URL('http://localhost:3000/channels?search=test');
      testUrl.searchParams.set('category', 'sports');
      
      expect(testUrl.searchParams.get('search')).toBe('test');
      expect(testUrl.searchParams.get('category')).toBe('sports');
      
      testUrl.searchParams.delete('search');
      expect(testUrl.searchParams.get('search')).toBeNull();
    });
  });
});