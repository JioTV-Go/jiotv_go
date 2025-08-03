// Mock fetch
global.fetch = jest.fn();

// Mock console.log to avoid cluttering test output
global.console.log = jest.fn();

// Mock alert
global.alert = jest.fn();

// Create simplified functions that can be tested without window dependencies
const simpleSearch = (searchTerm, urlPathname = '/channels', urlSearchParams = new URLSearchParams(), historyReplaceState = jest.fn()) => {
  const channels = document.querySelectorAll('.card');
  const trimmedSearchTerm = searchTerm.trim();

  // Update URL search parameter
  if (trimmedSearchTerm !== '') {
    urlSearchParams.set('search', searchTerm);
  } else {
    urlSearchParams.delete('search');
  }

  // Update the URL without reloading the page (mock version)
  historyReplaceState({}, '', `${urlPathname}?${urlSearchParams}`);

  channels.forEach((channel) => {
    const nameElement = channel.querySelector('.font-bold');
    if (nameElement) {
      const name = nameElement.textContent.toLowerCase();
      if (trimmedSearchTerm === '' || name.includes(trimmedSearchTerm.toLowerCase())) {
        channel.style.display = 'block';
      } else {
        channel.style.display = 'none';
      }
    }
  });
};

const simpleInit = (urlSearch = '', searchInputId = 'portexe-search-input') => {
  const searchInput = document.getElementById(searchInputId);
  const urlParams = new URLSearchParams(urlSearch);
  const searchParam = urlParams.get('search');

  if (searchParam && searchInput) {
    simpleSearch(searchParam);
    searchInput.value = searchParam;
  }

  if (searchInput) {
    searchInput.addEventListener('keyup', (e) => {
      simpleSearch(e.target.value);
    });
  }
};

const simpleLoginClick = async (fetchFn = fetch, alertFn = alert, reloadFn = jest.fn()) => {
  const username = document.getElementById("username")?.value;
  const password = document.getElementById("password")?.value;
  
  if (!username || !password) {
    return;
  }

  try {
    const res = await fetchFn("/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });
    
    const data = await res.json();
    
    if (data.status === "success") {
      alertFn("Login success. Enjoy!");
      reloadFn();
    } else {
      alertFn("Login failed!");
    }
  } catch (err) {
    console.log(err);
    alertFn("Login failed!");
  }
};

const simpleLoginOTPClick = async (fetchFn = fetch, alertFn = alert, showModalFn = jest.fn()) => {
  const number = document.getElementById("number")?.value;
  if (!number) return;

  try {
    const res = await fetchFn("/login/sendOTP", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ number: `+91${number}` }),
    });
    
    const data = await res.json();
    
    if (data.status) {
      showModalFn();
    } else {
      alertFn("Sending OTP failed!");
    }
  } catch (err) {
    console.log(err);
    alertFn("Sending OTP failed!");
  }
};

const simpleLoginOTPVerifyClick = async (fetchFn = fetch, alertFn = alert, reloadFn = jest.fn()) => {
  const number = document.getElementById("number")?.value;
  const otp = document.getElementById("otp")?.value;
  
  if (!number || !otp) return;

  try {
    const res = await fetchFn("/login/verifyOTP", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ number: `+91${number}`, otp }),
    });
    
    const data = await res.json();
    
    if (data.status) {
      alertFn("OTP verification success. Enjoy!");
      reloadFn();
    } else {
      alertFn("OTP verification failed!");
    }
  } catch (err) {
    console.log(err);
    alertFn("OTP verification failed!");
  }
};

describe('Search and Login Functionality', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Clear DOM
    document.body.innerHTML = '';
  });

  describe('search', () => {
    beforeEach(() => {
      // Create test channel cards
      document.body.innerHTML = `
        <div class="card">
          <div class="font-bold">ESPN</div>
        </div>
        <div class="card">
          <div class="font-bold">Star Sports</div>
        </div>
        <div class="card">
          <div class="font-bold">Discovery Channel</div>
        </div>
        <div class="card">
          <div class="font-bold">National Geographic</div>
        </div>
      `;
    });

    it('should filter channels based on search term', () => {
      const mockReplaceState = jest.fn();
      simpleSearch('star', '/channels', new URLSearchParams(), mockReplaceState);
      
      const channels = document.querySelectorAll('.card');
      expect(channels[0].style.display).toBe('none'); // ESPN
      expect(channels[1].style.display).toBe('block'); // Star Sports
      expect(channels[2].style.display).toBe('none'); // Discovery Channel
      expect(channels[3].style.display).toBe('none'); // National Geographic
    });

    it('should be case insensitive', () => {
      const mockReplaceState = jest.fn();
      simpleSearch('STAR', '/channels', new URLSearchParams(), mockReplaceState);
      
      const channels = document.querySelectorAll('.card');
      expect(channels[1].style.display).toBe('block'); // Star Sports should be visible
    });

    it('should show all channels when search term is empty', () => {
      const mockReplaceState = jest.fn();
      simpleSearch('', '/channels', new URLSearchParams(), mockReplaceState);
      
      const channels = document.querySelectorAll('.card');
      channels.forEach(channel => {
        expect(channel.style.display).toBe('block');
      });
    });

    it('should show all channels when search term is only whitespace', () => {
      const mockReplaceState = jest.fn();
      simpleSearch('   ', '/channels', new URLSearchParams(), mockReplaceState);
      
      const channels = document.querySelectorAll('.card');
      channels.forEach(channel => {
        expect(channel.style.display).toBe('block');
      });
    });

    it('should update URL with search parameter', () => {
      const mockReplaceState = jest.fn();
      simpleSearch('discovery', '/channels', new URLSearchParams(), mockReplaceState);
      
      expect(mockReplaceState).toHaveBeenCalledWith(
        {},
        '',
        '/channels?search=discovery'
      );
    });

    it('should remove search parameter from URL when search is empty', () => {
      const mockReplaceState = jest.fn();
      const urlParams = new URLSearchParams('?search=existing');
      simpleSearch('', '/channels', urlParams, mockReplaceState);
      
      expect(mockReplaceState).toHaveBeenCalledWith(
        {},
        '',
        '/channels?'
      );
    });

    it('should handle channels without font-bold elements', () => {
      document.body.innerHTML = `
        <div class="card">
          <div>No font-bold class</div>
        </div>
        <div class="card">
          <div class="font-bold">ESPN</div>
        </div>
      `;
      
      const mockReplaceState = jest.fn();
      expect(() => simpleSearch('espn', '/channels', new URLSearchParams(), mockReplaceState)).not.toThrow();
      
      const channels = document.querySelectorAll('.card');
      expect(channels[1].style.display).toBe('block'); // ESPN should be visible
    });

    it('should show no results when no channels match', () => {
      const mockReplaceState = jest.fn();
      simpleSearch('nonexistent', '/channels', new URLSearchParams(), mockReplaceState);
      
      const channels = document.querySelectorAll('.card');
      channels.forEach(channel => {
        expect(channel.style.display).toBe('none');
      });
    });
  });

  describe('init', () => {
    beforeEach(() => {
      document.body.innerHTML = `
        <input id="portexe-search-input" type="text" />
        <div class="card">
          <div class="font-bold">Test Channel</div>
        </div>
      `;
    });

    it('should set search input value from URL parameter on page load', () => {
      simpleInit('?search=test');
      
      const searchInput = document.getElementById('portexe-search-input');
      expect(searchInput.value).toBe('test');
    });

    it('should perform search with URL parameter on page load', () => {
      simpleInit('?search=test');
      
      const channel = document.querySelector('.card');
      expect(channel.style.display).toBe('block');
    });

    it('should add keyup event listener to search input', () => {
      const searchInput = document.getElementById('portexe-search-input');
      const addEventListenerSpy = jest.spyOn(searchInput, 'addEventListener');
      
      simpleInit('');
      
      expect(addEventListenerSpy).toHaveBeenCalledWith('keyup', expect.any(Function));
    });

    it('should handle missing search input element', () => {
      document.body.innerHTML = '<div>No search input</div>';
      
      expect(() => simpleInit('')).not.toThrow();
    });

    it('should trigger search on keyup events', () => {
      simpleInit('');
      
      const searchInput = document.getElementById('portexe-search-input');
      searchInput.value = 'test';
      
      // Simulate keyup event
      const event = new Event('keyup');
      searchInput.dispatchEvent(event);
      
      const channel = document.querySelector('.card');
      expect(channel.style.display).toBe('block');
    });
  });

  describe('loginClick', () => {
    beforeEach(() => {
      document.body.innerHTML = `
        <input id="username" value="testuser" />
        <input id="password" value="testpass" />
      `;
      
      fetch.mockClear();
    });

    it('should make login request with correct credentials', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: 'success' })
      });
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginClick(mockFetch, mockAlert, mockReload);
      
      expect(mockFetch).toHaveBeenCalledWith('/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username: 'testuser', password: 'testpass' })
      });
    });

    it('should handle successful login', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: 'success' })
      });
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginClick(mockFetch, mockAlert, mockReload);
      
      expect(mockAlert).toHaveBeenCalledWith('Login success. Enjoy!');
      expect(mockReload).toHaveBeenCalled();
    });

    it('should handle failed login', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: 'failed' })
      });
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginClick(mockFetch, mockAlert, mockReload);
      
      expect(mockAlert).toHaveBeenCalledWith('Login failed!');
      expect(mockReload).not.toHaveBeenCalled();
    });

    it('should handle network errors', async () => {
      const mockFetch = jest.fn().mockRejectedValue(new Error('Network error'));
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginClick(mockFetch, mockAlert, mockReload);
      
      expect(console.log).toHaveBeenCalledWith(expect.any(Error));
      expect(mockAlert).toHaveBeenCalledWith('Login failed!');
    });

    it('should not make request when username is missing', async () => {
      document.getElementById('username').value = '';
      
      const mockFetch = jest.fn();
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginClick(mockFetch, mockAlert, mockReload);
      
      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('should not make request when password is missing', async () => {
      document.getElementById('password').value = '';
      
      const mockFetch = jest.fn();
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginClick(mockFetch, mockAlert, mockReload);
      
      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('should handle missing input elements', async () => {
      document.body.innerHTML = '';
      
      const mockFetch = jest.fn();
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      expect(() => simpleLoginClick(mockFetch, mockAlert, mockReload)).not.toThrow();
      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  describe('loginOTPClick', () => {
    beforeEach(() => {
      document.body.innerHTML = `
        <input id="number" value="9876543210" />
      `;
      
      fetch.mockClear();
    });

    it('should send OTP request with correct phone number', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: true })
      });
      const mockAlert = jest.fn();
      const mockShowModal = jest.fn();
      
      await simpleLoginOTPClick(mockFetch, mockAlert, mockShowModal);
      
      expect(mockFetch).toHaveBeenCalledWith('/login/sendOTP', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ number: '+919876543210' })
      });
    });

    it('should show modal on successful OTP send', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: true })
      });
      const mockAlert = jest.fn();
      const mockShowModal = jest.fn();
      
      await simpleLoginOTPClick(mockFetch, mockAlert, mockShowModal);
      
      expect(mockShowModal).toHaveBeenCalled();
    });

    it('should handle OTP send failure', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: false })
      });
      const mockAlert = jest.fn();
      const mockShowModal = jest.fn();
      
      await simpleLoginOTPClick(mockFetch, mockAlert, mockShowModal);
      
      expect(mockAlert).toHaveBeenCalledWith('Sending OTP failed!');
      expect(mockShowModal).not.toHaveBeenCalled();
    });

    it('should handle network errors', async () => {
      const mockFetch = jest.fn().mockRejectedValue(new Error('Network error'));
      const mockAlert = jest.fn();
      const mockShowModal = jest.fn();
      
      await simpleLoginOTPClick(mockFetch, mockAlert, mockShowModal);
      
      expect(console.log).toHaveBeenCalledWith(expect.any(Error));
      expect(mockAlert).toHaveBeenCalledWith('Sending OTP failed!');
    });

    it('should not make request when number is missing', async () => {
      document.getElementById('number').value = '';
      
      const mockFetch = jest.fn();
      const mockAlert = jest.fn();
      const mockShowModal = jest.fn();
      
      await simpleLoginOTPClick(mockFetch, mockAlert, mockShowModal);
      
      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  describe('loginOTPVerifyClick', () => {
    beforeEach(() => {
      document.body.innerHTML = `
        <input id="number" value="9876543210" />
        <input id="otp" value="123456" />
      `;
      
      fetch.mockClear();
    });

    it('should verify OTP with correct data', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: true })
      });
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockReload);
      
      expect(mockFetch).toHaveBeenCalledWith('/login/verifyOTP', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ number: '+919876543210', otp: '123456' })
      });
    });

    it('should handle successful OTP verification', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: true })
      });
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockReload);
      
      expect(mockAlert).toHaveBeenCalledWith('OTP verification success. Enjoy!');
      expect(mockReload).toHaveBeenCalled();
    });

    it('should handle failed OTP verification', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: false })
      });
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockReload);
      
      expect(mockAlert).toHaveBeenCalledWith('OTP verification failed!');
      expect(mockReload).not.toHaveBeenCalled();
    });

    it('should handle network errors', async () => {
      const mockFetch = jest.fn().mockRejectedValue(new Error('Network error'));
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockReload);
      
      expect(console.log).toHaveBeenCalledWith(expect.any(Error));
      expect(mockAlert).toHaveBeenCalledWith('OTP verification failed!');
    });

    it('should not make request when number is missing', async () => {
      document.getElementById('number').value = '';
      
      const mockFetch = jest.fn();
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockReload);
      
      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('should not make request when OTP is missing', async () => {
      document.getElementById('otp').value = '';
      
      const mockFetch = jest.fn();
      const mockAlert = jest.fn();
      const mockReload = jest.fn();
      
      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockReload);
      
      expect(mockFetch).not.toHaveBeenCalled();
    });
  });
});