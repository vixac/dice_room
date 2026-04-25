function copyRoomLink() {
  navigator.clipboard.writeText(window.location.href).then(() => {
    const btn = document.getElementById("share-btn");
    btn.textContent = "Copied!";
    setTimeout(() => { btn.textContent = "Copy room link"; }, 2000);
  });
}

function hashString(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  return hash;
}

function pickColor(name) {
  const colors = [
    "#1db954", "#e91e63", "#9c27b0", "#03a9f4",
    "#ff9800", "#8bc34a", "#f44336", "#00bcd4"
  ];
  const hash = hashString(name);
  return colors[Math.abs(hash) % colors.length];
}

// --- Apply style to the dice selector itself ---
function styleDiceSelector(select) {
  const dice = select.value;
  if (diceColors[dice]) {
    select.style.backgroundColor = diceColors[dice];
    select.style.color = "#121212"; // black text
  } else {
    select.style.backgroundColor = "#222"; // fallback
    select.style.color = "#fff"; // white text
  }
}

const diceColors = {
  d4: "#f44336",   // red
  d6: "#ff9800",   // orange
  d8: "#ffeb3b",   // yellow
  d10: "#03a9f4",  // blue
  d12: "#9c27b0",  // purple
  d100: "#77ff01",  // lime
  d20: "#1db954",  // spotify green
};

function applyLogColors() {
  document.querySelectorAll(".username").forEach(el => {
    el.style.color = pickColor(el.dataset.name);
  });

  document.querySelectorAll(".dice").forEach(el => {
    
    const dice = el.dataset.dice;
    if (diceColors[dice]) {
         el.style.backgroundColor = diceColors[dice];
         el.style.color = "#121212";
    } else {
        el.style.backgroundColor = "#ffffff";
    }
  });
  // 🎲 Handle the dice selector
  const diceSelect = document.getElementById("dice");
  if (diceSelect) {
    const dice = diceSelect.value;
    if (diceColors[dice]) {
      diceSelect.style.borderColor = diceColors[dice]; // ✅ match border
      diceSelect.style.backgroundColor = diceColors[dice]; // background
    } else {
      diceSelect.style.backgroundColor = "#222"; // fallback dark bg
       diceSelect.style.borderColor = "#444"; // subtle fallback border
      
    }
  }
}
// --- Append a new entry to the log ---
function appendLogEntry(m, logList) {
  const li = document.createElement("li");
  li.className = "log-entry new";

  const descDiv = document.createElement("div");
  descDiv.className = "desc";
  if (m.desc) {
    const descSpan = document.createElement("span");
    descSpan.className = "desc-text";
    descSpan.textContent = m.desc;
    const hr = document.createElement("hr");
    hr.className = "desc-separator";
    descDiv.appendChild(descSpan);
    descDiv.appendChild(hr);
  }

  const diceSpan = document.createElement("span");
  diceSpan.className = "dice";
  diceSpan.dataset.dice = m.dice;
  diceSpan.textContent = m.dice;

  const userSpan = document.createElement("span");
  userSpan.className = "username";
  userSpan.dataset.name = m.user;
  userSpan.textContent = m.user;

  const metaSpan = document.createElement("span");
  metaSpan.className = "meta";
  metaSpan.textContent = "rolled";

  const resultSpan = document.createElement("span");
  resultSpan.className = "result";
  resultSpan.textContent = m.result;

  const timeSpan = document.createElement("span");
  timeSpan.className = "time";
  timeSpan.textContent = m.time;

  li.appendChild(descDiv);
  li.appendChild(diceSpan);
  li.appendChild(userSpan);
  li.appendChild(metaSpan);
  li.appendChild(resultSpan);
  li.appendChild(timeSpan);

  logList.insertBefore(li, logList.firstChild);
  // Force a reflow so the browser registers the starting state
    li.offsetHeight; // ⚡ forces reflow
    li.classList.add("animate-in");
// trigger animation in next tick
  

  li.addEventListener("transitionend", () => {
    li.classList.remove("new", "animate-in");
  });

    applyLogColors();
 // logList.scrollTop = logList.scrollHeight;
}

// --- Initialize on page load ---
document.addEventListener("DOMContentLoaded", () => {
    const logList = document.getElementById("log");
    const diceSelect = document.getElementById("dice");
      // Style selector immediately on load
    if (diceSelect) {
        // Restore saved preference (localStorage stays in the browser, never sent to server)
        const saved = localStorage.getItem("selectedDice");
        if (saved) {
            diceSelect.value = saved;
        }
        styleDiceSelector(diceSelect);

        diceSelect.addEventListener("change", () => {
            localStorage.setItem("selectedDice", diceSelect.value);
            styleDiceSelector(diceSelect);
        });
    }

    if (logList) {
  // Style server-rendered entries immediately
        applyLogColors();

  // Connect to SSE
        const evtSource = new EventSource(HOST_PREFIX + `/events/${ROOM_ID}`);
        evtSource.onmessage = (event) => {
            try {
                const m = JSON.parse(event.data);
                appendLogEntry(m, logList);
            } catch (err) {
            console.error("Invalid SSE payload", event.data, err);
            }
        };
    }
});