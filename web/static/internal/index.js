const search = (searchTerm) => {
  const channels = document.querySelectorAll('.card');
  
  // Update URL search parameter
  updateUrlParameter('search', searchTerm);

  channels.forEach((channel) => {
    const nameElement = channel.querySelector('.font-bold');
    if (nameElement) {
      const name = nameElement.textContent.toLowerCase();
      channel.style.display = name.includes(searchTerm.toLowerCase()) ? 'block' : 'none';
    }
  });
};

const init = () => {
  const searchInput = safeGetElementById('portexe-search-input');

  // Check for search parameter on page load
  const urlParams = getCurrentUrlParams();
  const searchParam = urlParams.get('search');

  if (searchParam && searchInput) {
    search(searchParam);
    searchInput.value = searchParam;
  }

  if (searchInput) {
    searchInput.addEventListener('keyup', (e) => {
      search(e.target.value);
    });
  }
};

// Call the init function to start the process
init();

const loginClick = () => {
  const elements = safeGetElementsById(["username", "password"]);
  const { username: usernameElement, password: passwordElement } = elements;
  
  if (!usernameElement || !passwordElement) {
    return;
  }

  const username = usernameElement.value;
  const password = passwordElement.value;
  
  if (!username || !password) {
    return;
  }

  postJSON("/login", { username, password })
    .then((data) => {
      if (data.status === "success") {
        alert("Login success. Enjoy!");
        document.location.reload();
      } else {
        alert("Login failed!");
      }
    })
    .catch((err) => {
      console.log(err);
      alert("Login failed!");
    });
};

const loginOTPClick = () => {
  const numberElement = safeGetElementById("number");
  if (!numberElement) {
    return;
  }

  const number = numberElement.value;
  if (!number) {
    return;
  }

  postJSON("/login/sendOTP", { number: `+91${number}` })
    .then((data) => {
      if (data.status) {
        verify_otp_modal.showModal(); // skipcq: JS-0125
      } else {
        alert("Sending OTP failed!");
      }
    })
    .catch((err) => {
      console.log(err);
      alert("Sending OTP failed!");
    });
};

const loginOTPVerifyClick = () => {
  const elements = safeGetElementsById(["number", "otp"]);
  const { number: numberElement, otp: otpElement } = elements;
  
  if (!numberElement || !otpElement) {
    return;
  }

  const number = numberElement.value;
  const otp = otpElement.value;
  
  if (!number || !otp) {
    return;
  }

  postJSON("/login/verifyOTP", { number: `+91${number}`, otp })
    .then((data) => {
      if (data.status) {
        alert("OTP verification success. Enjoy!");
        document.location.reload();
      } else {
        alert("OTP verification failed!");
      }
    })
    .catch((err) => {
      console.log(err);
      alert("OTP verification failed!");
    });
};
