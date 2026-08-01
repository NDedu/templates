let tbody = document.getElementById("books-table").getElementsByTagName("tbody")[0];

let pageSize = 5;
let addBookModal;

function currentSortFromURL() {
    let params = new URLSearchParams(window.location.search);
    let col = params.get("sort") || "name";
    let dir = params.get("dir") === "desc" ? "desc" : "asc";
    return { col: col, dir: dir };
}

let currentSort = currentSortFromURL();

const apiHeaders = {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrfToken,
    "X-Device-ID": deviceId,
};

function currentPageFromURL() {
    let params = new URLSearchParams(window.location.search);
    let p = parseInt(params.get("page"));
    return p > 0 ? p : 1;
}

function setURLState(page, sortCol, sortDir) {
    let url = new URL(window.location);
    url.searchParams.set("page", page);
    url.searchParams.set("sort", sortCol);
    url.searchParams.set("dir", sortDir);
    history.replaceState(null, "", url);
}

function updateSortIcons() {
    document.querySelectorAll("#books-table th.sortable").forEach(function(th) {
        let icon = th.querySelector(".sort-icon");
        if (th.getAttribute("data-sort") === currentSort.col) {
            icon.textContent = currentSort.dir === "asc" ? "\u2191" : "\u2193";
            icon.style.opacity = "1";
        } else {
            icon.textContent = "\u2195";
            icon.style.opacity = "0.4";
        }
    });
}

function paginator(lastPage, currPg) {
    let p = document.getElementById("paginator");
    let html = "";

    html += `<li class="page-item ${currPg <= 1 ? "disabled" : ""}"><a href="#!" class="page-link page" data-page="${currPg - 1}">&lt;</a></li>`;

    for (let i = 1; i <= lastPage; i++) {
        html += `<li class="page-item ${i === currPg ? "active" : ""}"><a href="#!" class="page-link page" data-page="${i}">${i}</a></li>`;
    }

    html += `<li class="page-item ${currPg >= lastPage ? "disabled" : ""}"><a href="#!" class="page-link page" data-page="${currPg + 1}">&gt;</a></li>`;

    p.innerHTML = html;

    p.querySelectorAll(".page").forEach(function(link) {
        link.addEventListener("click", function(e) {
            e.preventDefault();
            let li = e.target.closest("li");
            if (li && li.classList.contains("disabled")) return;
            let page = parseInt(e.currentTarget.getAttribute("data-page"));
            if (page > 0) updateTable(pageSize, page);
        });
    });
}

function updateTable(pgSize, currPg) {
    let body = {
        page_size: pgSize,
        current_page: currPg,
        sort_col: currentSort.col,
        sort_dir: currentSort.dir,
    };

    fetch(apiPort + "/api/all-books", {
        method: "POST",
        headers: apiHeaders,
        body: JSON.stringify(body),
    })
    .then(function(response) {
        if (!response.ok) throw new Error("request failed: " + response.status);
        return response.json();
    })
    .then(function(data) {
        tbody.innerHTML = "";

        if (data.error || !data.books || data.books.length === 0) {
            let newRow = tbody.insertRow();
            let newCell = newRow.insertCell();
            newCell.colSpan = 3;
            newCell.innerText = "No books found";
        } else {
            data.books.forEach(function(i) {
                let newRow = tbody.insertRow();
                let newCell = newRow.insertCell();

                let item = document.createTextNode(i.name);
                newCell.appendChild(item);

                newCell = newRow.insertCell();
                let desc = i.description;
                if (desc.length > 100) {
                    desc = desc.substring(0, 100) + "...";
                }
                item = document.createTextNode(desc);
                newCell.appendChild(item);

                newCell = newRow.insertCell();
                newCell.innerHTML = '<a href="/books/' + i.uuid + '" title="Open" style="text-decoration:none">&#x2192;</a>';
            });
        }

        setURLState(data.current_page, currentSort.col, currentSort.dir);
        paginator(data.last_page, data.current_page);
        updateSortIcons();
    })
    .catch(function(err) {
        tbody.innerHTML = "";
        let newRow = tbody.insertRow();
        let newCell = newRow.insertCell();
        newCell.colSpan = 3;
        newCell.innerText = "Failed to load books";
        console.error(err);
    });
}

document.querySelectorAll("#books-table th.sortable").forEach(function(th) {
    th.addEventListener("click", function() {
        let col = th.getAttribute("data-sort");
        if (currentSort.col === col) {
            currentSort.dir = currentSort.dir === "asc" ? "desc" : "asc";
        } else {
            currentSort.col = col;
            currentSort.dir = "asc";
        }
        updateTable(pageSize, 1);
    });
});

document.addEventListener("DOMContentLoaded", function() {
    let modalEl = document.getElementById("add-book-modal");
    if (modalEl) {
        addBookModal = new bootstrap.Modal(modalEl);
    }
});

function showAddModal() {
    document.getElementById("add-name-input").value = "";
    document.getElementById("add-description-input").value = "";
    clearFieldErrors();
    addBookModal.show();
}

function confirmAddBook() {
    clearFieldErrors();
    let name = document.getElementById("add-name-input").value.trim();
    let description = document.getElementById("add-description-input").value.trim();

    if (!name) {
        showFieldErrors({ name: "Book name is required" });
        return;
    }

    fetch(apiPort + "/api/admin/add-book", {
        method: "POST",
        headers: apiHeaders,
        body: JSON.stringify({ name: name, description: description }),
    })
    .then(function(response) { return response.json(); })
    .then(function(data) {
        if (data.error) {
            if (data.errors) {
                showFieldErrors(data.errors);
            } else {
                addBookModal.hide();
                alert(data.message || "Failed to add book");
            }
        } else {
            addBookModal.hide();
            updateTable(pageSize, currentPageFromURL());
        }
    })
    .catch(function() {
        addBookModal.hide();
        alert("Failed to add book");
    });
}

updateTable(pageSize, currentPageFromURL());
