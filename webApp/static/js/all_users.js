let tbody = document.getElementById("users-table").getElementsByTagName("tbody")[0];

let pageSize = 5;

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
    document.querySelectorAll("#users-table th.sortable").forEach(function(th) {
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

    fetch(apiPort + "/api/admin/all-users", {
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

        if (data.error || !data.users || data.users.length === 0) {
            let newRow = tbody.insertRow();
            let newCell = newRow.insertCell();
            newCell.colSpan = 5;
            newCell.innerText = "No users found";
        } else {
            data.users.forEach(function(i) {
                let newRow = tbody.insertRow();
                let newCell = newRow.insertCell();

                let item = document.createTextNode(i.last_name + ", " + i.first_name);
                newCell.appendChild(item);

                newCell = newRow.insertCell();
                item = document.createTextNode(i.email);
                newCell.appendChild(item);

                newCell = newRow.insertCell();
                item = document.createTextNode(i.role);
                newCell.appendChild(item);

                newCell = newRow.insertCell();
                if (i.is_active) {
                    newCell.innerHTML = '<span class="badge bg-success">Active</span>';
                } else {
                    newCell.innerHTML = '<span class="badge bg-danger">Inactive</span>';
                }

                newCell = newRow.insertCell();
                newCell.innerHTML = '<a href="/admin/users/' + i.uuid + '" title="Open" style="text-decoration:none">&#x2192;</a>';
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
        newCell.colSpan = 5;
        newCell.innerText = "Failed to load users";
        console.error(err);
    });
}

document.querySelectorAll("#users-table th.sortable").forEach(function(th) {
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

updateTable(pageSize, currentPageFromURL());
