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
        document.getElementById("first-name").innerText = data.user.first_name;
        document.getElementById("last-name").innerText = data.user.last_name;
        document.getElementById("email").innerText = data.user.email;
        document.getElementById("role").innerText = data.user.role;
        document.getElementById("status").innerText = data.user.is_active ? "Active" : "Inactive";

        document.getElementById("first-name-input").value = data.user.first_name;
        document.getElementById("last-name-input").value = data.user.last_name;
        document.getElementById("email-input").value = data.user.email;
        document.getElementById("role-input").value = data.user.role;
        document.getElementById("status-input").value = data.user.is_active ? "1" : "0";

        let img = document.getElementById("profile-image");
        if (data.user.profile_image) {
            let tempImg = new Image();
            tempImg.onload = () => {
                img.src = data.user.profile_image;
                img.classList.remove("invisible");
            };
            tempImg.src = data.user.profile_image;
        } else {
            img.classList.remove("invisible");
        }

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
    document.getElementById("first-name-input").value = document.getElementById("first-name").innerText;
    document.getElementById("last-name-input").value = document.getElementById("last-name").innerText;
    document.getElementById("email-input").value = document.getElementById("email").innerText;
    document.getElementById("role-input").value = document.getElementById("role").innerText;
    document.getElementById("status-input").value = document.getElementById("status").innerText === "Active" ? "1" : "0";
}

function deleteUser() {
    deleteModal.show();
}

function confirmDeleteUser() {
    deleteModal.hide();

    fetch(apiPort + "/api/admin/delete-user/" + uuid, {
            method: "DELETE",
        headers: apiHeaders,
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            showError(data.message || "Failed to delete user");
        } else {
            window.location.href = "/admin/all-users";
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
            document.getElementById("first-name").innerText = payload.first_name;
            document.getElementById("last-name").innerText = payload.last_name;
            document.getElementById("email").innerText = payload.email;
            document.getElementById("role").innerText = payload.role;
            document.getElementById("status").innerText = payload.is_active ? "Active" : "Inactive";

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
    document.getElementById("reset-password-input").value = "";
    document.getElementById("reset-confirm-password-input").value = "";
    clearFieldErrors();
    resetPasswordModal.show();
}

function confirmResetPassword() {
    clearFieldErrors();

    let pw = document.getElementById("reset-password-input").value;
    let confirmPw = document.getElementById("reset-confirm-password-input").value;

    let errs = {};
    if (!pw) {
        errs.password = "Password required";
    } else if (pw.length < 8) {
        errs.password = "Password too short, minimum 8 characters";
    } else if (pw.length > 72) {
        errs.password = "Password too long, maximum 72 characters";
    }
    if (pw !== confirmPw) {
        errs.confirm_password = "Passwords do not match";
    }
    if (Object.keys(errs).length > 0) {
        showFieldErrors(errs);
        return;
    }

    resetPasswordModal.hide();

    fetch(apiPort + "/api/admin/reset-user-password", {
        method: "POST",
        headers: apiHeaders,
        body: JSON.stringify({
            uuid: uuid,
            password: pw,
            confirm_password: confirmPw,
        }),
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            if (data.errors) {
                showFieldErrors(data.errors);
            } else {
                showError(data.message);
            }
        } else {
            document.getElementById("reset-password-input").value = "";
            document.getElementById("reset-confirm-password-input").value = "";
            showSuccess("Password updated");
        }
    })
    .catch(function() {
        showError("Failed to reset password");
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

function uploadProfileImage(input) {
    if (!input.files || !input.files[0]) return;

    let formData = new FormData();
    formData.append("profile_image", input.files[0]);

    fetch(apiPort + "/api/admin/upload-profile-image/" + uuid, {
        method: "POST",
        headers: {
            "X-CSRF-Token": csrfToken,
            "X-Device-ID": deviceId,
        },
        body: formData,
    })
    .then(response => response.json())
    .then(function(data) {
        if (data.error) {
            showError(data.message || "Failed to upload image");
        } else {
            document.getElementById("profile-image").src = data.profile_image + "?v=" + Date.now();
            showSuccess("Profile image updated");
        }
    })
    .catch(function() {
        showError("Failed to upload image");
    })
    .finally(function() {
        input.value = "";
    });
}
