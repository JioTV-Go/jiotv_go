const fetchTableData = async () => {
  return new Promise((resolve) => {
    // Simulate an asynchronous data fetching operation
    setTimeout(() => {
      const searchData = [];
      const tableEl = document.getElementById("portexe-data-table");

      Array.from(tableEl.children[1].children).forEach((_bodyRowEl) => {
        const rowData = Array.from(_bodyRowEl.children).map(
          (_celEl) => _celEl.innerHTML
        );
        searchData.push(rowData);
      }); // tbody

      resolve(searchData);
    }, 1000); // Simulated delay of 1 second
  });
};

const search = (arr, searchTerm) => {
  if (!searchTerm) return arr;
  return arr.filter((_row) =>
    _row[1].toLowerCase().includes(searchTerm.toLowerCase())
  );
};

const refreshTable = (data) => {
  const tableBody = document.getElementById("portexe-data-table").children[1];
  tableBody.innerHTML = "";

  data.forEach((_row) => {
    const curRow = document.createElement("tr");
    _row.forEach((_dataItem) => {
      const curCell = document.createElement("td");
      curCell.innerHTML = _dataItem;
      curRow.appendChild(curCell);
    });

    tableBody.appendChild(curRow);
  });
};

const init = async () => {
  const initialTableData = await fetchTableData();

  const searchInput = document.getElementById("portexe-search-input");
  searchInput.addEventListener("keyup", async (e) => {
    const filteredData = search(initialTableData, e.target.value);
    refreshTable(filteredData);
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
