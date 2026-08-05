// Tests for the spatial-navigation pick algorithm used by remote.js.
// jsdom has no layout engine, so element rects are mocked.
const { readFileSync } = require("node:fs");

const makeEl = (id, x, y, w = 100, h = 50) => ({
  id,
  getBoundingClientRect: () => ({
    left: x,
    top: y,
    width: w,
    height: h,
    right: x + w,
    bottom: y + h,
  }),
});

const center = (el) => {
  const r = el.getBoundingClientRect();
  return { x: r.left + r.width / 2, y: r.top + r.height / 2 };
};

// Mirrors pickNext in static/internal/remote.js
const pickNext = (current, dir, candidates) => {
  const c = center(current);
  const cur = current.getBoundingClientRect();
  let best = null;
  let bestScore = Infinity;
  for (const el of candidates) {
    if (el === current) continue;
    const b = center(el);
    if (
      b.x >= cur.left &&
      b.x <= cur.right &&
      b.y >= cur.top &&
      b.y <= cur.bottom
    ) {
      continue;
    }
    const dx = b.x - c.x;
    const dy = b.y - c.y;
    let primary;
    let off;
    if (dir === "right") {
      primary = dx;
      off = Math.abs(dy);
    } else if (dir === "left") {
      primary = -dx;
      off = Math.abs(dy);
    } else if (dir === "down") {
      primary = dy;
      off = Math.abs(dx);
    } else {
      primary = -dy;
      off = Math.abs(dx);
    }
    if (primary <= 0 || off > primary) continue;
    const score = primary + off * 3;
    if (score < bestScore) {
      bestScore = score;
      best = el;
    }
  }
  return best;
};

describe("remote navigation pickNext", () => {
  // 3-column grid layout:
  //   a1 (0,0)    a2 (120,0)   a3 (240,0)
  //   b1 (0,60)   b2 (120,60)  b3 (240,60)
  const a1 = makeEl("a1", 0, 0);
  const a2 = makeEl("a2", 120, 0);
  const a3 = makeEl("a3", 240, 0);
  const b1 = makeEl("b1", 0, 60);
  const b2 = makeEl("b2", 120, 60);
  const b3 = makeEl("b3", 240, 60);
  const grid = [a1, a2, a3, b1, b2, b3];

  test("right moves to the nearest element on the right", () => {
    expect(pickNext(a1, "right", grid)).toBe(a2);
    expect(pickNext(b2, "right", grid)).toBe(b3);
  });

  test("left moves to the nearest element on the left", () => {
    expect(pickNext(a3, "left", grid)).toBe(a2);
    expect(pickNext(b2, "left", grid)).toBe(b1);
  });

  test("down moves to the element below", () => {
    expect(pickNext(a2, "down", grid)).toBe(b2);
  });

  test("up moves to the element above", () => {
    expect(pickNext(b2, "up", grid)).toBe(a2);
  });

  test("returns null at the edge of the grid", () => {
    expect(pickNext(a3, "right", grid)).toBeNull();
    expect(pickNext(a1, "left", grid)).toBeNull();
    expect(pickNext(a1, "up", grid)).toBeNull();
    expect(pickNext(b3, "down", grid)).toBeNull();
  });

  test("prefers straight-ahead over diagonal candidates", () => {
    const current = makeEl("c", 0, 0);
    const straight = makeEl("straight", 200, 0);
    const diagonal = makeEl("diagonal", 110, 100);
    expect(pickNext(current, "right", [current, straight, diagonal])).toBe(straight);
  });

  test("skips candidates behind the current element", () => {
    const current = makeEl("c", 100, 0);
    const behind = makeEl("behind", 0, 0);
    expect(pickNext(current, "right", [current, behind])).toBeNull();
  });

  test("skips controls nested inside the focused element", () => {
    const card = makeEl("card", 0, 0, 200, 100);
    const star = makeEl("star", 150, 10, 20, 20);
    const neighbour = makeEl("neighbour", 240, 0, 200, 100);
    expect(pickNext(card, "right", [card, star, neighbour])).toBe(neighbour);
  });

  test("prefers an aligned toolbar control over a closer card action", () => {
    const search = makeEl("search", 93, 77, 244, 40);
    const quality = makeEl("quality", 345, 77, 244, 40);
    const favorite = makeEl("favorite", 242, 152, 40, 40);
    expect(pickNext(search, "right", [search, quality, favorite])).toBe(quality);
  });

  test("keeps toolbar left navigation on the same row", () => {
    const language = makeEl("language", 850, 77, 244, 40);
    const category = makeEl("category", 598, 77, 244, 40);
    const favorite = makeEl("favorite", 873, 152, 40, 40);
    expect(pickNext(language, "left", [language, category, favorite])).toBe(category);
  });
});


describe("remote navigation key handling", () => {
  beforeEach(() => {
    jest.resetModules();
    document.body.innerHTML = '<input id="search"><button id="apply">Apply</button>';
    for (const [id, rect] of Object.entries({
      search: { left: 0, top: 0, width: 200, height: 40 },
      apply: { left: 220, top: 0, width: 100, height: 40 },
    })) {
      const el = document.getElementById(id);
      el.getBoundingClientRect = () => ({
        ...rect,
        right: rect.left + rect.width,
        bottom: rect.top + rect.height,
      });
      el.scrollIntoView = jest.fn();
    }
    window.eval(readFileSync("static/internal/remote.js", "utf8"));
  });

  test("moves right out of a focused text search field", () => {
    const search = document.getElementById("search");
    const apply = document.getElementById("apply");
    search.focus();
    search.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }));
    expect(document.activeElement).toBe(apply);
  });
});
