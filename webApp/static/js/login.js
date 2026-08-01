let loginMessages = document.getElementById("login-messages");

function showError(msg) {

    loginMessages.classList.add("alert-danger");
    loginMessages.classList.remove("alert-success");
    loginMessages.classList.remove("d-none");
    loginMessages.innerText = msg;
}

function showSuccess() {

    loginMessages.classList.remove("alert-danger");
    loginMessages.classList.add("alert-success");
    loginMessages.classList.remove("d-none");
    loginMessages.innerText = "Login successful";
}

document.getElementById("login_form").addEventListener("submit", function(e) {
    e.preventDefault();
    clearFieldErrors();

    let form = this;
    if (!form.checkValidity()) {
        form.classList.add("was-validated");
        return;
    }
    form.classList.add("was-validated");

    fetch(apiPort + "/api/authenticate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            email: document.getElementById("email").value,
            password: document.getElementById("password").value,
        })
    })
    .then(r => r.json())
    .then(data => {
        if (!data.error) {
            showSuccess();
            setTimeout(() => { location.href = "/"; }, 500);
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
