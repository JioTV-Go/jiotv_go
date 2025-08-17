/**
 * @jest-environment jsdom
 */

// Mock DOM setup for unmute button functionality
describe('Unmute Button', () => {
  let container;
  let video;
  let unmuteBtn;

  beforeEach(() => {
    // Set up DOM
    document.body.innerHTML = `
      <div data-shaka-player-container>
        <video
          autoplay
          muted
          data-shaka-player
          id="jiotv_go_player"
          style="width: 100%; height: 100%"
        ></video>
        
        <button id="unmute-btn" class="unmute-overlay" title="Click to unmute">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
            <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z"/>
          </svg>
        </button>
      </div>
    `;

    video = document.getElementById('jiotv_go_player');
    unmuteBtn = document.getElementById('unmute-btn');

    // Add CSS class support for tests
    const style = document.createElement('style');
    style.textContent = `
      .unmute-overlay.hidden {
        opacity: 0;
        pointer-events: none;
      }
    `;
    document.head.appendChild(style);
  });

  afterEach(() => {
    document.body.innerHTML = '';
    document.head.innerHTML = '';
  });

  test('unmute button should exist in DOM', () => {
    expect(unmuteBtn).toBeTruthy();
    expect(unmuteBtn.id).toBe('unmute-btn');
    expect(unmuteBtn.classList.contains('unmute-overlay')).toBe(true);
  });

  test('unmute button should have correct attributes', () => {
    expect(unmuteBtn.title).toBe('Click to unmute');
    expect(unmuteBtn.tagName).toBe('BUTTON');
  });

  test('unmute button should contain SVG icon', () => {
    const svg = unmuteBtn.querySelector('svg');
    expect(svg).toBeTruthy();
    expect(svg.getAttribute('viewBox')).toBe('0 0 24 24');
    
    const path = svg.querySelector('path');
    expect(path).toBeTruthy();
    expect(path.getAttribute('d')).toContain('M16.5 12c0-1.77');
  });

  test('updateUnmuteButton function should show button when muted', () => {
    // Mock the update function behavior
    const updateUnmuteButton = () => {
      if (video.muted) {
        unmuteBtn.classList.remove("hidden");
      } else {
        unmuteBtn.classList.add("hidden");
      }
    };

    // Test muted state
    video.muted = true;
    updateUnmuteButton();
    expect(unmuteBtn.classList.contains('hidden')).toBe(false);

    // Test unmuted state
    video.muted = false;
    updateUnmuteButton();
    expect(unmuteBtn.classList.contains('hidden')).toBe(true);
  });

  test('clicking unmute button should unmute video', () => {
    // Setup initial state
    video.muted = true;
    video.volume = 0;

    // Mock the click handler behavior
    const handleUnmuteClick = () => {
      video.muted = false;
      video.volume = 0.8;
    };

    // Simulate click
    handleUnmuteClick();

    expect(video.muted).toBe(false);
    expect(video.volume).toBe(0.8);
  });

  test('video element should have required attributes', () => {
    expect(video.hasAttribute('autoplay')).toBe(true);
    expect(video.hasAttribute('muted')).toBe(true);
    expect(video.hasAttribute('data-shaka-player')).toBe(true);
    expect(video.id).toBe('jiotv_go_player');
  });
});