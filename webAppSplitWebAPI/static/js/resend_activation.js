let resendMessages = document.getElementById("resend-activation-messages");

function showError(msg) {

    resendMessages.classList.add("alert-danger");
    resendMessages.classList.remove("alert-success");
    resendMessages.classList.remove("d-none");
    resendMessages.innerText = msg;
}

function showSuccess(msg) {

    resendMessages.classList.remove("alert-danger");
    resendMessages.classList.add("alert-success");
    resendMessages.classList.remove("d-none");
    resendMessages.innerText = msg;
}

document.getElementById("resend_activation_form").addEventListener("submit", function(e) {
    e.preventDefault();

    let form = this;
    if (!form.checkValidity()) {
        form.classList.add("was-validated");
        return;
    }
    form.classList.add("was-validated");

    fetch(apiPort + "/api/resend-activation", {
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
            showSuccess(data.message);
            let btn = document.getElementById("resend_btn");
            btn.disabled = true;
            btn.innerText = "Try again in 10s";
            let seconds = 10;
            let countdown = setInterval(() => {
                seconds--;
                btn.innerText = "Try again in " + seconds + "s";
                if (seconds <= 0) {
                    clearInterval(countdown);
                    btn.disabled = false;
                    btn.innerText = "Resend activation link";
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
