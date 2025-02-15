const hubHash = document
  .getElementById("gameContainer")
  .getAttribute("data-hubhash");
const name = document.getElementById("gameContainer").getAttribute("data-name");

const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const ws = new WebSocket( protocol + "//" + window.location.host + "/ws?hub=" + hubHash + "&name=" + encodeURIComponent(name));

ws.onopen = function () {
  console.log("Connected to hub " + hubHash);
};

ws.onmessage = function (event) {
  try {
    const data = JSON.parse(event.data);
    console.log("data is", data);

    const activePlayers = [];
    const eliminatedPlayers = [];
    data.players.forEach(function (player) {
      if (player.eliminated) {
        eliminatedPlayers.push(player);
      } else {
        activePlayers.push(player);
      }
    });

    let activeHTML = "";
    activePlayers.forEach(function (player) {
      activeHTML += "<div class='activePlayerCard'>";
      activeHTML += "<p class='activePlayerName'>" + player.name + "</p>";
      activeHTML += "<p class='activePlayerScore'>" + player.score + "</p>";
      if (player.response) {
        activeHTML += "<p class='activePlayerResponse'>";
        activeHTML +=
          typeof player.response === "number" ? "Responded" : player.response;
        activeHTML += "</p>";
      }
      activeHTML += "</div>";
    });
    document.getElementById("activeList").innerHTML = activeHTML;

    let elimHTML = "";
    eliminatedPlayers.forEach(function (player) {
      elimHTML += "<li>" + player.name + " (Score: " + player.score + ")</li>";
    });
    document.getElementById("eliminatedList").innerHTML = elimHTML;

    const player = data.players.find((p) => p.name === name);
    if (player && player.eliminated) {
      alert("You have been eliminated from the game.");
    }

    if (data.type === "result") {
      let info = "<h3>Round Result</h3>";
      info += "<h4>Responses:</h4><ul>";
      data.players.forEach(function (player) {
        info += "<li>" + player.name + ": " + player.response + "</li>";
      });
      info += "</ul>";
      info += "<p>Target: " + data.target.toFixed(2) + "</p>";
      info += "<p>Winner(s): " + data.winners.join(", ") + "</p>";
      document.getElementById("roundInfo").innerHTML = info;
      document.getElementById("roundInfo").style.display = "block";
    } else if (data.type === "gameover") {
      let leaderboardHTML = "<h3>Leaderboard</h3><ol>";
      data.players
        .sort((a, b) => b.score - a.score)
        .forEach(function (player) {
          leaderboardHTML +=
            "<li>" + player.name + " (Score: " + player.score + ")</li>";
        });
      leaderboardHTML += "</ol>";
      document.getElementById("leaderboard").innerHTML = leaderboardHTML;
      document.getElementById("gameOverDialog").style.display = "block";
    } else {
      document.getElementById("roundInfo").innerHTML = "";
      document.getElementById("roundInfo").style.display = "none";
    }
  } catch (e) {
    console.error("Error parsing message", e);
  }
};

ws.onclose = function () {
  alert("WebSocket connection closed.");
};

document.getElementById("inviteBtn").addEventListener("click", function () {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  navigator.clipboard.writeText(
    window.location.protocol + "//" + window.location.host + "/" + hubHash,
  );
});

let numberButtons = "";
for (let i = 0; i < 9; i++) {
  numberButtons += "<button class='nullButton'" + i + "</button>";
}
for (let i = 0; i < 101; i++) {
  numberButtons +=
    "<button class='numButton' data-number='" + i + "'>" + i + "</button>";
}
document.getElementById("numberGrid").innerHTML = numberButtons;

document.querySelectorAll(".numButton").forEach(function (button) {
  button.addEventListener("click", function () {
    const num = button.getAttribute("data-number");
    ws.send(num);
    console.log("Sent number:", num);
  });
});

document.getElementById("restartBtn").addEventListener("click", function () {
  document.getElementById("gameOverDialog").style.display = "none";
});
document.getElementById("exitBtn").addEventListener("click", function () {
  window.location.href = "/";
});
