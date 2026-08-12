---
name: diagnose-playback
description: Diagnoses JioTV Go live, catch-up, HLS, DASH/Widevine DRM, browser, and IPTV playback failures end to end. Use when playback fails, stalls, returns upstream 400/401/403, reports decoder or demuxer errors, or differs by channel, programme, browser, or IPTV client.
---

# Diagnose Playback

Find the first failing hop. Player errors are downstream symptoms, not root causes.

## Establish the case

Record channel ID, live/catch-up/premium mode, programme `srno` and timestamps when applicable, requested quality, client, and whether DRM is enabled. Do not compare cases until these inputs match.

Run with `JIOTV_DEBUG=true JIOTV_LOG_TO_STDOUT=true`. Reproduce one failing request and retain status, redirect location, content type, and relevant server log lines for each hop. Never publish credentials, encrypted `auth` values, SSO tokens, license challenges, or HDNEA values.

## Trace the chain

- HLS live: `/live/...` -> `/render.m3u8` master -> rewritten `/render.m3u8` variant -> `/render.ts`, `/render.aac`, or `/render.key`.
- HLS catch-up: `/catchup/stream/...` -> same render chain, with programme time query parameters preserved.
- Web DRM: `/play/...` -> `/render.mpd` -> `/render.dash/...`; license requests go through `/drm`.
- IPTV DRM: `/live/mpd/:channelID` plus `/live/key/:channelID`.

For browser-only failures, inspect Network and Console in a real browser. For IPTV-only failures, inspect generated `#KODIPROP` manifest and license URLs. A successful page or master manifest is insufficient: continue through one variant and one media segment, or through MPD, media segment, and license response for DRM.

Classify the first failure before editing:

- Local route/input failure: local 4xx/5xx before an upstream request.
- Upstream entitlement/content failure: JioTV rejects a specific account, channel, or programme consistently.
- Token scope/expiry failure: manifest succeeds but variant, segment, key, or license returns 400/401/403.
- Rewrite/proxy failure: local manifest drops query data, builds the wrong route, changes required headers, or returns the wrong content type.
- Client capability failure: fetched media is valid but codec, container, secure-context, or Widevine support is absent.

## Project gotchas

- Live and catch-up HDNEA tokens have different ACLs. Cache and compare them by stream kind; never reuse a live token for catch-up.
- Preserve complete media URI query strings, especially catch-up `vbegin` and `vend`. Extension detection must use the URI path without discarding its query.
- Upstream tokens can arrive in `hdnea`, `__hdnea__`, URL query data, cookies, or API fields. Avoid sending duplicate conflicting values.
- Browser-added `Origin`, `Referer`, and fetch headers can make key or license upstreams reject otherwise valid requests. Use existing player-header helpers.
- Widevine needs a supported browser and a secure context (`localhost` or HTTPS). Separate unsupported-client failures from server failures.
- One channel or programme failing while adjacent cases work can be upstream content state. Prove this with a small comparison set before changing shared proxy code.

## Fix and verify

Fix the earliest shared boundary that violates the chain. Preserve encrypted upstream URLs and existing credential-refresh flow. Do not add retries unless evidence shows a transient failure and retry semantics are bounded.

Add one focused regression test using `httptest` or Fiber's test client for the failed contract. Then rerun the exact failing chain through the first previously failing media request, plus the focused package or frontend test. For browser behavior, verify in a real browser; for IPTV behavior, exercise the playlist/manifest/license route used by the client.
