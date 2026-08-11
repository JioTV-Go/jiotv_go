# Web Package Guidelines

These instructions apply to `web/` and supplement the repository-root `AGENTS.md`.

## Package Shape

- `views/` contains Go `html/template` files rendered by the Fiber server.
- `static/internal/` contains project-owned vanilla JavaScript, Tailwind source, and generated CSS.
- `static/external/` contains vendored browser libraries; do not edit them as project source.
- `test/` contains Jest tests running under jsdom.
- `web.go` embeds `views/*` and `static/*` into the Go binary. Changes must work through the server, not only when files are opened directly.

Do not introduce a frontend framework, bundler, TypeScript, CSS preprocessor, or new dependency unless the task requires it. Reuse existing browser APIs and project scripts.

## Commands

Run these from `web/`:

```bash
npm ci
npm run build
npm test -- --watchAll=false --ci
npm run test:coverage
```

For a focused test:

```bash
npm test -- --watchAll=false --ci test/<name>.test.js
```

Use `npm run watch` only during interactive CSS work; it is a long-running process.

## Templates and JavaScript

- Keep shared markup in existing templates such as `navbar.html`, `footer.html`, `styling.html`, and `channel_list.html`.
- Preserve Go template actions and server-provided field names. Trace the corresponding Fiber render map before renaming or removing one.
- JavaScript is loaded as classic scripts, not ES modules. Load order matters: templates commonly load `utils.js` before scripts that consume its globals.
- Keep reusable behavior in `static/internal/*.js`; avoid adding substantial inline scripts to templates.
- Use DOM APIs and `textContent` for untrusted API or user-derived values. Do not interpolate such values into `innerHTML`.
- Preserve route contracts used by the server and players, including `/play/`, `/catchup/`, `/static/`, and render/proxy URLs.
- Keep controls keyboard-usable and labeled. Player, remote-navigation, dialog, and toggle changes must preserve focus behavior and ARIA state.

## Styling

- Tailwind CSS 4 and DaisyUI 5 are configured in `static/internal/input.css`; there is no Tailwind or PostCSS configuration file.
- Prefer existing DaisyUI components and Tailwind utilities. Add custom CSS only for behavior or styling utilities cannot express cleanly.
- `static/internal/tailwind.css` is generated, minified, committed output. Never edit it manually.
- Run `npm run build` after changing `input.css`, template class names, dynamically generated class names, or frontend dependencies, and include the regenerated `tailwind.css` when it changes.
- Verify responsive layout, overflow, focus visibility, and text fit on mobile and desktop. jsdom cannot prove layout or media playback.

## Tests

- Add behavior-focused tests under `test/*.test.js` for changed JavaScript contracts.
- Tests run in jsdom via `jest.config.js`. Set up the smallest required DOM in `beforeEach` and reset mutated globals, mocks, storage, document head/body, and history between cases.
- Prefer loading the real source with `readFileSync` plus `window.eval`, as existing regression tests do. Do not copy production functions into a test: duplicated implementations can pass while source is broken.
- Mock only browser or network boundaries absent from jsdom, such as `fetch`, `matchMedia`, media APIs, element geometry, and player libraries.
- Assert observable DOM, URL, storage, accessibility-state, or request behavior. Avoid probabilistic assertions and tests of third-party framework behavior.
- A bug fix needs one regression case that fails for the reported behavior before the source fix.

## Verification by Change Type

- JavaScript logic: run the focused Jest test, then the full Jest suite.
- Template or styling: run `npm run build`, the relevant Jest tests, and inspect the served page in a real browser at mobile and desktop sizes.
- Player behavior: use a real browser and inspect console and network through the first media request; a rendered player shell is not proof of playback.
- Dependency changes: use `npm install` to update `package-lock.json`, then run `npm ci`, `npm run build`, and the full Jest suite.
