let messagesEl = document.getElementById("messages-inner");
let saveModal;

const apiHeaders = {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrfToken,
    "X-Device-ID": deviceId,
};

document.addEventListener("DOMContentLoaded", function() {
    saveModal = new bootstrap.Modal(document.getElementById("save-modal"));
});

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

function showSaveModal() {
    clearFieldErrors();

    let errs = {};
    if (!document.getElementById("first-name").value.trim()) errs.first_name = "First name is required";
    if (!document.getElementById("last-name").value.trim()) errs.last_name = "Last name is required";
    if (!document.getElementById("email").value.trim()) errs.email = "Email required";

    let pw = document.getElementById("password").value;
    if (!pw) {
        errs.password = "Password required";
    } else if (pw.length < 8) {
        errs.password = "Password too short, minimum 8 characters";
    } else if (pw.length > 72) {
        errs.password = "Password too long, maximum 72 characters";
    }

    if (Object.keys(errs).length > 0) {
        showFieldErrors(errs);
        return;
    }
    saveModal.show();
}

function confirmSave() {
    saveModal.hide();
    clearFieldErrors();

    let payload = {
        first_name: document.getElementById("first-name").value.trim(),
        last_name: document.getElementById("last-name").value.trim(),
        email: document.getElementById("email").value.trim(),
        password: document.getElementById("password").value,
        role: document.getElementById("role").value,
        is_active: document.getElementById("status").value === "1",
    };

    fetch(apiPort + "/api/admin/add-user", {
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
                showError(data.message || "Failed to create user");
            }
        } else {
            showSuccess("User created");
            setTimeout(function() {
                window.location.href = "/admin/all-users";
            }, 500);
        }
    })
    .catch(function() {
        showError("Failed to create user");
    });
}
