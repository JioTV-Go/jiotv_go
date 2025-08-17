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
        const response = await fetch('/channels');
        if (!response.ok) {
            throw new Error('Failed to fetch channels');
        }

        const channelsData = await response.json();

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
    const similarChannelsContainer = document.getElementById('similar_channels');
    const similarChannelsParent = document.getElementById('similar_channels_parent');

    if (!similarChannelsContainer || !similarChannelsParent) return;

    // Clear existing content
    similarChannelsContainer.innerHTML = '';

    if (!similarChannels || similarChannels.length === 0) {
        similarChannelsParent.style.display = 'none';
        return;
    }

    // Create channel cards
    similarChannels.forEach(channel => {
        const channelCard = document.createElement('a');
        channelCard.href = `/play/${channel.channel_id}`;
        channelCard.className = 'card relative border border-primary shadow-lg hover:shadow-xl hover:bg-base-300 transition-all duration-200 ease-in-out scale-100 hover:scale-105 group';
        channelCard.setAttribute('data-channel-id', channel.channel_id);
        channelCard.setAttribute('tabindex', '0');

        // Determine logo URL (handle both custom and regular channels)
        const logoURL = (channel.logoUrl && (channel.logoUrl.startsWith('http://') || channel.logoUrl.startsWith('https://')))
            ? channel.logoUrl
            : `/jtvimage/${channel.logoUrl}`;

        channelCard.innerHTML = `
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
        `;

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
    const shownameElement = document.getElementById('showname');
    const descriptionElement = document.getElementById('description');
    const episodePosterElement = document.getElementById('episodePoster');
    shownameElement.innerText = shows[0].showname;
    descriptionElement.innerText = shows[0].description;
    const posterUrl = new URL("/jtvposter/", window.location.href);
    posterUrl.pathname += shows[0].episodePoster;
    episodePosterElement.src = posterUrl.href;

    const keywordsElement = document.getElementById('keywords');
    const keywords = shows[0].keywords;
    keywords.forEach((keyword) => {
        const keywordElement = document.createElement('div');
        keywordElement.className = 'badge badge-outline';
        keywordElement.innerText = keyword;
        keywordsElement.appendChild(keywordElement);
    });

    const e_hour = document.getElementById('e_hour');
    const e_minute = document.getElementById('e_minute');
    const e_second = document.getElementById('e_second');

    const endEpochTime = shows[0].endEpoch;
    function updateTimer() {
        const currentTime = new Date().getTime();
        const difference = endEpochTime - currentTime;

        if (difference <= 0) {
            clearInterval(timerInterval);
            document.getElementById('countdown_hour').style.removeProperty('display');
            document.getElementById('countdown_minute').style.removeProperty('display');
            updateEPG(epgData);
            return;
        }

        const differenceDate = new Date(difference);
        const hours = differenceDate.getUTCHours();
        const minutes = differenceDate.getUTCMinutes();
        const seconds = differenceDate.getUTCSeconds();


        if (hours === 0) {
            document.getElementById('countdown_hour').style.display = 'none';
        } else {
            e_hour.setAttribute('style', `--value:${hours.toString().padStart(2, '0')};`);
        }
        if (hours === 0 && minutes === 0) {
            document.getElementById('countdown_minute').style.display = 'none';
        } else {
            e_minute.setAttribute('style', `--value:${minutes.toString().padStart(2, '0')};`);
        }
        e_second.setAttribute('style', `--value:${seconds.toString().padStart(2, '0')};`);
    }

    // Initial call to update the timer
    updateTimer();

    // Set the interval to update the timer every second
    const timerInterval = setInterval(updateTimer, 1000);
}

const epgParent = document.getElementById('epg_parent');
epgParent.style.display = 'none';

(async () => {
    // Load EPG data
    const epgResponse = await fetch(`/epg/${channelID}/${offset}`);

    if (!epgResponse.ok) {
        console.error('Failed to fetch EPG data');
        return;
    }

    const epgData = await epgResponse.json();
    epgParent.style.display = 'block';
    updateEPG(epgData);

    // Load similar channels
    await loadSimilarChannels();
})();
