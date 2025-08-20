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

const url = new URL(window.location.href);
// do regex to get channelID
const channelID = url.pathname.match(/\/play\/(.*)/)[1];
const offset = 0;

// Cache for channels data with 1-hour expiry
let channelsCache = {
    data: null,
    timestamp: 0,
    expiryTime: 60 * 60 * 1000 // 1 hour in milliseconds
};

// Function to get cached channels or fetch new ones
async function getCachedChannels() {
    const now = Date.now();

    // Check if cache is valid
    if (channelsCache.data && (now - channelsCache.timestamp < channelsCache.expiryTime)) {
        return channelsCache.data;
    }

    try {
        const channelsData = await getJSON('/channels');
        
        // Update cache
        channelsCache.data = channelsData;
        channelsCache.timestamp = now;

        return channelsData;
    } catch (error) {
        console.error('Error fetching channels:', error);
        // Return cached data if available, even if expired
        return channelsCache.data || null;
    }
}

// Function to get current channel info from channels list
function getCurrentChannelInfo(channelsData, currentChannelID) {
    if (!channelsData || !channelsData.result) return null;

    return channelsData.result.find(channel => channel.channel_id === currentChannelID);
}

// Function to get random similar channels based on category and language
function getSimilarChannels(channelsData, currentChannel, maxChannels = 12) {
    if (!channelsData || !channelsData.result || !currentChannel) return [];

    const currentChannelID = currentChannel.channel_id;
    const currentCategory = currentChannel.channelCategoryId;
    const currentLanguage = currentChannel.channelLanguageId;

    // Filter channels by same category and language, excluding current channel
    let similarChannels = channelsData.result.filter(channel => {
        return channel.channel_id !== currentChannelID &&
            (channel.channelCategoryId === currentCategory && channel.channelLanguageId === currentLanguage);
    });

    // Shuffle the array to randomize selection
    for (let i = similarChannels.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [similarChannels[i], similarChannels[j]] = [similarChannels[j], similarChannels[i]];
    }

    return similarChannels.slice(0, maxChannels);
}

// Function to render similar channels
function renderSimilarChannels(similarChannels) {
    const elements = safeGetElementsById(['similar_channels', 'similar_channels_parent']);
    const { similar_channels: similarChannelsContainer, similar_channels_parent: similarChannelsParent } = elements;

    if (!similarChannelsContainer || !similarChannelsParent) return;

    // Clear existing content
    similarChannelsContainer.innerHTML = '';

    if (!similarChannels || similarChannels.length === 0) {
        similarChannelsParent.style.display = 'none';
        return;
    }

    // Create channel cards
    similarChannels.forEach(channel => {
        // Determine logo URL (handle both custom and regular channels)
        const logoURL = (channel.logoUrl && (channel.logoUrl.startsWith('http://') || channel.logoUrl.startsWith('https://')))
            ? channel.logoUrl
            : `/jtvimage/${channel.logoUrl}`;

        const channelCard = createElement('a', {
            href: `/play/${channel.channel_id}`,
            className: 'card relative border border-primary shadow-lg hover:shadow-xl hover:bg-base-300 transition-all duration-200 ease-in-out scale-100 hover:scale-105 group',
            'data-channel-id': channel.channel_id,
            tabindex: '0'
        }, '', `
            <div class="flex flex-col items-center p-2">
                <img
                    src="${logoURL}"
                    loading="lazy"
                    alt="${channel.channel_name}"
                    class="h-12 w-12 rounded-full bg-gray-200"
                    onerror="this.style.display='none'"
                />
                <span class="text-sm font-bold mt-1 text-center line-clamp-2">${channel.channel_name}</span>
            </div>
        `);

        similarChannelsContainer.appendChild(channelCard);
    });

    similarChannelsParent.style.display = 'block';
}

// Function to load similar channels
async function loadSimilarChannels() {
    try {
        const channelsData = await getCachedChannels();
        if (!channelsData) return;

        const currentChannel = getCurrentChannelInfo(channelsData, channelID);
        if (!currentChannel) return;

        const similarChannels = getSimilarChannels(channelsData, currentChannel);
        renderSimilarChannels(similarChannels);
    } catch (error) {
        console.error('Error loading similar channels:', error);
    }
}


function updateEPG(epgData) {
    const shows = getCurrentAndNextTwoShows(epgData);
    const elements = safeGetElementsById(['showname', 'description', 'episodePoster', 'keywords']);
    const { showname: shownameElement, description: descriptionElement, episodePoster: episodePosterElement, keywords: keywordsElement } = elements;
    
    if (shows.length === 0) return;
    
    if (shownameElement) shownameElement.textContent = shows[0].showname;
    if (descriptionElement) descriptionElement.textContent = shows[0].description;
    
    if (episodePosterElement) {
        const posterUrl = new URL("/jtvposter/", window.location.href);
        posterUrl.pathname += shows[0].episodePoster;
        episodePosterElement.src = posterUrl.href;
    }

    if (keywordsElement && shows[0].keywords) {
        // Clear existing keywords
        keywordsElement.innerHTML = '';
        
        shows[0].keywords.forEach((keyword) => {
            const keywordElement = createElement('div', {
                className: 'badge badge-outline'
            }, keyword);
            keywordsElement.appendChild(keywordElement);
        });
    }

    const timerElements = safeGetElementsById(['e_hour', 'e_minute', 'e_second']);
    const { e_hour, e_minute, e_second } = timerElements;

    const endEpochTime = shows[0].endEpoch;
    function updateTimer() {
        const currentTime = new Date().getTime();
        const difference = endEpochTime - currentTime;

        if (difference <= 0) {
            clearInterval(timerInterval);
            const countdownElements = safeGetElementsById(['countdown_hour', 'countdown_minute']);
            const { countdown_hour, countdown_minute } = countdownElements;
            
            if (countdown_hour) countdown_hour.style.removeProperty('display');
            if (countdown_minute) countdown_minute.style.removeProperty('display');
            updateEPG(epgData);
            return;
        }

        const differenceDate = new Date(difference);
        const hours = differenceDate.getUTCHours();
        const minutes = differenceDate.getUTCMinutes();
        const seconds = differenceDate.getUTCSeconds();

        if (hours === 0) {
            const countdownHour = safeGetElementById('countdown_hour');
            if (countdownHour) countdownHour.style.display = 'none';
        } else {
            if (e_hour) e_hour.setAttribute('style', `--value:${hours.toString().padStart(2, '0')};`);
        }
        
        if (hours === 0 && minutes === 0) {
            const countdownMinute = safeGetElementById('countdown_minute');
            if (countdownMinute) countdownMinute.style.display = 'none';
        } else {
            if (e_minute) e_minute.setAttribute('style', `--value:${minutes.toString().padStart(2, '0')};`);
        }
        
        if (e_second) e_second.setAttribute('style', `--value:${seconds.toString().padStart(2, '0')};`);
    }

    // Initial call to update the timer
    updateTimer();

    // Set the interval to update the timer every second
    const timerInterval = setInterval(updateTimer, 1000);
}

const epgParent = safeGetElementById('epg_parent');
if (epgParent) epgParent.style.display = 'none';

(async () => {
    // Load EPG data
    try {
        const epgData = await getJSON(`/epg/${channelID}/${offset}`);
        if (epgParent) epgParent.style.display = 'block';
        updateEPG(epgData);

        // Load similar channels
        await loadSimilarChannels();
    } catch (error) {
        console.error('Failed to fetch EPG data:', error);
    }
})();
