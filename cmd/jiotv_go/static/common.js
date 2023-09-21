const getCurrentTheme = () => {
  const htmlTag = document.getElementsByTagName("html")[0];
  if (htmlTag.getAttribute("data-theme")) {
    return htmlTag.getAttribute("data-theme");
  } else {
    // return system theme
    if (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    ) {
      return "dark";
    }
    return "light";
  }
};

const toggleTheme = () => {
  // toggle or add attribute "data-theme" to html tag
  const htmlTag = document.getElementsByTagName("html")[0];
  if (getCurrentTheme() == "dark") {
    htmlTag.setAttribute("data-theme", "light");
  } else {
    htmlTag.setAttribute("data-theme", "dark");
  }
};

const initializeTheme = () => {
  const sunIcon = document.getElementById("sunIcon");
  const moonIcon = document.getElementById("moonIcon");

  if (getCurrentTheme() == "light") {
    sunIcon.classList.replace("swap-on", "swap-off");
    moonIcon.classList.replace("swap-off", "swap-on");
  }
};

initializeTheme();
