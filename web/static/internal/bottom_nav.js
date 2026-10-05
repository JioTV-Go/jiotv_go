// bottom_nav.js — mobile bottom tab bar. Search scrolls up and focuses the
// header search field; Browse reuses window.toggleSidebar from channels.js.
// The bar is hidden on desktop (CSS md:hidden), so this is a no-op there.
(() => {
  const searchInput = document.getElementById("portexe-search-input");

  window.focusSearch = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
    if (searchInput) requestAnimationFrame(() => searchInput.focus());
  };

  // Highlight Home only on the home page; other screens leave all tabs neutral.
  if (window.location.pathname === "/") {
    document
      .querySelector('.bottom-nav a[href="/"]')
      ?.classList.add("is-active");
  }
})();
