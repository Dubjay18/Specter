// ---------- Scroll reveal ----------
const revealEls = document.querySelectorAll(".reveal");
const revealObserver = new IntersectionObserver(
  (entries) => {
    entries.forEach((e) => {
      if (e.isIntersecting) {
        e.target.classList.add("in-view");
        revealObserver.unobserve(e.target);
      }
    });
  },
  { threshold: 0.15 }
);
revealEls.forEach((el) => revealObserver.observe(el));

// ---------- Hero terminal typing ----------
const termLines = [
  { text: "$ curl -H \"X-User-ID: user-123\" specter.local/profile", cls: "c-cmd" },
  { text: "→ mirrored to shadow:8081  (async, discarded)", cls: "c-comment" },
  { text: "→ live:8080 responded  200 OK  in 14ms", cls: "" },
  { text: "→ shadow:8081 responded  200 OK  in 19ms", cls: "" },
  { text: "⚡ divergence: field \"total\" 42.10 != 41.85", cls: "c-str" },
  { text: "✓ logged to divergence store", cls: "c-cmd" },
];

async function typeTerminal() {
  const el = document.getElementById("hero-term");
  if (!el) return;
  el.innerHTML = "";
  for (const line of termLines) {
    const lineEl = document.createElement("div");
    const span = document.createElement("span");
    if (line.cls) span.className = line.cls;
    lineEl.appendChild(span);
    el.appendChild(lineEl);
    for (let i = 0; i < line.text.length; i++) {
      span.textContent += line.text[i];
      await sleep(8 + Math.random() * 14);
    }
    await sleep(220);
  }
  const cursor = document.createElement("span");
  cursor.className = "cursor";
  el.appendChild(cursor);
}
function sleep(ms) { return new Promise((r) => setTimeout(r, ms)); }

let typed = false;
const termObserver = new IntersectionObserver((entries) => {
  entries.forEach((e) => {
    if (e.isIntersecting && !typed) {
      typed = true;
      typeTerminal();
      termObserver.disconnect();
    }
  });
});
const heroTerm = document.getElementById("hero-term");
if (heroTerm) termObserver.observe(heroTerm);

// ---------- Diagram: packet run + divergence pulse ----------
const diagram = document.getElementById("diagram");
if (diagram) {
  const packet = document.getElementById("packet");
  const divergence = document.getElementById("dg-divergence");
  const diagramObserver = new IntersectionObserver((entries) => {
    entries.forEach((e) => {
      if (e.isIntersecting) {
        packet.classList.add("run");
        setTimeout(() => divergence.classList.add("show"), 900);
      } else {
        packet.classList.remove("run");
        divergence.classList.remove("show");
      }
    });
  }, { threshold: 0.4 });
  diagramObserver.observe(diagram);
}

// ---------- Hash ring canvas ----------
const ringCanvas = document.getElementById("ring-canvas");
if (ringCanvas) {
  const ctx = ringCanvas.getContext("2d");
  const size = 560;
  const dpr = Math.min(window.devicePixelRatio || 1, 2);
  ringCanvas.width = size * dpr;
  ringCanvas.height = size * dpr;
  ctx.scale(dpr, dpr);

  const cx = size / 2;
  const cy = size / 2;
  const radius = size / 2 - 60;

  const nodeCount = 5;
  const nodeColors = ["#2dd4bf", "#a78bfa", "#fbbf24", "#fb7185", "#60a5fa"];
  const nodes = [];
  for (let n = 0; n < nodeCount; n++) {
    const virtuals = 4 + Math.floor(Math.random() * 3);
    for (let v = 0; v < virtuals; v++) {
      nodes.push({
        angle: Math.random() * Math.PI * 2,
        color: nodeColors[n],
        nodeId: n,
      });
    }
  }

  let keyAngle = 0;
  let running = true;
  const ringObserver = new IntersectionObserver((entries) => {
    entries.forEach((e) => { running = e.isIntersecting; });
  });
  ringObserver.observe(ringCanvas);

  function nearestClockwise(angle) {
    let best = null;
    let bestDelta = Infinity;
    for (const n of nodes) {
      let delta = n.angle - angle;
      if (delta < 0) delta += Math.PI * 2;
      if (delta < bestDelta) { bestDelta = delta; best = n; }
    }
    return best;
  }

  function draw() {
    ctx.clearRect(0, 0, size, size);

    // outer ring
    ctx.beginPath();
    ctx.arc(cx, cy, radius, 0, Math.PI * 2);
    ctx.strokeStyle = "rgba(230,237,243,0.12)";
    ctx.lineWidth = 1.5;
    ctx.stroke();

    // nodes
    nodes.forEach((n) => {
      const x = cx + Math.cos(n.angle) * radius;
      const y = cy + Math.sin(n.angle) * radius;
      ctx.beginPath();
      ctx.arc(x, y, 6, 0, Math.PI * 2);
      ctx.fillStyle = n.color;
      ctx.shadowColor = n.color;
      ctx.shadowBlur = 12;
      ctx.fill();
      ctx.shadowBlur = 0;
    });

    // rotating key
    const kx = cx + Math.cos(keyAngle) * radius;
    const ky = cy + Math.sin(keyAngle) * radius;
    const owner = nearestClockwise(keyAngle);

    // line from key to owner
    if (owner) {
      const ox = cx + Math.cos(owner.angle) * radius;
      const oy = cy + Math.sin(owner.angle) * radius;
      ctx.beginPath();
      ctx.moveTo(kx, ky);
      ctx.arc(cx, cy, radius, keyAngle, owner.angle);
      ctx.strokeStyle = "rgba(167,139,250,0.5)";
      ctx.lineWidth = 2;
      ctx.stroke();

      ctx.beginPath();
      ctx.arc(ox, oy, 10, 0, Math.PI * 2);
      ctx.strokeStyle = owner.color;
      ctx.lineWidth = 2;
      ctx.stroke();
    }

    ctx.beginPath();
    ctx.arc(kx, ky, 6, 0, Math.PI * 2);
    ctx.fillStyle = "#a78bfa";
    ctx.shadowColor = "#a78bfa";
    ctx.shadowBlur = 14;
    ctx.fill();
    ctx.shadowBlur = 0;

    // center label
    ctx.fillStyle = "rgba(230,237,243,0.35)";
    ctx.font = "12px monospace";
    ctx.textAlign = "center";
    ctx.fillText("consistent hash ring", cx, cy);
  }

  function tick() {
    if (running) {
      keyAngle += 0.004;
      if (keyAngle > Math.PI * 2) keyAngle -= Math.PI * 2;
      draw();
    }
    requestAnimationFrame(tick);
  }
  draw();
  tick();
}

// ---------- Background particle field ----------
const bgCanvas = document.getElementById("bg-canvas");
if (bgCanvas) {
  const bctx = bgCanvas.getContext("2d");
  let w, h, particles;

  function resize() {
    w = bgCanvas.width = window.innerWidth;
    h = bgCanvas.height = window.innerHeight;
  }
  window.addEventListener("resize", resize);
  resize();

  const count = Math.min(70, Math.floor((window.innerWidth * window.innerHeight) / 18000));
  particles = Array.from({ length: count }, () => ({
    x: Math.random() * w,
    y: Math.random() * h,
    r: Math.random() * 1.6 + 0.4,
    vx: (Math.random() - 0.5) * 0.15,
    vy: (Math.random() - 0.5) * 0.15,
    hue: Math.random() > 0.5 ? "45,212,191" : "167,139,250",
  }));

  function drawBg() {
    bctx.clearRect(0, 0, w, h);
    particles.forEach((p) => {
      p.x += p.vx;
      p.y += p.vy;
      if (p.x < 0) p.x = w; if (p.x > w) p.x = 0;
      if (p.y < 0) p.y = h; if (p.y > h) p.y = 0;
      bctx.beginPath();
      bctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
      bctx.fillStyle = `rgba(${p.hue},0.5)`;
      bctx.fill();
    });
    requestAnimationFrame(drawBg);
  }
  drawBg();
}
