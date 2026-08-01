let registerMessages = document.getElementById("register-messages");

function showError(msg) {

    registerMessages.classList.add("alert-danger");
    registerMessages.classList.remove("alert-success");
    registerMessages.classList.remove("d-none");
    registerMessages.innerText = msg;
}

function showSuccess(msg) {

    registerMessages.classList.remove("alert-danger");
    registerMessages.classList.add("alert-success");
    registerMessages.classList.remove("d-none");
    registerMessages.innerText = msg;
}

document.getElementById("register_form").addEventListener("submit", function(e) {
    e.preventDefault();
    clearFieldErrors();

    let form = this;
    if (!form.checkValidity()) {
        form.classList.add("was-validated");
        return;
    }
    form.classList.add("was-validated");

    let email = document.getElementById("email").value;
    let confirmEmail = document.getElementById("confirm-email").value;
    if (email !== confirmEmail) {
        showError("Emails do not match");
        return;
    }

    let passw = document.getElementById("password").value;
    let confirmPassw = document.getElementById("confirm-password").value;
    if (passw !== confirmPassw) {
        showError("Passwords do not match");
        return;
    }

    fetch(apiPort + "/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            first_name: document.getElementById("first-name").value.trim(),
            last_name: document.getElementById("last-name").value.trim(),
            email: email,
            confirm_email: confirmEmail,
            password: passw,
            confirm_password: confirmPassw,
        })
    })
    .then(r => r.json())
    .then(data => {
        if (!data.error) {
            showSuccess(data.message);
            document.getElementById("register_form").classList.add("d-none");
        } else if (data.errors) {
            showFieldErrors(data.errors);
        } else {
            showError(data.message);
        }
    })
    .catch(() => {
        showError("Something went wrong");
    })
})
