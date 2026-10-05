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
            channel.channelCategoryId === currentCategory &&
            channel.channelLanguageId === currentLanguage;
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

    // Build the exact same tile the home page uses, so channel cards look and
    // behave identically everywhere. Untrusted text (name) is set via
    // textContent / attributes rather than interpolated into innerHTML.
    similarChannels.forEach(channel => {
        similarChannelsContainer.appendChild(buildChannelCard(channel));
    });

    similarChannelsParent.style.display = 'block';

    // Hydrate the newly-added cards the same way the home grid is hydrated:
    // favourite state, catchup routing/styling, and the lazy "Now:" line.
    if (typeof updateFavoriteButtonStates === 'function') updateFavoriteButtonStates();
    if (typeof applyCatchupToCards === 'function') applyCatchupToCards();
    if (typeof styleCatchupCards === 'function') styleCatchupCards();
    if (typeof window.hydrateNowTiles === 'function') window.hydrateNowTiles();
}

// Mirror of the server-rendered home channel card (web/views/channel_list.html).
// Keep the two in sync so tiles are consistent across the app.
function buildChannelCard(channel) {
    const logoURL = (channel.logoUrl && (channel.logoUrl.startsWith('http://') || channel.logoUrl.startsWith('https://')))
        ? channel.logoUrl
        : `/jtvimage/${channel.logoUrl}`;
    const id = channel.channel_id;
    const requiresSubscription = !!channel.requiresSubscription;

    const card = createElement('a', {
        href: `/play/${id}`,
        className: 'card relative overflow-hidden border border-primary bg-base-100 shadow-sm group',
        'data-channel-id': id,
        'data-channel-name': channel.channel_name,
        'data-catchup-available': channel.isCatchupAvailable ? 'true' : 'false',
        ...(requiresSubscription ? { 'data-requires-subscription': 'true' } : {}),
        tabindex: '0'
    });

    const tile = createElement('div', { className: 'logo-tile relative w-full aspect-[4/3] overflow-hidden' });
    const img = createElement('img', {
        src: logoURL,
        loading: 'lazy',
        alt: channel.channel_name,
        className: 'h-full w-full object-contain p-4 transition-transform duration-300 group-hover:scale-110'
    });
    img.addEventListener('error', () => { img.style.display = 'none'; });
    tile.appendChild(img);

    if (requiresSubscription) {
        const badge = createElement('span', {
            className: 'badge badge-sm absolute top-2 left-2 z-10 gap-1 border-0 text-[0.65rem] bg-base-100/85 text-base-content/80 shadow-sm backdrop-blur-sm',
            title: 'May require a separate subscription'
        }, '', '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="h-2.5 w-2.5" aria-hidden="true"><path fill-rule="evenodd" d="M12 1.5a5.25 5.25 0 0 0-5.25 5.25v3a3 3 0 0 0-3 3v6.75a3 3 0 0 0 3 3h10.5a3 3 0 0 0 3-3v-6.75a3 3 0 0 0-3-3v-3c0-2.9-2.35-5.25-5.25-5.25Zm3.75 8.25v-3a3.75 3.75 0 1 0-7.5 0v3h7.5Z" clip-rule="evenodd" /></svg>');
        badge.appendChild(document.createTextNode('Subscription'));
        tile.appendChild(badge);
    } else {
        tile.appendChild(createElement('span', { className: 'live-badge absolute top-2 left-2 z-10' }, 'LIVE'));
    }
    card.appendChild(tile);

    const body = createElement('div', { className: 'px-2.5 py-2.5 border-t border-base-300/40' });
    const nowLine = createElement('p', { className: 'now-playing hidden text-center leading-snug line-clamp-1' }, '',
        '<span class="font-semibold text-primary">Now:</span> <span class="now-title text-sm font-bold"></span>');
    body.appendChild(nowLine);
    card.appendChild(body);

    const favBtn = createElement('button', {
        id: `favorite-btn-${id}`,
        className: 'favorite-btn absolute btn btn-ghost btn-circle bg-base-100/80 backdrop-blur-sm top-2 right-2 z-10 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity duration-200',
        'aria-label': 'Add to favorites'
    }, '',
        `<svg id="star-icon-${id}" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557l-4.204-3.602a.562.562 0 0 1 .321-.988l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" /></svg><svg id="x-icon-${id}" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6 hidden"><path fill-rule="evenodd" d="M10.788 3.21c.448-1.077 1.976-1.077 2.424 0l2.082 5.006 5.404.434c1.164.093 1.636 1.545.749 2.305l-4.117 3.527 1.257 5.273c.271 1.136-.964 2.033-1.96 1.425L12 18.354 7.373 21.18c-.996.608-2.231-.29-1.96-1.425l1.257-5.273-4.117-3.527c-.887-.76-.415-2.212.749-2.305l5.404-.434 2.082-5.005Z" clip-rule="evenodd" /></svg>`);
    favBtn.addEventListener('click', (event) => {
        event.preventDefault();
        if (typeof toggleFavorite === 'function') toggleFavorite(id);
    });
    card.appendChild(favBtn);

    return card;
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
    if (descriptionElement) {
        descriptionElement.textContent = shows[0].description;
        updateNowPlayingDescription();
    }
    
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
