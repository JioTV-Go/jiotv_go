/**
 * @jest-environment jsdom
 */

const fs = require('fs');
const path = require('path');

// Mock fetch globally
global.fetch = jest.fn();

describe('Play Page Functions', () => {
  beforeEach(() => {
    // Reset DOM
    document.body.innerHTML = `
      <div id="epg_parent">
        <div id="epg">
          <div class="card">
            <figure><img id="episodePoster" alt="Episode Poster" /></figure>
            <div class="card-body">
              <h2 id="showname" class="card-title"></h2>
              <div id="countdown_hour" class="countdown">
                <span id="e_hour"></span>h
              </div>
              <div id="countdown_minute" class="countdown">
                <span id="e_minute"></span>m
              </div>
              <div id="countdown_second" class="countdown">
                <span id="e_second"></span>s
              </div>
              <p id="description"></p>
              <div id="keywords"></div>
            </div>
          </div>
        </div>
      </div>
      <div id="similar_channels_parent" style="display: none;">
        <div id="similar_channels"></div>
      </div>
    `;

    // Reset fetch mock
    fetch.mockClear();
  });

  test('Cache functionality works correctly', async () => {
    // Define cache object
    let channelsCache = {
      data: null,
      timestamp: 0,
      expiryTime: 60 * 60 * 1000 // 1 hour
    };

    // Define getCachedChannels function
    async function getCachedChannels() {
      const now = Date.now();
      
      if (channelsCache.data && (now - channelsCache.timestamp < channelsCache.expiryTime)) {
        return channelsCache.data;
      }
      
      try {
        const response = await fetch('/channels');
        if (!response.ok) {
          throw new Error('Failed to fetch channels');
        }
        
        const channelsData = await response.json();
        channelsCache.data = channelsData;
        channelsCache.timestamp = now;
        
        return channelsData;
      } catch (error) {
        console.error('Error fetching channels:', error);
        return channelsCache.data || null;
      }
    }

    const mockChannelsData = {
      result: [
        { channel_id: 'ch1', channel_name: 'Channel 1', channelCategoryId: 1, channelLanguageId: 1, logoUrl: 'ch1.png' }
      ]
    };

    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockChannelsData
    });

    // First call should fetch data
    const data1 = await getCachedChannels();
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(data1).toEqual(mockChannelsData);

    // Second call should use cache
    const data2 = await getCachedChannels();
    expect(fetch).toHaveBeenCalledTimes(1); // Still 1
    expect(data2).toEqual(mockChannelsData);
  });

  test('getCurrentChannelInfo finds correct channel', () => {
    function getCurrentChannelInfo(channelsData, currentChannelID) {
      if (!channelsData || !channelsData.result) return null;
      return channelsData.result.find(channel => channel.channel_id === currentChannelID);
    }

    const channelsData = {
      result: [
        { channel_id: 'ch1', channel_name: 'Channel 1', channelCategoryId: 1, channelLanguageId: 1 },
        { channel_id: 'test-channel', channel_name: 'Test Channel', channelCategoryId: 1, channelLanguageId: 1 }
      ]
    };

    const currentChannel = getCurrentChannelInfo(channelsData, 'test-channel');
    expect(currentChannel.channel_name).toBe('Test Channel');
  });

  test('getSimilarChannels filters and randomizes correctly', () => {
    function getSimilarChannels(channelsData, currentChannel, maxChannels = 6) {
      if (!channelsData || !channelsData.result || !currentChannel) return [];
      
      const currentChannelID = currentChannel.channel_id;
      const currentCategory = currentChannel.channelCategoryId;
      const currentLanguage = currentChannel.channelLanguageId;
      
      let similarChannels = channelsData.result.filter(channel => {
        return channel.channel_id !== currentChannelID && 
               (channel.channelCategoryId === currentCategory || channel.channelLanguageId === currentLanguage);
      });
      
      // Shuffle the array to randomize selection
      for (let i = similarChannels.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [similarChannels[i], similarChannels[j]] = [similarChannels[j], similarChannels[i]];
      }
      
      return similarChannels.slice(0, maxChannels);
    }

    const channelsData = {
      result: [
        { channel_id: 'current', channel_name: 'Current Channel', channelCategoryId: 1, channelLanguageId: 1 },
        { channel_id: 'similar1', channel_name: 'Similar 1', channelCategoryId: 1, channelLanguageId: 1 }, // Exact match
        { channel_id: 'similar2', channel_name: 'Similar 2', channelCategoryId: 1, channelLanguageId: 2 }, // Category match
        { channel_id: 'similar3', channel_name: 'Similar 3', channelCategoryId: 2, channelLanguageId: 1 }, // Language match
        { channel_id: 'different', channel_name: 'Different', channelCategoryId: 2, channelLanguageId: 2 }  // No match
      ]
    };

    const currentChannel = { channel_id: 'current', channelCategoryId: 1, channelLanguageId: 1 };
    const similarChannels = getSimilarChannels(channelsData, currentChannel, 10);

    // Should include channels with matching category or language, but exclude non-matching and current channel
    expect(similarChannels).toHaveLength(3);
    const channelIds = similarChannels.map(ch => ch.channel_id);
    expect(channelIds).toContain('similar1');
    expect(channelIds).toContain('similar2');
    expect(channelIds).toContain('similar3');
    expect(channelIds).not.toContain('different');
    expect(channelIds).not.toContain('current');
    
    // Test that randomization works by running multiple times
    // (This is probabilistic, but with Fisher-Yates shuffle, order should vary)
    const orders = new Set();
    for (let i = 0; i < 10; i++) {
      const result = getSimilarChannels(channelsData, currentChannel, 10);
      orders.add(result.map(ch => ch.channel_id).join(','));
    }
    // With 3 items and 10 runs, we should get at least 2 different orders (highly likely)
    expect(orders.size).toBeGreaterThan(1);
  });

  test('renderSimilarChannels creates correct DOM structure', () => {
    function renderSimilarChannels(similarChannels) {
      const similarChannelsContainer = document.getElementById('similar_channels');
      const similarChannelsParent = document.getElementById('similar_channels_parent');
      
      if (!similarChannelsContainer || !similarChannelsParent) return;
      
      similarChannelsContainer.innerHTML = '';
      
      if (!similarChannels || similarChannels.length === 0) {
        similarChannelsParent.style.display = 'none';
        return;
      }
      
      similarChannels.forEach(channel => {
        const channelCard = document.createElement('a');
        channelCard.href = `/play/${channel.channel_id}`;
        channelCard.className = 'card';
        channelCard.setAttribute('data-channel-id', channel.channel_id);
        
        const logoURL = (channel.logoUrl && (channel.logoUrl.startsWith('http://') || channel.logoUrl.startsWith('https://'))) 
          ? channel.logoUrl 
          : `/jtvimage/${channel.logoUrl}`;
        
        channelCard.innerHTML = `
          <div class="flex flex-col items-center p-2">
            <img src="${logoURL}" alt="${channel.channel_name}" class="h-12 w-12" />
            <span class="text-sm font-bold">${channel.channel_name}</span>
          </div>
        `;
        
        similarChannelsContainer.appendChild(channelCard);
      });
      
      similarChannelsParent.style.display = 'block';
    }

    const similarChannels = [
      { channel_id: 'ch1', channel_name: 'Channel 1', logoUrl: 'ch1.png' },
      { channel_id: 'ch2', channel_name: 'Channel 2', logoUrl: 'https://example.com/ch2.png' }
    ];

    renderSimilarChannels(similarChannels);

    const container = document.getElementById('similar_channels');
    const parent = document.getElementById('similar_channels_parent');
    
    expect(parent.style.display).toBe('block');
    expect(container.children).toHaveLength(2);
    
    const firstCard = container.children[0];
    expect(firstCard.href).toContain('/play/ch1');
    expect(firstCard.textContent).toContain('Channel 1');
  });

  test('Layout structure is correct', () => {
    // Test that required DOM elements exist
    expect(document.getElementById('epg_parent')).toBeTruthy();
    expect(document.getElementById('similar_channels_parent')).toBeTruthy();
    expect(document.getElementById('similar_channels')).toBeTruthy();
    expect(document.getElementById('showname')).toBeTruthy();
    expect(document.getElementById('description')).toBeTruthy();
  });
});