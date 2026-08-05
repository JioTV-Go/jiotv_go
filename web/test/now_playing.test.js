const { readFileSync } = require("node:fs");

const loadNowPlayingScript = () => {
  window.eval(readFileSync("static/internal/now_playing.js", "utf8"));
};

const setDimensions = (element, clientHeight, scrollHeight) => {
  Object.defineProperties(element, {
    clientHeight: { configurable: true, value: clientHeight },
    scrollHeight: { configurable: true, value: scrollHeight },
  });
};

describe("now-playing description", () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <p id="description" class="line-clamp-3"></p>
      <button id="description-toggle" type="button" class="hidden" aria-expanded="false"></button>
      <div id="keywords"></div>
    `;
  });

  test("shows Read more only when description exceeds three lines", () => {
    const description = document.getElementById("description");
    setDimensions(description, 60, 100);
    loadNowPlayingScript();

    updateNowPlayingDescription();

    const button = document.getElementById("description-toggle");
    expect(button.hidden).toBe(false);
    expect(button.classList.contains("hidden")).toBe(false);
    expect(button.textContent).toBe("Read more");
    expect(description.classList.contains("line-clamp-3")).toBe(true);
  });

  test("expands and collapses a long description", () => {
    const description = document.getElementById("description");
    setDimensions(description, 60, 100);
    loadNowPlayingScript();
    updateNowPlayingDescription();

    const button = document.getElementById("description-toggle");
    button.click();
    expect(description.classList.contains("line-clamp-3")).toBe(false);
    expect(button.textContent).toBe("Show less");
    expect(button.getAttribute("aria-expanded")).toBe("true");

    button.click();
    expect(description.classList.contains("line-clamp-3")).toBe(true);
    expect(button.textContent).toBe("Read more");
    expect(button.getAttribute("aria-expanded")).toBe("false");
  });

  test("keeps short descriptions expanded and hides the control", () => {
    const description = document.getElementById("description");
    setDimensions(description, 60, 60);
    loadNowPlayingScript();

    updateNowPlayingDescription();

    const button = document.getElementById("description-toggle");
    expect(button.hidden).toBe(true);
    expect(description.classList.contains("line-clamp-3")).toBe(false);
  });
});
