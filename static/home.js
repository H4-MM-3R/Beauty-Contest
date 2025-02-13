document.getElementById("createHubBtn").addEventListener("click", function () {
    fetch("/create-hub", {
        method: "POST",
    })
        .then((response) => response.json())
        .then((data) => {
            if (data.hub) {
                let name = prompt("Enter your name:");
                if (!name) {
                    alert("Name is required");
                    return;
                }
                // Redirect to the game page with hub hash and name.
                window.location.href =
                    "/" + data.hub + "?name=" + encodeURIComponent(name);
            }
        })
        .catch((err) => {
            alert("Error creating hub: " + err);
        });
});

document.getElementById("joinHubForm").addEventListener("submit", function (e) {
    e.preventDefault();
    let hubHash = document.getElementById("hubInput").value.trim();
    let name = document.getElementById("username").value.trim();
    if (hubHash && name) {
        window.location.href = "/" + hubHash + "?name=" + encodeURIComponent(name);
    }
});
