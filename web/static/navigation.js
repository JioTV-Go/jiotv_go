let focusableChannels = [];
let currentFocusIndex = -1;

function updateFocusableChannels() {
  const allCards = document.querySelectorAll('.card[data-channel-id]');
  focusableChannels = Array.from(allCards).filter(card => card.style.display !== 'none');
  
  // Reset focus if the list is empty or significantly changed.
  // For now, just reset if empty, or try to set to 0 if not.
  if (focusableChannels.length > 0) {
    // If currentFocusIndex is out of bounds or was -1, try to set to 0
    if (currentFocusIndex < 0 || currentFocusIndex >= focusableChannels.length) {
       // setNewFocus will handle if 0 is a valid index
    }
    // If currentFocusIndex is still valid, let it be. setNewFocus will be called later if needed.
    // However, the instruction was: currentFocusIndex = (focusableChannels.length > 0) ? 0 : -1;
    // This would mean always resetting to the first element after an update.
    // Let's defer the decision of resetting currentFocusIndex until after setNewFocus is called.
    // For now, a simple reset to 0 if list is not empty, else -1
    currentFocusIndex = (focusableChannels.length > 0) ? 0 : -1;
  } else {
    currentFocusIndex = -1;
  }
}

function removeOldFocus() {
  if (currentFocusIndex >= 0 && currentFocusIndex < focusableChannels.length && focusableChannels[currentFocusIndex]) {
    focusableChannels[currentFocusIndex].classList.remove('channel-focused');
  }
}

function setNewFocus(newIndex) {
  if (newIndex >= 0 && newIndex < focusableChannels.length) {
    removeOldFocus(); // Remove focus from the old element
    currentFocusIndex = newIndex;
    const newFocusedElement = focusableChannels[currentFocusIndex];
    newFocusedElement.classList.add('channel-focused');
    newFocusedElement.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  } else if (focusableChannels.length === 0) {
    removeOldFocus(); // Ensure no focus if list is empty
    currentFocusIndex = -1;
  }
}

function handleKeyDown(event) {
  // Ignore key events if the target is an input, select, or textarea
  if (event.target.tagName === 'INPUT' || event.target.tagName === 'SELECT' || event.target.tagName === 'TEXTAREA') {
    return;
  }

  // If there are no focusable channels, or if no channel is currently focused (e.g. initial state before any interaction)
  // and the pressed key is not one that should initiate focus (e.g. arrow keys), then do nothing.
  // For this implementation, we'll allow arrow keys to initiate focus if currentFocusIndex is -1 and channels exist.
  if (focusableChannels.length === 0) {
    return;
  }

  let newIndex = currentFocusIndex;

  switch (event.key) {
    case 'ArrowUp':
    case 'ArrowLeft':
      event.preventDefault();
      if (currentFocusIndex === -1 && focusableChannels.length > 0) { // Initial focus
        newIndex = 0;
      } else {
        newIndex = (currentFocusIndex - 1 + focusableChannels.length) % focusableChannels.length;
      }
      setNewFocus(newIndex);
      break;
    case 'ArrowDown':
    case 'ArrowRight':
      event.preventDefault();
      if (currentFocusIndex === -1 && focusableChannels.length > 0) { // Initial focus
        newIndex = 0;
      } else {
        newIndex = (currentFocusIndex + 1) % focusableChannels.length;
      }
      setNewFocus(newIndex);
      break;
    case 'Enter':
      event.preventDefault();
      if (currentFocusIndex >= 0 && currentFocusIndex < focusableChannels.length && focusableChannels[currentFocusIndex]) {
        const targetElement = focusableChannels[currentFocusIndex];
        if (targetElement && targetElement.href) {
          window.location.href = targetElement.href;
        }
      }
      break;
    default:
      return; // Exit if the key is not handled
  }
}

document.addEventListener('DOMContentLoaded', () => {
  updateFocusableChannels();
  if (focusableChannels.length > 0) {
    // If currentFocusIndex was set to 0 by updateFocusableChannels, setNewFocus will apply it.
    // Or, if we want to explicitly set focus to the first item always on load:
    setNewFocus(0); 
  }
  document.addEventListener('keydown', handleKeyDown);
});
