console.log("hola, mundo (desde js)!")

// function made by chatgpt!
document.addEventListener("htmx:afterRequest", function(evt) {
    if (evt.detail.xhr.getResponseHeader("HX-Location")) {
        setTimeout(() => {
            window.location.reload(); // Forces full reload after redirect
        }, 100); // Small delay to ensure redirection is processed
    }
});

//function made by chatgpt!
function handleResponse(event) {
    let xhr = event.detail.xhr;
    let status = xhr.getResponseHeader("HX-Status");
    let message = xhr.getResponseHeader("HX-Message");
    
    // mainly use for panel changes, when there's no need to reload
    if (status === "202") {
        return ;
    } else if (status === "200") {
        alert(message);
    } else if (status === "400") {
        alert(message);
    } else {
        alert("Something went wrong. Please try again.");
    }
}

// function made by chatgpt!
function validateForm(event) {
    let form = event.target.closest("form");
    if (!form.checkValidity()) {
        alert("Please finish filling the form! Check if you miss any required ('*') field!");
        event.preventDefault(); // Prevent HTMX from sending request
        form.reportValidity(); // Show default validation messages
    }
}

function revealPass(element) {
    const passInput = element.previousElementSibling;

    if (passInput.type === "password") {
        passInput.type = "text";
        element.classList.remove("fa-eye");
        element.classList.add("fa-eye-slash");
    } else if (passInput.type === "text") {
        passInput.type = "password";
        element.classList.add("fa-eye");
        element.classList.remove("fa-eye-slash");
    }

}

