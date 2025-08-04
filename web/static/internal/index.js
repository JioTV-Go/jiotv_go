const search = (searchTerm) => {
  const channels = document.querySelectorAll('.card');
  const urlParams = new URLSearchParams(window.location.search);

  // Update URL search parameter
  if (searchTerm.trim() !== '') {
    urlParams.set('search', searchTerm);
  } else {
    urlParams.delete('search');
  }

  // Update the URL without reloading the page
  window.history.replaceState({}, '', `${window.location.pathname}?${urlParams}`);

  channels.forEach((channel) => {
    const name = channel.querySelector('.font-bold').textContent.toLowerCase();
    if (name.includes(searchTerm.toLowerCase())) {
      channel.style.display = 'block';
    } else {
      channel.style.display = 'none';
    }
  });
};

const init = () => {
  const searchInput = document.getElementById('portexe-search-input');

  // Check for search parameter on page load
  const urlParams = new URLSearchParams(window.location.search);
  const searchParam = urlParams.get('search');

  if (searchParam) {
    search(searchParam);
    searchInput.value = searchParam; // Set input value to the search parameter
  }

  searchInput.addEventListener('keyup', (e) => {
    search(e.target.value);
  });
};

// Call the init function to start the process
init();

const loginClick = () => {
  // create a popup to enter username and password
  // then redirect to /login?username=xxx&password=xxx
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;
  if (!username || !password) {
    return;
  }

  const url = "/login";

  fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  })
    .then((res) => res.json())
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
  // Fetch number from input
  const number = document.getElementById("number").value;
  if (!number) {
    return;
  }


  const url = "/login/sendOTP";

  fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ number: `+91${number}` }),
  })
    .then((res) => res.json())
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
  // Fetch number and OTP from input
  const number = document.getElementById("number").value;
  const otp = document.getElementById("otp").value;
  if (!number || !otp) {
    return;
  }

  const url = "/login/verifyOTP";

  fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ number: `+91${number}`, otp }),
  })
    .then((res) => res.json())
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
