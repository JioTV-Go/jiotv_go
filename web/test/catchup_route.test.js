// Regression guard: the catchup mode toggle used to treat the sub-route in
// /catchup/play/<id> as the channel id and redirect to /play/play, which loaded
// a player for a channel named "play".
const { readFileSync } = require("node:fs");

beforeEach(() => {
  // utils.js supplies getLocalStorageItem/getCurrentUrlParams used by common.js.
  window.eval(readFileSync("static/internal/utils.js", "utf8"));
  window.eval(readFileSync("static/internal/common.js", "utf8"));
});

describe("parseChannelRoute", () => {
  test("parses top-level channel routes", () => {
    expect(window.parseChannelRoute("/play/143")).toEqual({
      mode: "play",
      id: "143",
    });
    expect(window.parseChannelRoute("/catchup/143")).toEqual({
      mode: "catchup",
      id: "143",
    });
    expect(window.parseChannelRoute("/catchup/143/")).toEqual({
      mode: "catchup",
      id: "143",
    });
  });

  test("ignores nested catchup sub-routes", () => {
    expect(window.parseChannelRoute("/catchup/play/143")).toBeNull();
    expect(window.parseChannelRoute("/catchup/render/143")).toBeNull();
    expect(window.parseChannelRoute("/catchup/stream/143")).toBeNull();
  });

  test("ignores a bare sub-route name in the id position", () => {
    expect(window.parseChannelRoute("/catchup/play")).toBeNull();
    expect(window.parseChannelRoute("/play/render")).toBeNull();
  });

  test("ignores unrelated paths", () => {
    expect(window.parseChannelRoute("/")).toBeNull();
    expect(window.parseChannelRoute("/player/143")).toBeNull();
    expect(window.parseChannelRoute("/premium/providers")).toBeNull();
  });
});
