// now_tiles.js — lazily fills each channel card's "Now:" title with the show
// currently airing. There is no bulk now-playing feed, only per-channel EPG, so
// we fetch only for cards scrolled into view to avoid hammering the upstream.
// window.hydrateNowTiles() can be called again after new cards are added to the
// DOM (e.g. the play page's Similar Channels strip) to observe them too.
(() => {
  const currentShow = (epgData) => {
    const now = Date.now();
    const shows = (epgData && epgData.epg) || [];
    return shows.find(
      (show) => new Date(show.startEpoch) <= now && now < new Date(show.endEpoch)
    );
  };

  const loaded = new WeakSet();
  let observer = null;

  const fillCard = async (card) => {
    const channelId = card.dataset.channelId;
    const titleEl = card.querySelector(".now-title");
    const lineEl = card.querySelector(".now-playing");
    if (!channelId || !titleEl || !lineEl) return;

    try {
      const epgData = await getJSON(`/epg/${channelId}/0`);
      const show = currentShow(epgData);
      if (show && show.showname) {
        titleEl.textContent = show.showname;
        lineEl.classList.remove("hidden");
      }
    } catch (error) {
      // Missing EPG for a channel is expected; leave the line hidden.
    }
  };

  const hydrate = () => {
    const cards = document.querySelectorAll("a.card[data-channel-id]");
    if (!cards.length) return;

    if (!("IntersectionObserver" in window)) {
      cards.forEach((card) => {
        if (!loaded.has(card)) {
          loaded.add(card);
          fillCard(card);
        }
      });
      return;
    }

    if (!observer) {
      observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (!entry.isIntersecting || loaded.has(entry.target)) return;
            loaded.add(entry.target);
            observer.unobserve(entry.target);
            fillCard(entry.target);
          });
        },
        { rootMargin: "200px" }
      );
    }

    cards.forEach((card) => {
      if (!loaded.has(card)) observer.observe(card);
    });
  };

  window.hydrateNowTiles = hydrate;
  document.addEventListener("DOMContentLoaded", hydrate);
})();
