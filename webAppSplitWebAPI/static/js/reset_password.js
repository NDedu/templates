let resetPasswordMessages = document.getElementById("reset-password-messages");

function showError(msg) {

    resetPasswordMessages.classList.add("alert-danger");
    resetPasswordMessages.classList.remove("alert-success");
    resetPasswordMessages.classList.remove("d-none");
    resetPasswordMessages.innerText = msg;
}

function showSuccess() {

    resetPasswordMessages.classList.remove("alert-danger");
    resetPasswordMessages.classList.add("alert-success");
    resetPasswordMessages.classList.remove("d-none");
    resetPasswordMessages.innerText = "Password reset";
}

document.getElementById("reset_password_form").addEventListener("submit", function(e) {
    e.preventDefault();
    clearFieldErrors();

    let form = this;
    if (!form.checkValidity()) {
        form.classList.add("was-validated");
        return;
    }
    form.classList.add("was-validated");

    if(document.getElementById("password").value !== document.getElementById("verify-password").value) {

        showError("Passwords do not match!")
        return
    }

    fetch(apiPort + "/api/reset-password", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            password: document.getElementById("password").value,
            email: resetEmail,
            signed_url: signedUrl,
        })
    })
    .then(r => r.json())
    .then(data => {
        if (!data.error) {
            showSuccess();
            setTimeout(function() {
                location.href = "/login"
            }, 2000)
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
