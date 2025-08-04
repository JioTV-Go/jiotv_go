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
    const epgResponse = await fetch(`/epg/${channelID}/${offset}`);

    if (!epgResponse.ok) {
        console.error('Failed to fetch EPG data');
        return;
    }

    const epgData = await epgResponse.json();
    epgParent.style.display = 'block';
    updateEPG(epgData);
})();
