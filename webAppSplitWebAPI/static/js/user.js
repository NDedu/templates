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

let uuid = window.location.pathname.split("/").pop()
let isSelf = currentUserUUID === uuid;
let editing = false;
let saveModal;
let deleteModal;
let resetPasswordModal;

document.addEventListener("DOMContentLoaded", function() {
    saveModal = new bootstrap.Modal(document.getElementById("save-modal"));
    deleteModal = new bootstrap.Modal(document.getElementById("delete-modal"));
    resetPasswordModal = new bootstrap.Modal(document.getElementById("reset-password-modal"));

    if (isSelf) {
        document.getElementById("delete-cancel-btn").classList.add("d-none");
    }
});

fetch(apiPort + "/api/admin/get-user/" + uuid, requestOptions)
.then(response => response.json())
.then(function(data) {
    if (!data.error && data.user) {
        document.getElementById("first-name").innerHTML = data.user.first_name;
        document.getElementById("last-name").innerHTML = data.user.last_name;
        document.getElementById("email").innerHTML = data.user.email;
        document.getElementById("role").innerHTML = data.user.role;
        document.getElementById("status").innerHTML = data.user.is_active ? "Active" : "Inactive";

        document.getElementById("first-name-input").value = data.user.first_name;
        document.getElementById("last-name-input").value = data.user.last_name;
        document.getElementById("email-input").value = data.user.email;
        document.getElementById("role-input").value = data.user.role;
        document.getElementById("status-input").value = data.user.is_active ? "1" : "0";

        showStatusBadge(data.user.is_active);
    }
})

function toggleEdit() {
    let btn = document.getElementById("delete-cancel-btn");
    if (!editing) {
        document.querySelectorAll(".user-display").forEach(el => el.classList.add("d-none"));
        document.querySelectorAll(".user-edit").forEach(el => el.classList.remove("d-none"));
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
    document.querySelectorAll(".user-display").forEach(el => el.classList.remove("d-none"));
    document.querySelectorAll(".user-edit").forEach(el => el.classList.add("d-none"));
    document.getElementById("edit-btn").innerHTML = "Edit";
    btn.innerHTML = "Delete";
    btn.setAttribute("onclick", "deleteUser()");
    editing = false;

    // Reset inputs to current display values
    document.getElementById("first-name-input").value = document.getElementById("first-name").innerHTML;
    document.getElementById("last-name-input").value = document.getElementById("last-name").innerHTML;
    document.getElementById("email-input").value = document.getElementById("email").innerHTML;
    document.getElementById("role-input").value = document.getElementById("role").innerHTML;
    document.getElementById("status-input").value = document.getElementById("status").innerHTML === "Active" ? "1" : "0";
}

function deleteUser() {
    deleteModal.show();
}

function confirmDeleteUser() {
    deleteModal.hide();

    fetch(apiPort + "/api/admin/delete-user/" + uuid, {
        credentials: "include",
        method: "DELETE",
        headers: apiHeaders,
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            showError(data.message || "Failed to delete user");
        } else {
            sendWs({
                action: "deleteUser",
                user_uuid: uuid,
            });
            setTimeout(() => {
                window.location.href = "/admin/all-users";
            }, 100);
        }
    })
    .catch(function() {
        showError("Failed to delete user");
    });
}

function saveUser() {
    saveModal.hide();
    clearFieldErrors();

    let payload = {
        uuid: uuid,
        first_name: document.getElementById("first-name-input").value.trim(),
        last_name: document.getElementById("last-name-input").value.trim(),
        email: document.getElementById("email-input").value.trim(),
        role: document.getElementById("role-input").value,
        is_active: document.getElementById("status-input").value === "1",
    };

    fetch(apiPort + "/api/admin/edit-user", {
        credentials: "include",
        method: "POST",
        headers: apiHeaders,
        body: JSON.stringify(payload),
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            if (data.errors) {
                showFieldErrors(data.errors);
            } else {
                showError(data.message || "Failed to update user");
            }
        } else {
            document.getElementById("first-name").innerHTML = payload.first_name;
            document.getElementById("last-name").innerHTML = payload.last_name;
            document.getElementById("email").innerHTML = payload.email;
            document.getElementById("role").innerHTML = payload.role;
            document.getElementById("status").innerHTML = payload.is_active ? "Active" : "Inactive";

            let btn = document.getElementById("delete-cancel-btn");
            document.querySelectorAll(".user-display").forEach(el => el.classList.remove("d-none"));
            document.querySelectorAll(".user-edit").forEach(el => el.classList.add("d-none"));
            document.getElementById("edit-btn").innerHTML = "Edit";
            btn.innerHTML = "Delete";
            btn.setAttribute("onclick", "deleteUser()");
            editing = false;

            showStatusBadge(payload.is_active);
            showSuccess("User updated");
        }
    })
    .catch(function() {
        showError("Failed to update user");
    });
}

function resetPassword() {
    resetPasswordModal.show();
}

function confirmResetPassword() {
    resetPasswordModal.hide();

    let email = document.getElementById("email").innerHTML;

    fetch(apiPort + "/api/admin/reset-user-password", {
        credentials: "include",
        method: "POST",
        headers: apiHeaders,
        body: JSON.stringify({ email: email }),
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            showError(data.message);
        } else {
            showSuccess("Password reset email sent");
        }
    })
    .catch(function() {
        showError("Failed to send password reset email");
    });
}

function showStatusBadge(isActive) {
    if (isActive) {
        document.getElementById("active").classList.remove("d-none");
        document.getElementById("inactive").classList.add("d-none");
    } else {
        document.getElementById("active").classList.add("d-none");
        document.getElementById("inactive").classList.remove("d-none");
    }
}
