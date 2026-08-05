// now_playing.js — description disclosure for live and catchup cards.
function updateNowPlayingDescription() {
  const description = document.getElementById("description");
  const button = document.getElementById("description-toggle");
  if (!description || !button) return;

  description.classList.add("line-clamp-3");
  const isLong = description.scrollHeight > description.clientHeight;
  button.hidden = !isLong;
  button.classList.toggle("hidden", !isLong);
  button.setAttribute("aria-expanded", "false");
  button.textContent = "Read more";
  if (!isLong) description.classList.remove("line-clamp-3");
}
if (!window.__nowPlayingListenerRegistered) {
  document.addEventListener("click", (event) => {
    const button = event.target.closest("#description-toggle");
    if (!button) return;
    const description = document.getElementById("description");
    if (!description) return;

    const expanded = button.getAttribute("aria-expanded") === "true";
    description.classList.toggle("line-clamp-3", expanded);
    button.setAttribute("aria-expanded", String(!expanded));
    button.textContent = expanded ? "Read more" : "Show less";
  });
  window.__nowPlayingListenerRegistered = true;
}

document.addEventListener("DOMContentLoaded", updateNowPlayingDescription);
