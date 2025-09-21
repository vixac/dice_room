// Minimal JS (optional: could use htmx or websockets later)
// Right now, it's just a placeholder.
console.log("Dice app JS loaded");

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

const diceColors = {
  d4: "#f44336",   // red
  d6: "#ff9800",   // orange
  d8: "#ffeb3b",   // yellow
  d10: "#03a9f4",  // blue
  d12: "#9c27b0",  // purple
  d20: "#1db954",  // spotify green
};

function applyLogColors() {
  document.querySelectorAll(".username").forEach(el => {
    el.style.color = pickColor(el.dataset.name);
  });

  document.querySelectorAll(".dice").forEach(el => {
    
    const dice = el.dataset.dice;
    console.log("Looking at dice " + dice);
    if (diceColors[dice]) {
        console.log("Modifying style for " + dice);
         el.style.backgroundColor = diceColors[dice];
         el.style.color = "#121212";
    } else {
        console.log("Missing style for " + dice);
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

document.addEventListener("DOMContentLoaded", () => {
  const diceSelect = document.getElementById("dice");
  if (diceSelect) {
    diceSelect.addEventListener("change", applyLogColors);
    applyLogColors(); // initial
  }
});