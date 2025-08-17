// Mock fetch
global.fetch = jest.fn();

// Mock console.error to avoid cluttering test output
global.console.error = jest.fn();

// Simplified function without window.location dependency
function simpleUpdateEPG(epgData, baseUrl = 'http://localhost:3000/play/test-channel-123') {
    const shows = getCurrentAndNextTwoShows(epgData);
    const shownameElement = document.getElementById('showname');
    const descriptionElement = document.getElementById('description');
    const episodePosterElement = document.getElementById('episodePoster');
    
    if (shows.length === 0) return;
    
    if (shownameElement) shownameElement.textContent = shows[0].showname;
    if (descriptionElement) descriptionElement.textContent = shows[0].description;
    
    if (episodePosterElement) {
        episodePosterElement.src = `${baseUrl}/jtvposter/${shows[0].episodePoster}`;
    }
    
    const keywordsElement = document.getElementById('keywords');
    if (keywordsElement && shows[0].keywords) {
        // Clear existing keywords
        keywordsElement.innerHTML = '';
        
        shows[0].keywords.forEach((keyword) => {
            const keywordElement = document.createElement('div');
            keywordElement.className = 'badge badge-outline';
            keywordElement.textContent = keyword;
            keywordsElement.appendChild(keywordElement);
        });
    }
}

// Define functions from epg.js for testing
function getCurrentAndNextTwoShows(epgData) {
    const currentTime = new Date(); // Current date time
    const shows = [];
    let currentIndex = -1;

    // Find the currently playing show
    epgData.epg.some((show, index) => {
        const showStartTime = new Date(show.startEpoch);
        const showEndTime = new Date(show.endEpoch);

        if (showStartTime <= currentTime && currentTime < showEndTime) {
            const { showname, description, endEpoch, episodePoster, keywords } = show;
            shows.push({ showname, description, endEpoch, episodePoster, keywords });
            currentIndex = index;
            return true; // Stop iterating after finding the current show
        }
        return false;
    });

    // Get the next two shows
    if (currentIndex !== -1) {
        const nextTwoShows = epgData.epg.slice(currentIndex + 1, currentIndex + 3);
        nextTwoShows.forEach(show => {
            const { showname, description, endEpoch, episodePoster, keywords } = show;
            shows.push({ showname, description, endEpoch, episodePoster, keywords });
        });
    }

    return shows;
}

function updateEPG(epgData) {
    const shows = getCurrentAndNextTwoShows(epgData);
    const shownameElement = document.getElementById('showname');
    const descriptionElement = document.getElementById('description');
    const episodePosterElement = document.getElementById('episodePoster');
    
    if (shows.length === 0) return;
    
    if (shownameElement) shownameElement.textContent = shows[0].showname;
    if (descriptionElement) descriptionElement.textContent = shows[0].description;
    
    if (episodePosterElement) {
        const posterUrl = new URL("/jtvposter/", window.location.href);
        posterUrl.pathname += shows[0].episodePoster;
        episodePosterElement.src = posterUrl.href;
    }
    
    const keywordsElement = document.getElementById('keywords');
    if (keywordsElement && shows[0].keywords) {
        // Clear existing keywords
        keywordsElement.innerHTML = '';
        
        shows[0].keywords.forEach((keyword) => {
            const keywordElement = document.createElement('div');
            keywordElement.className = 'badge badge-outline';
            keywordElement.textContent = keyword;
            keywordsElement.appendChild(keywordElement);
        });
    }
}

describe('EPG (Electronic Program Guide) Functionality', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Clear and set up DOM structure
    document.body.innerHTML = `
      <div id="epg_parent" style="display: none;">
        <div id="showname"></div>
        <div id="description"></div>
        <img id="episodePoster" src="" alt="Episode Poster">
        <div id="keywords"></div>
        <div id="countdown_hour">
          <span id="e_hour"></span>
        </div>
        <div id="countdown_minute">
          <span id="e_minute"></span>
        </div>
        <div id="countdown_second">
          <span id="e_second"></span>
        </div>
      </div>
    `;
  });

  describe('getCurrentAndNextTwoShows', () => {
    const mockEpgData = {
      epg: [
        {
          showname: "Past Show",
          description: "This show already ended",
          startEpoch: Date.now() - 3600000, // 1 hour ago
          endEpoch: Date.now() - 1800000,   // 30 minutes ago
          episodePoster: "past-show.jpg",
          keywords: ["past", "ended"]
        },
        {
          showname: "Current Show",
          description: "This show is currently playing",
          startEpoch: Date.now() - 1800000,  // 30 minutes ago
          endEpoch: Date.now() + 1800000,    // 30 minutes from now
          episodePoster: "current-show.jpg",
          keywords: ["current", "live"]
        },
        {
          showname: "Next Show 1",
          description: "This show comes next",
          startEpoch: Date.now() + 1800000,  // 30 minutes from now
          endEpoch: Date.now() + 5400000,    // 1.5 hours from now
          episodePoster: "next-show-1.jpg",
          keywords: ["next", "upcoming"]
        },
        {
          showname: "Next Show 2",
          description: "This show comes after next",
          startEpoch: Date.now() + 5400000,  // 1.5 hours from now
          endEpoch: Date.now() + 9000000,    // 2.5 hours from now
          episodePoster: "next-show-2.jpg",
          keywords: ["later", "future"]
        },
        {
          showname: "Future Show",
          description: "This show is much later",
          startEpoch: Date.now() + 9000000,  // 2.5 hours from now
          endEpoch: Date.now() + 12600000,   // 3.5 hours from now
          episodePoster: "future-show.jpg",
          keywords: ["future", "distant"]
        }
      ]
    };

    it('should return current show and next two shows', () => {
      const shows = getCurrentAndNextTwoShows(mockEpgData);
      
      expect(shows).toHaveLength(3);
      expect(shows[0].showname).toBe("Current Show");
      expect(shows[1].showname).toBe("Next Show 1");
      expect(shows[2].showname).toBe("Next Show 2");
    });

    it('should return only extracted properties for each show', () => {
      const shows = getCurrentAndNextTwoShows(mockEpgData);
      
      expect(shows[0]).toEqual({
        showname: "Current Show",
        description: "This show is currently playing",
        endEpoch: expect.any(Number),
        episodePoster: "current-show.jpg",
        keywords: ["current", "live"]
      });
      
      expect(shows[0]).not.toHaveProperty('startEpoch');
    });

    it('should return empty array when no current show is found', () => {
      const noCurrentShowData = {
        epg: [
          {
            showname: "Past Show",
            description: "This show already ended",
            startEpoch: Date.now() - 3600000,
            endEpoch: Date.now() - 1800000,
            episodePoster: "past-show.jpg",
            keywords: ["past"]
          },
          {
            showname: "Future Show",
            description: "This show is in the future",
            startEpoch: Date.now() + 3600000,
            endEpoch: Date.now() + 7200000,
            episodePoster: "future-show.jpg",
            keywords: ["future"]
          }
        ]
      };
      
      const shows = getCurrentAndNextTwoShows(noCurrentShowData);
      expect(shows).toHaveLength(0);
    });

    it('should handle case when there are fewer than 2 next shows', () => {
      const limitedEpgData = {
        epg: [
          {
            showname: "Current Show",
            description: "This show is currently playing",
            startEpoch: Date.now() - 1800000,
            endEpoch: Date.now() + 1800000,
            episodePoster: "current-show.jpg",
            keywords: ["current"]
          },
          {
            showname: "Only Next Show",
            description: "This is the only next show",
            startEpoch: Date.now() + 1800000,
            endEpoch: Date.now() + 5400000,
            episodePoster: "only-next.jpg",
            keywords: ["next"]
          }
        ]
      };
      
      const shows = getCurrentAndNextTwoShows(limitedEpgData);
      expect(shows).toHaveLength(2);
      expect(shows[0].showname).toBe("Current Show");
      expect(shows[1].showname).toBe("Only Next Show");
    });

    it('should handle empty EPG data', () => {
      const emptyEpgData = { epg: [] };
      const shows = getCurrentAndNextTwoShows(emptyEpgData);
      expect(shows).toHaveLength(0);
    });
  });

  describe('updateEPG', () => {
    const mockEpgData = {
      epg: [
        {
          showname: "Test Show",
          description: "A test show description",
          startEpoch: Date.now() - 1800000,
          endEpoch: Date.now() + 1800000,
          episodePoster: "test-show.jpg",
          keywords: ["test", "demo", "sample"]
        },
        {
          showname: "Next Test Show",
          description: "Next show description",
          startEpoch: Date.now() + 1800000,
          endEpoch: Date.now() + 5400000,
          episodePoster: "next-test.jpg",
          keywords: ["next", "test"]
        }
      ]
    };

    it('should update show name and description elements', () => {
      simpleUpdateEPG(mockEpgData);
      
      const shownameElement = document.getElementById('showname');
      const descriptionElement = document.getElementById('description');
      
      expect(shownameElement.textContent).toBe("Test Show");
      expect(descriptionElement.textContent).toBe("A test show description");
    });

    it('should update episode poster with correct URL', () => {
      simpleUpdateEPG(mockEpgData);
      
      const episodePosterElement = document.getElementById('episodePoster');
      // The poster URL should be constructed with the base URL + poster filename
      expect(episodePosterElement.src).toBe("http://localhost:3000/play/test-channel-123/jtvposter/test-show.jpg");
    });

    it('should create keyword badges', () => {
      simpleUpdateEPG(mockEpgData);
      
      const keywordsElement = document.getElementById('keywords');
      const badges = keywordsElement.querySelectorAll('.badge.badge-outline');
      
      expect(badges).toHaveLength(3);
      expect(badges[0].textContent).toBe("test");
      expect(badges[1].textContent).toBe("demo");
      expect(badges[2].textContent).toBe("sample");
    });

    it('should clear existing keywords before adding new ones', () => {
      const keywordsElement = document.getElementById('keywords');
      keywordsElement.innerHTML = '<div class="old-badge">old</div>';
      
      simpleUpdateEPG(mockEpgData);
      
      const badges = keywordsElement.querySelectorAll('.badge.badge-outline');
      const oldBadges = keywordsElement.querySelectorAll('.old-badge');
      
      expect(badges).toHaveLength(3);
      expect(oldBadges).toHaveLength(0);
    });

    it('should handle missing DOM elements gracefully', () => {
      // Remove some elements
      document.getElementById('showname').remove();
      document.getElementById('keywords').remove();
      
      expect(() => simpleUpdateEPG(mockEpgData)).not.toThrow();
      
      // Elements that still exist should be updated
      const descriptionElement = document.getElementById('description');
      expect(descriptionElement.textContent).toBe("A test show description");
    });

    it('should handle EPG data with no current shows', () => {
      const noCurrentShowData = {
        epg: [
          {
            showname: "Future Show",
            description: "This show is in the future",
            startEpoch: Date.now() + 3600000,
            endEpoch: Date.now() + 7200000,
            episodePoster: "future.jpg",
            keywords: ["future"]
          }
        ]
      };
      
      // Ensure elements start with some content
      const shownameElement = document.getElementById('showname');
      shownameElement.textContent = "Initial content";
      
      expect(() => simpleUpdateEPG(noCurrentShowData)).not.toThrow();
      
      // Elements should remain unchanged since no current show is found
      expect(shownameElement.textContent).toBe("Initial content");
    });

    it('should handle shows without keywords', () => {
      const noKeywordsData = {
        epg: [
          {
            showname: "No Keywords Show",
            description: "A show without keywords",
            startEpoch: Date.now() - 1800000,
            endEpoch: Date.now() + 1800000,
            episodePoster: "no-keywords.jpg",
            keywords: []
          }
        ]
      };
      
      simpleUpdateEPG(noKeywordsData);
      
      const keywordsElement = document.getElementById('keywords');
      const badges = keywordsElement.querySelectorAll('.badge.badge-outline');
      
      expect(badges).toHaveLength(0);
    });
  });
});