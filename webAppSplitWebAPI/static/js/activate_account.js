let activateMessages = document.getElementById("activate-messages");

function showError(msg) {

    activateMessages.classList.add("alert-danger");
    activateMessages.classList.remove("alert-success");
    activateMessages.innerText = msg;
}

function showSuccess(msg) {

    activateMessages.classList.remove("alert-danger");
    activateMessages.classList.add("alert-success");
    activateMessages.innerHTML = msg + '<br><a href="/login">Go to login</a>';
}

fetch(apiPort + "/api/activate-account", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
        email: activateEmail,
        signed_url: signedUrl,
    })
})
.then(r => r.json())
.then(data => {
    if (!data.error) {
        showSuccess(data.message);
    } else {
        showError(data.message);
    }
})
.catch(() => {
    showError("Something went wrong");
})
