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



const showOTPError = (message) => {
  safeGetElementById("otp-error-message").textContent = message;
  otp_error_modal.showModal(); // skipcq: JS-0125
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
        showOTPError("We couldn’t send the OTP. Check your number and try again.");
      }
    })
    .catch((err) => {
      console.log(err);
      showOTPError("We couldn’t send the OTP. Check your connection and try again.");
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
        verify_otp_modal.close(); // skipcq: JS-0125
        otp_success_modal.showModal(); // skipcq: JS-0125
      } else {
        showOTPError("The OTP is incorrect or expired. Try again.");
      }
    })
    .catch((err) => {
      console.log(err);
      showOTPError("We couldn’t verify the OTP. Check your connection and try again.");
    });
};
