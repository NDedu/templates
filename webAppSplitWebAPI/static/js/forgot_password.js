let forgotPasswordMessages = document.getElementById("forgot-password-messages");

function showError(msg) {

    forgotPasswordMessages.classList.add("alert-danger");
    forgotPasswordMessages.classList.remove("alert-success");
    forgotPasswordMessages.classList.remove("d-none");
    forgotPasswordMessages.innerText = msg;
}

function showSuccess() {

    forgotPasswordMessages.classList.remove("alert-danger");
    forgotPasswordMessages.classList.add("alert-success");
    forgotPasswordMessages.classList.remove("d-none");
    forgotPasswordMessages.innerText = "Email sent";
}

document.getElementById("forgot_password_form").addEventListener("submit", function(e) {
    e.preventDefault();

    let form = this;
    if (!form.checkValidity()) {
        form.classList.add("was-validated");
        return;
    }
    form.classList.add("was-validated");

    fetch(apiPort + "/api/forgot-password", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            email: document.getElementById("email").value,
        })
    })
    .then(r => r.json())
    .then(data => {
        if (!data.error) {
            showSuccess();
            let btn = document.getElementById("reset_btn");
            btn.disabled = true;
            btn.innerText = "Try again in 10s";
            let seconds = 10;
            let countdown = setInterval(() => {
                seconds--;
                btn.innerText = "Try again in " + seconds + "s";
                if (seconds <= 0) {
                    clearInterval(countdown);
                    btn.disabled = false;
                    btn.innerText = "Reset password";
                }
            }, 1000);
        } else {
            showError(data.message);
        }
    })
    .catch(() => {
        showError("Something went wrong");
    })
})
