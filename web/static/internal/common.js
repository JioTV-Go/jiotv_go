const htmlTag = document.getElementsByTagName("html")[0];

const getCurrentTheme = () => {
  if (localStorage.getItem("theme")) {
    // return local storage theme
    return localStorage.getItem("theme");
  } else if (htmlTag.hasAttribute("data-theme")) {
    // return data-theme attribute
    localStorage.setItem("theme", htmlTag.getAttribute("data-theme"));
    return htmlTag.getAttribute("data-theme");
  } else {
    // return system theme
    if (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    ) {
      localStorage.setItem("theme", "dark");
      return "dark";
    }
    localStorage.setItem("theme", "light");
    return "light";
  }
};

const toggleTheme = () => {
  // toggle or add attribute "data-theme" to html tag
  if (getCurrentTheme() == "dark") {
    localStorage.setItem("theme", "light");
    htmlTag.setAttribute("data-theme", "light");
  } else {
    localStorage.setItem("theme", "dark");
    htmlTag.setAttribute("data-theme", "dark");
  }
};

const initializeTheme = () => {
  const sunIcon = document.getElementById("sunIcon");
  const moonIcon = document.getElementById("moonIcon");

  if (getCurrentTheme() == "light") {
    sunIcon.classList.replace("swap-on", "swap-off");
    moonIcon.classList.replace("swap-off", "swap-on");
    htmlTag.setAttribute("data-theme", "light");
  }
};

initializeTheme();
