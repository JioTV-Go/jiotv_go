const qualityRadios = document.querySelectorAll(".quality-radio");
const applyButton = safeGetElementById("filter-apply-button", true);

const LANGUAGE_STORAGE_KEY = "selectedLanguages";
const CATEGORY_STORAGE_KEY = "selectedCategories";

// Collapse a checkbox selection into a URL parameter. Empty string means "all"
// (no filter): triggered by the "0" (All) box, an empty selection, or every box
// being checked. Shared by the language and category filters.
const computeSelectionParam = (values, total) => {
  const isAll = values.includes("0") || values.length === 0 || values.length === total;
  return isAll ? "" : values.filter(val => val !== "0").join(",");
};

// The same selection is mirrored into a cookie so the server can filter the
// home grid on the first render instead of the page flashing every channel
// before a client-side redirect narrows it down.
// Values are checkbox ids joined by commas, so they need no escaping.
const readFilterCookie = (name) => {
  const match = document.cookie.split("; ").find(entry => entry.startsWith(name + "="));
  return match ? match.slice(name.length + 1) : "";
};

const writeFilterCookie = (name, value) => {
  document.cookie = value
    ? `${name}=${value}; path=/; max-age=31536000; samesite=lax`
    : `${name}=; path=/; max-age=0; samesite=lax`;
};

const updateQualityLabel = (value) => {
  const label = safeGetElementById("quality-picker-label", true);
  if (!label) return;
  if (!value || value === "auto") {
    label.textContent = "Quality";
    return;
  }
  const radio = document.querySelector(`.quality-radio[value="${value}"]`);
  const text = radio?.closest("label")?.querySelector(".label-text")?.textContent?.trim() || value;
  label.textContent = `Quality (${text})`;
};

const updateLanguageLabel = (count) => {
  const label = safeGetElementById("language-picker-label", true);
  if (label) label.textContent = count > 0 ? `Language (${count})` : "Language";
};

const updateCategoryLabel = (count) => {
  const label = safeGetElementById("category-picker-label", true);
  if (label) label.textContent = count > 0 ? `Categories (${count})` : "Categories";
};

// Save one checkbox group and return its URL param value, or null when the
// group is not on the page (a template can render either picker alone).
const persistFilter = (checkboxClass, storageKey, urlParam) => {
  const checkboxes = document.querySelectorAll(checkboxClass);
  if (!checkboxes.length) return null;

  const selected = [];
  checkboxes.forEach(cb => { if (cb.checked) selected.push(cb.value); });
  const param = computeSelectionParam(selected, checkboxes.length);

  if (param) setLocalStorageItem(storageKey, param.split(","));
  else removeLocalStorageItem(storageKey);
  writeFilterCookie(urlParam, param);
  return param;
};

// One Apply commits language and category together, so changing both costs a
// single navigation instead of two.
const applyFilters = () => {
  const applied = [
    ["language", persistFilter(".language-checkbox", LANGUAGE_STORAGE_KEY, "language")],
    ["category", persistFilter(".category-checkbox", CATEGORY_STORAGE_KEY, "category")],
  ];

  // Only the home grid reads these params, so filters applied from another
  // screen (e.g. the player) navigate home instead of reloading in place.
  const isHome = window.location.pathname === "/";
  const url = new URL(isHome ? window.location.href : "/", window.location.origin);
  applied.forEach(([urlParam, param]) => {
    if (param === null) return;
    if (param) url.searchParams.set(urlParam, param);
    else url.searchParams.delete(urlParam);
  });
  window.location.href = url.toString();
};

if (applyButton) applyButton.addEventListener("click", applyFilters);

// On page load restore both filters. A value present in the URL wins; otherwise
// the saved selection is shown, and the grid only reloads when the cookie the
// server filtered on disagrees with it.
(() => {
  const urlParams = getCurrentUrlParams();
  const redirectUrl = new URL(window.location.href);
  const isHome = window.location.pathname === "/";
  let needsRedirect = false;

  const restore = (checkboxClass, storageKey, urlParam, updateLabel) => {
    const checkboxes = document.querySelectorAll(checkboxClass);
    if (!checkboxes.length) return;
    const fromUrl = urlParams.get(urlParam);

    if (fromUrl) {
      const values = fromUrl.split(",");
      checkboxes.forEach(cb => {
        cb.checked = cb.value === "0" ? false : values.includes(cb.value);
      });
      updateLabel(values.length);
      return;
    }

    const saved = getLocalStorageItem(storageKey, []);
    if (Array.isArray(saved) && saved.length) {
      checkboxes.forEach(cb => { cb.checked = saved.includes(cb.value); });
      updateLabel(saved.length);

      // Off home the params do nothing and a reload would restart playback.
      // On home the server already filtered from the cookie, so only a stale
      // cookie (a selection saved before cookies were used, or a browser that
      // refuses them) still costs one redirect.
      if (!isHome) return;

      const value = saved.join(",");
      if (readFilterCookie(urlParam) === value) return;

      writeFilterCookie(urlParam, value);
      if (readFilterCookie(urlParam) !== value) redirectUrl.searchParams.set(urlParam, value);
      needsRedirect = true;
      return;
    }

    checkboxes.forEach(cb => { cb.checked = cb.value === "0"; });
    updateLabel(0);

    // Nothing saved, so a leftover cookie would filter the grid while the
    // drawer claims everything is selected.
    if (isHome && readFilterCookie(urlParam)) {
      writeFilterCookie(urlParam, "");
      needsRedirect = true;
    }
  };

  restore(".language-checkbox", LANGUAGE_STORAGE_KEY, "language", updateLanguageLabel);
  restore(".category-checkbox", CATEGORY_STORAGE_KEY, "category", updateCategoryLabel);

  if (needsRedirect) window.location.replace(redirectUrl.toString());
})();

// Mobile category drawer. On desktop the sidebar is a static column (the
// -translate-x-full is overridden by md:translate-x-0), so these only matter
// on small screens where the header hamburger drives them.
(() => {
  const sidebar = safeGetElementById("category-sidebar", true);
  const backdrop = safeGetElementById("sidebar-backdrop", true);
  if (!sidebar) return;

  const openSidebar = () => {
    sidebar.classList.remove("-translate-x-full");
    if (backdrop) backdrop.classList.remove("hidden");
    document.body.style.overflow = "hidden";
  };
  const closeSidebar = () => {
    sidebar.classList.add("-translate-x-full");
    if (backdrop) backdrop.classList.add("hidden");
    document.body.style.overflow = "";
  };

  window.openSidebar = openSidebar;
  window.closeSidebar = closeSidebar;
  window.toggleSidebar = () =>
    sidebar.classList.contains("-translate-x-full") ? openSidebar() : closeSidebar();

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") closeSidebar();
  });
})();

// Setup Select All toggle behavior
const setupSelectAll = (checkboxClass, allValue) => {
  const checkboxes = document.querySelectorAll(checkboxClass);
  const allCheckbox = Array.from(checkboxes).find(cb => cb.value === allValue);
  if (!allCheckbox) return;

  const otherCheckboxes = Array.from(checkboxes).filter(cb => cb.value !== allValue);

  allCheckbox.addEventListener("change", () => {
    if (allCheckbox.checked) {
      otherCheckboxes.forEach(cb => {
        cb.checked = false;
      });
    } else {
      const anyChecked = otherCheckboxes.some(item => item.checked);
      if (!anyChecked) {
        allCheckbox.checked = true;
      }
    }
  });

  otherCheckboxes.forEach(cb => {
    cb.addEventListener("change", () => {
      if (cb.checked) {
        allCheckbox.checked = false;
      } else {
        const anyChecked = otherCheckboxes.some(item => item.checked);
        if (!anyChecked) {
          allCheckbox.checked = true;
        }
      }
    });
  });
};

document.addEventListener('DOMContentLoaded', () => {
  // Run select-all wiring first so the language checkboxes behave correctly
  // even if favorites code below ever throws.
  setupSelectAll(".language-checkbox", "0");
  setupSelectAll(".category-checkbox", "0");
  updateFavoriteButtonStates();
  displayFavoriteChannels();
});

const onQualityChange = (elem) => {
  const quality = elem.value;

  if (quality === "auto") {
    updateUrlParameter("q", "");
    removeLocalStorageItem("quality");
  } else {
    updateUrlParameter("q", quality);
    setLocalStorageItem("quality", quality);
  }
  updateQualityLabel(quality);

  // Update all channel card href attributes with new query parameter.
  // Only target channel cards (a[href]); dropdown-content panels also carry
  // the .card class but are <div>s with no href, and touching them crashes.
  const currentParams = getCurrentUrlParams();
  document.querySelectorAll("a.card[data-channel-id]").forEach((cardElem) => {
    const href = cardElem.getAttribute("href");
    if (href) cardElem.setAttribute("href", href.split("?")[0] + "?" + currentParams.toString());
  });
};

const selectQualityRadio = (value) => {
  const radio = document.querySelector(`.quality-radio[value="${value}"]`);
  if (!radio) return false;
  radio.checked = true;
  return true;
};

qualityRadios.forEach((radio) => {
  radio.addEventListener("change", () => {
    if (radio.checked) onQualityChange(radio);
  });
});

const storedQuality = getLocalStorageItem("quality");
const urlQuality = getCurrentUrlParams().get("q");
if (urlQuality && selectQualityRadio(urlQuality)) {
  onQualityChange({ value: urlQuality });
} else if (storedQuality && selectQualityRadio(storedQuality)) {
  updateQualityLabel(storedQuality);
} else {
  updateQualityLabel("auto");
}


const scrollToTop = () => {
  window.scrollTo({
    top: 0,
    behavior: "smooth",
  });
};

// Favorite Channels Functionality
const FAVORITES_STORAGE_KEY = "favoriteChannels";

function getFavoriteChannels() {
  return getLocalStorageItem(FAVORITES_STORAGE_KEY, []);
}

function saveFavoriteChannels(favoriteIds) {
  setLocalStorageItem(FAVORITES_STORAGE_KEY, favoriteIds);
}

function displayFavoriteChannels() {
  const favoriteIds = getFavoriteChannels();
  const elements = safeGetElementsById([
    "favorite-channels-section",
    "favorite-channels-container", 
    "original-channels-grid"
  ]);
  
  const { 
    "favorite-channels-section": favoriteChannelsSection,
    "favorite-channels-container": favoriteChannelsContainer,
    "original-channels-grid": originalChannelsGrid 
  } = elements;

  if (!favoriteChannelsSection || !favoriteChannelsContainer || !originalChannelsGrid) {
    // Pages without the home favourites section (e.g. the play page) still call
    // this via toggleFavorite; there is simply no grid to reorganise there.
    return;
  }

  // Move all cards to a temporary fragment to prevent issues with live collections
  // or ensure they are detached before re-appending.
  // However, a simpler approach for now is to just re-append.
  // This might cause a brief flicker for a large number of cards.
  
  // Clear favorite container before potentially hiding it or re-populating
  // while (favoriteChannelsContainer.firstChild) {
  //   favoriteChannelsContainer.removeChild(favoriteChannelsContainer.firstChild);
  // }
  // The logic below of appending will move them, so explicit clearing is not strictly necessary
  // if we iterate over ALL cards and move them to correct container.

  if (favoriteIds.length > 0) {
    favoriteChannelsSection.style.display = 'block'; // Or 'flex' or 'grid' depending on layout
  } else {
    favoriteChannelsSection.style.display = 'none';
  }

  const allChannelCards = document.querySelectorAll('a.card[data-channel-id]');

  // Create DocumentFragments to batch DOM updates
  const favoriteFragment = document.createDocumentFragment();
  const originalFragment = document.createDocumentFragment();

  allChannelCards.forEach(card => {
    const cardChannelId = card.dataset.channelId;
    if (favoriteIds.includes(cardChannelId)) {
      favoriteFragment.appendChild(card);
    } else {
      originalFragment.appendChild(card);
    }
  });

  // Append fragments to their respective containers
  favoriteChannelsContainer.appendChild(favoriteFragment);
  originalChannelsGrid.appendChild(originalFragment);
}

function toggleFavorite(channelId) {
  const favoriteIds = getFavoriteChannels();
  const index = favoriteIds.indexOf(channelId);

  if (index > -1) { // Channel was a favorite, removing it
    favoriteIds.splice(index, 1);
    updateFavoriteButtonState(channelId, false);
  } else { // Channel was not a favorite, adding it
    favoriteIds.push(channelId);
    updateFavoriteButtonState(channelId, true);
  }
  
  saveFavoriteChannels(favoriteIds);
  displayFavoriteChannels(); // Refresh the channel lists
}

function updateFavoriteButtonStates() {
  const favoriteIds = getFavoriteChannels();
  const favoriteButtons = document.querySelectorAll(".favorite-btn");

  favoriteButtons.forEach(button => {
    const channelId = button.id.replace("favorite-btn-", "");
    updateFavoriteButtonState(channelId, favoriteIds.includes(channelId));
  });
}
