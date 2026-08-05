// remote.js — keyboard & TV-remote navigation.
// TV remotes send arrow keys, Enter and Back. Interactive elements are
// already natively focusable (links, buttons, inputs, cards with
// tabindex=0) and Enter activates them, so this file adds the missing
// pieces: spatial arrow-key focus movement, an entry point when nothing
// is focused, and a Back action on player pages.
(function () {
  const FOCUSABLE =
    'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), iframe, [tabindex]:not([tabindex="-1"])';
  // Controls that need arrows for their native value/caret behaviour.
  // Text inputs use arrows for spatial remote navigation. Enter still opens
  // native select menus; only controls with essential arrow value/caret use
  // retain native arrows.
  function keepsArrowKeys(el) {
    if (!el) return false;
    if (el.tagName === "TEXTAREA" || el.tagName === "SELECT") return true;
    return el.tagName === "INPUT" && ["number", "range"].includes(el.type);
  }
  const KEY_DIRS = {
    ArrowUp: "up",
    ArrowDown: "down",
    ArrowLeft: "left",
    ArrowRight: "right",
  };

  function isVisible(el) {
    if (typeof el.checkVisibility === "function") {
      return el.checkVisibility();
    }
    const r = el.getBoundingClientRect();
    return r.width > 0 || r.height > 0;
  }

  // While a dialog is open, navigation stays trapped inside it.
  function getFocusables() {
    const scope = document.querySelector("dialog[open]") || document;
    return Array.from(scope.querySelectorAll(FOCUSABLE)).filter(isVisible);
  }
  function center(el) {
    const r = el.getBoundingClientRect();
    return { x: r.left + r.width / 2, y: r.top + r.height / 2 };
  }

  // Nearest candidate in the pressed direction: it must lie along that
  // axis, and off-axis drift is penalised 3x so a same-row/column neighbour wins.
  // Controls nested inside the focused element (e.g. a card's favorite
  // button) are skipped — they stay reachable via Tab.
  function pickNext(current, dir, candidates) {
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
      // Stay inside a 45° cone. This keeps toolbar D-pad navigation on
      // its row instead of diving into a card action diagonally below.
      if (primary <= 0 || off > primary) continue;
      const score = primary + off * 3;
      if (score < bestScore) {
        bestScore = score;
        best = el;
      }
    }
    return best;
  }

  function focusElement(el) {
    el.focus({ preventScroll: true });
    el.scrollIntoView({ block: "nearest", inline: "nearest", behavior: "smooth" });
  }

  document.addEventListener("keydown", (e) => {
    if (e.ctrlKey || e.altKey || e.metaKey || e.shiftKey) return;
    const isBackKey =
      e.key === "Escape" ||
      e.key === "Backspace" ||
      e.key === "BrowserBack" ||
      e.key === "GoBack" ||
      e.key === "Back";
    if (isBackKey) {
      // Dialogs handle Escape natively; don't fight them.
      if (document.querySelector("dialog[open]")) return;
      const active = document.activeElement;
      if (keepsArrowKeys(active)) return;
      if (e.key === "Backspace" && active !== document.body) return;
      // Player pages: Back returns to the previous page, like a TV remote.
      if (document.getElementById("playerIframe")) {
        e.preventDefault();
        window.history.back();
      }
      return;
    }

    // Some TV remotes identify the OK key as Select instead of Enter.
    if (e.key === "Select" && document.activeElement instanceof HTMLElement) {
      e.preventDefault();
      document.activeElement.click();
      return;
    }

    const dir = KEY_DIRS[e.key];
    if (!dir) return;

    const active = document.activeElement;
    if (keepsArrowKeys(active)) return;

    const candidates = getFocusables();
    if (candidates.length === 0) return;

    let next = null;
    if (active && active !== document.body && candidates.includes(active)) {
      next = pickNext(active, dir, candidates);
    } else {
      // Nothing focused yet: enter at the first element (last for up/left).
      next =
        dir === "up" || dir === "left"
          ? candidates[candidates.length - 1]
          : candidates[0];
    }
    if (next) {
      e.preventDefault();
      focusElement(next);
    }
  });
})();
