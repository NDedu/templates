let messagesEl = document.getElementById("messages-inner");

function showSuccess(msg) {
    messagesEl.classList.remove("d-none", "alert-danger");
    messagesEl.classList.add("alert-success");
    messagesEl.innerText = msg;
}

function showError(msg) {
    messagesEl.classList.remove("d-none", "alert-success");
    messagesEl.classList.add("alert-danger");
    messagesEl.innerText = msg;
}

const apiHeaders = {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrfToken,
    "X-Device-ID": deviceId,
};

const requestOptions = {
    credentials: "include",
    method: 'GET',
    headers: apiHeaders,
}

let bookUUID = window.location.pathname.split("/").pop();
let editing = false;
let saveModal;
let deleteModal;

document.addEventListener("DOMContentLoaded", function() {
    if (isAdmin) {
        saveModal = new bootstrap.Modal(document.getElementById("save-modal"));
        deleteModal = new bootstrap.Modal(document.getElementById("delete-modal"));
    }
});

fetch(apiPort + "/api/book/" + bookUUID, requestOptions)
.then(response => response.json())
.then(function(data) {
    if (data && data.name !== undefined) {
        document.getElementById("name").innerText = data.name;
        document.getElementById("description").innerText = data.description;

        document.getElementById("name-input").value = data.name;
        document.getElementById("description-input").value = data.description;

        if (data.image) {
            document.getElementById("book-image").src = data.image;
            document.getElementById("book-image-wrapper").classList.remove("d-none");
        }
    }
})
.catch(function() {
    showError("Failed to load book");
});

function toggleEdit() {
    let btn = document.getElementById("delete-cancel-btn");
    if (!editing) {
        document.querySelectorAll(".book-display").forEach(el => el.classList.add("d-none"));
        document.querySelectorAll(".book-edit").forEach(el => el.classList.remove("d-none"));
        document.getElementById("edit-btn").innerHTML = "Save";
        btn.innerHTML = "Cancel";
        btn.setAttribute("onclick", "cancelEdit()");
        editing = true;
    } else {
        saveModal.show();
    }
}

function cancelEdit() {
    clearFieldErrors();
    let btn = document.getElementById("delete-cancel-btn");
    document.querySelectorAll(".book-display").forEach(el => el.classList.remove("d-none"));
    document.querySelectorAll(".book-edit").forEach(el => el.classList.add("d-none"));
    document.getElementById("edit-btn").innerHTML = "Edit";
    btn.innerHTML = "Delete";
    btn.setAttribute("onclick", "deleteBook()");
    editing = false;

    document.getElementById("name-input").value = document.getElementById("name").innerText;
    document.getElementById("description-input").value = document.getElementById("description").innerText;
}

function deleteBook() {
    deleteModal.show();
}

function confirmDeleteBook() {
    deleteModal.hide();

    fetch(apiPort + "/api/admin/delete-book/" + bookUUID, {
        credentials: "include",
        method: "DELETE",
        headers: apiHeaders,
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            showError(data.message || "Failed to delete book");
        } else {
            window.location.href = "/all-books";
        }
    })
    .catch(function() {
        showError("Failed to delete book");
    });
}

function saveBook() {
    saveModal.hide();
    clearFieldErrors();

    let name = document.getElementById("name-input").value.trim();
    let description = document.getElementById("description-input").value.trim();

    fetch(apiPort + "/api/admin/edit-book", {
        credentials: "include",
        method: "POST",
        headers: apiHeaders,
        body: JSON.stringify({
            uuid: bookUUID,
            name: name,
            description: description,
        }),
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            if (data.errors) {
                showFieldErrors(data.errors);
            } else {
                showError(data.message || "Failed to update book");
            }
        } else {
            document.getElementById("name").innerText = name;
            document.getElementById("description").innerText = description;

            let btn = document.getElementById("delete-cancel-btn");
            document.querySelectorAll(".book-display").forEach(el => el.classList.remove("d-none"));
            document.querySelectorAll(".book-edit").forEach(el => el.classList.add("d-none"));
            document.getElementById("edit-btn").innerHTML = "Edit";
            btn.innerHTML = "Delete";
            btn.setAttribute("onclick", "deleteBook()");
            editing = false;

            showSuccess("Book updated");
        }
    })
    .catch(function() {
        showError("Failed to update book");
    });
}
