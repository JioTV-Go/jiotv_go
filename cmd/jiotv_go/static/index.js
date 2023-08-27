// Two dimensional array
const tableData = () => {
    const searchData = [];
    const tableEl = document.getElementById("portexe-data-table");
    // gets an HTML collection
    // console.log(tableEl.children);
  
    // .from creates an array from the HTML Collection
    // console.log(Array.from(tableEl.children[1].children));
    Array.from(tableEl.children[1].children).forEach(_bodyRowEl => {
      searchData.push(
        Array.from(_bodyRowEl.children).map(_celEl => {
          return _celEl.innerHTML;
        })
      );
    }); // tbody
    return searchData;
  };
  
  const search = (arr, searchTerm) => {
    if (!searchTerm) return arr;
    return arr.filter(_row => {
      return _row.find(_item =>
        _item.toLowerCase().includes(searchTerm.toLowerCase())
      );
    });
  };
  
  // Refresh table
  const refreshTable = data => {
    const tableBody = document.getElementById("portexe-data-table").children[1];
    tableBody.innerHTML = "";
  
    data.forEach(_row => {
      const curRow = document.createElement("tr");
      _row.forEach(_dataItem => {
        // if _dataItem is a <button> element
        const curCell = document.createElement("td");
        curCell.innerHTML = _dataItem;
        curRow.appendChild(curCell);
      });
  
      tableBody.appendChild(curRow);
    });
  };
  
  // Put in document
  const init = () => {
  
    const initialTableData = tableData();
  
    const searchInput = document.getElementById("portexe-search-input");
    searchInput.addEventListener("keyup", e => {
      // console.log(search(initialTableData, e.target.value));
      refreshTable(search(initialTableData, e.target.value));
    });
  };
  
init();

loginClick = () => {
  // create a popup to enter username and password
  // then redirect to /login?username=xxx&password=xxx
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;
  if (!username || !password) {
    return;
  }
  const url = "/login?username=" + username + "&password=" + password;
  
  fetch(url)
  .then((response) => {
    if (response.status == 200) {
      alert("Login success! Enjoy!");
    } else {
      alert("Login failed! Please try again!");
    }
  }).catch((error) => {
    console.log(error);
  });
}
