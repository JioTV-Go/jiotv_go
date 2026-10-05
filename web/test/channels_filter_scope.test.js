/**
 * @jest-environment jsdom
 */
const { readFileSync } = require("node:fs");

// The filter drawer now also renders on /play, where the language/category
// params are meaningless. These cases pin where Apply navigates and that the
// player is never reloaded just to restore a saved selection.
const utilsSource = readFileSync("static/internal/utils.js", "utf8");
const channelsSource = readFileSync("static/internal/channels.js", "utf8");

// jsdom neither navigates nor lets us redefine window.location, so the real
// sources run with window shadowed by a proxy that serves a stub location.
const loadChannelsScript = (url) => {
  const parsed = new URL(url);
  const location = {
    href: url,
    origin: parsed.origin,
    pathname: parsed.pathname,
    search: parsed.search,
    replace: jest.fn(),
  };

  window.__testWindow = new Proxy(window, {
    get: (target, prop) => {
      if (prop === "location") return location;
      const value = target[prop];
      return typeof value === "function" ? value.bind(target) : value;
    },
    set: (target, prop, value) => {
      target[prop] = value;
      return true;
    },
  });

  window.eval(
    `(function (window) {\n${utilsSource}\n${channelsSource}\n})(window.__testWindow);`
  );

  return location;
};

const renderDrawer = () => {
  document.body.innerHTML = `
    <aside id="category-sidebar" class="-translate-x-full"></aside>
    <div id="sidebar-backdrop" class="hidden"></div>
    <div class="section-title" id="quality-picker-label">Quality</div>
    <label><input type="radio" name="quality" value="auto" class="quality-radio" checked /><span class="label-text">Quality (Auto)</span></label>
    <label><input type="radio" name="quality" value="high" class="quality-radio" /><span class="label-text">High</span></label>
    <h2 class="section-title" id="language-picker-label">Language</h2>
    <input type="checkbox" name="language" value="0" class="language-checkbox" checked />
    <input type="checkbox" name="language" value="6" class="language-checkbox" />
    <h2 class="section-title" id="category-picker-label">Categories</h2>
    <input type="checkbox" name="category" value="0" class="category-checkbox" checked />
    <input type="checkbox" name="category" value="5" class="category-checkbox" />
    <input type="checkbox" name="category" value="8" class="category-checkbox" />
    <button id="filter-apply-button">Apply</button>
  `;
};

const clearCookies = () => {
  document.cookie.split("; ").forEach(entry => {
    const name = entry.split("=")[0];
    if (name) document.cookie = `${name}=; path=/; max-age=0`;
  });
};

beforeEach(() => {
  localStorage.clear();
  clearCookies();
  renderDrawer();
});

test("quality is a single-select radio and updates the picker label", () => {
  loadChannelsScript("http://localhost:5001/");

  const high = document.querySelector('.quality-radio[value="high"]');
  high.checked = true;
  high.dispatchEvent(new Event("change"));

  expect(document.querySelector('.quality-radio[value="auto"]').checked).toBe(false);
  expect(localStorage.getItem("quality")).toBe('"high"');
  expect(document.getElementById("quality-picker-label").textContent).toBe("Quality (High)");
});

test("Apply on the play page sends the filter to the home grid", () => {
  const location = loadChannelsScript("http://localhost:5001/play/143?q=high");

  document.querySelector('.category-checkbox[value="0"]').checked = false;
  document.querySelector('.category-checkbox[value="5"]').checked = true;
  document.getElementById("filter-apply-button").click();

  expect(location.href).toBe("http://localhost:5001/?category=5");
});

test("Apply on the home page keeps filtering in place", () => {
  const location = loadChannelsScript("http://localhost:5001/?language=6");

  document.querySelector('.category-checkbox[value="0"]').checked = false;
  document.querySelector('.category-checkbox[value="8"]').checked = true;
  document.getElementById("filter-apply-button").click();

  expect(location.href).toBe("http://localhost:5001/?language=6&category=8");
});

test("one Apply commits language and category in a single navigation", () => {
  const location = loadChannelsScript("http://localhost:5001/");

  document.querySelector('.language-checkbox[value="0"]').checked = false;
  document.querySelector('.language-checkbox[value="6"]').checked = true;
  document.querySelector('.category-checkbox[value="0"]').checked = false;
  document.querySelector('.category-checkbox[value="5"]').checked = true;
  document.getElementById("filter-apply-button").click();

  expect(location.href).toBe("http://localhost:5001/?language=6&category=5");
  expect(localStorage.getItem("selectedLanguages")).toBe('["6"]');
  expect(localStorage.getItem("selectedCategories")).toBe('["5"]');
});

test("Apply stores the selection in a cookie for the server to filter on", () => {
  loadChannelsScript("http://localhost:5001/play/143");

  document.querySelector('.category-checkbox[value="0"]').checked = false;
  document.querySelector('.category-checkbox[value="5"]').checked = true;
  document.querySelector('.category-checkbox[value="8"]').checked = true;
  document.getElementById("filter-apply-button").click();

  expect(document.cookie).toContain("category=5,8");
});

test("Apply clears the cookie when the filter goes back to All", () => {
  document.cookie = "category=5; path=/";
  loadChannelsScript("http://localhost:5001/");

  document.querySelector('.category-checkbox[value="0"]').checked = true;
  document.getElementById("filter-apply-button").click();

  expect(document.cookie).not.toContain("category=5");
});

test("a saved filter matching the cookie renders without a redirect", () => {
  localStorage.setItem("selectedCategories", JSON.stringify(["5"]));
  document.cookie = "category=5; path=/";

  const location = loadChannelsScript("http://localhost:5001/");

  expect(location.replace).not.toHaveBeenCalled();
  expect(document.querySelector('.category-checkbox[value="5"]').checked).toBe(true);
  expect(document.getElementById("category-picker-label").textContent).toBe("Categories (1)");
});

test("a saved filter without a cookie writes it and reloads once", () => {
  localStorage.setItem("selectedCategories", JSON.stringify(["5"]));

  const location = loadChannelsScript("http://localhost:5001/");

  expect(document.cookie).toContain("category=5");
  expect(location.replace).toHaveBeenCalledWith("http://localhost:5001/");
});

test("a cookie left over from a cleared selection is dropped", () => {
  document.cookie = "language=6; path=/";

  const location = loadChannelsScript("http://localhost:5001/");

  expect(document.cookie).not.toContain("language=6");
  expect(location.replace).toHaveBeenCalledWith("http://localhost:5001/");
});

test("a saved filter is shown, not reloaded, on the play page", () => {
  localStorage.setItem("selectedCategories", JSON.stringify(["5"]));

  const location = loadChannelsScript("http://localhost:5001/play/143");

  expect(location.replace).not.toHaveBeenCalled();
  expect(document.querySelector('.category-checkbox[value="5"]').checked).toBe(true);
  expect(document.querySelector('.category-checkbox[value="0"]').checked).toBe(false);
  expect(document.getElementById("category-picker-label").textContent).toBe("Categories (1)");
});

test("Browse can toggle the drawer on the play page", () => {
  loadChannelsScript("http://localhost:5001/play/143");

  const sidebar = document.getElementById("category-sidebar");
  window.toggleSidebar();
  expect(sidebar.classList.contains("-translate-x-full")).toBe(false);
  window.toggleSidebar();
  expect(sidebar.classList.contains("-translate-x-full")).toBe(true);
});
