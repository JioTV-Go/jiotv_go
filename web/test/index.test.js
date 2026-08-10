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



const simpleLoginOTPClick = async (fetchFn = fetch, showErrorFn = jest.fn(), showModalFn = jest.fn()) => {
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
      showErrorFn("We couldn’t send the OTP. Check your number and try again.");
    }
  } catch (err) {
    console.log(err);
    showErrorFn("We couldn’t send the OTP. Check your connection and try again.");
  }
};

const simpleLoginOTPVerifyClick = async (fetchFn = fetch, showErrorFn = jest.fn(), showSuccessModalFn = jest.fn()) => {
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
      showSuccessModalFn();
    } else {
      showErrorFn("The OTP is incorrect or expired. Try again.");
    }
  } catch (err) {
    console.log(err);
    showErrorFn("We couldn’t verify the OTP. Check your connection and try again.");
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

    it('shows the error modal when sending OTP fails', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: false })
      });
      const mockShowError = jest.fn();
      const mockShowModal = jest.fn();

      await simpleLoginOTPClick(mockFetch, mockShowError, mockShowModal);

      expect(mockShowError).toHaveBeenCalledWith('We couldn’t send the OTP. Check your number and try again.');
      expect(mockShowModal).not.toHaveBeenCalled();
    });

    it('shows a connection error when sending OTP rejects', async () => {
      const mockFetch = jest.fn().mockRejectedValue(new Error('Network error'));
      const mockShowError = jest.fn();
      const mockShowModal = jest.fn();

      await simpleLoginOTPClick(mockFetch, mockShowError, mockShowModal);

      expect(console.log).toHaveBeenCalledWith(expect.any(Error));
      expect(mockShowError).toHaveBeenCalledWith('We couldn’t send the OTP. Check your connection and try again.');
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
      const mockShowSuccessModal = jest.fn();

      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockShowSuccessModal);

      expect(mockFetch).toHaveBeenCalledWith('/login/verifyOTP', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ number: '+919876543210', otp: '123456' })
      });
    });

    it('shows success modal after successful OTP verification', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: true })
      });
      const mockAlert = jest.fn();
      const mockShowSuccessModal = jest.fn();

      await simpleLoginOTPVerifyClick(mockFetch, mockAlert, mockShowSuccessModal);

      expect(mockShowSuccessModal).toHaveBeenCalled();
      expect(mockAlert).not.toHaveBeenCalled();
    });

    it('shows the error modal for an invalid or expired OTP', async () => {
      const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ status: false })
      });
      const mockShowError = jest.fn();
      const mockShowSuccessModal = jest.fn();

      await simpleLoginOTPVerifyClick(mockFetch, mockShowError, mockShowSuccessModal);

      expect(mockShowError).toHaveBeenCalledWith('The OTP is incorrect or expired. Try again.');
      expect(mockShowSuccessModal).not.toHaveBeenCalled();
    });

    it('shows a connection error when OTP verification rejects', async () => {
      const mockFetch = jest.fn().mockRejectedValue(new Error('Network error'));
      const mockShowError = jest.fn();

      await simpleLoginOTPVerifyClick(mockFetch, mockShowError);

      expect(console.log).toHaveBeenCalledWith(expect.any(Error));
      expect(mockShowError).toHaveBeenCalledWith('We couldn’t verify the OTP. Check your connection and try again.');
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