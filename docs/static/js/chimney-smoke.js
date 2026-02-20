// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

(() => {
  if (!document.getElementById('smoke-bed')) return;
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

  const glyphs = ['\u2591', '\u2591', '\u2591', '\u2592'];
  const particles = [];
  const MAX_PARTICLES = 5;

  function rand(lo, hi) { return lo + Math.random() * (hi - lo); }

  function spawn() {
    const bed = document.getElementById('smoke-bed');
    if (bed && particles.length < MAX_PARTICLES) {
      const el = document.createElement('span');
      el.className = 'smoke-particle';
      el.textContent = glyphs[Math.floor(Math.random() * glyphs.length)];
      bed.appendChild(el);

      particles.push({
        el: el,
        age: 0,
        lifetime: rand(3000, 5000),
        x0: rand(-0.2, 0.2),
        amp1: rand(0.15, 0.4),
        freq1: rand(0.8, 1.6),
        phase1: rand(0, Math.PI * 2),
        amp2: rand(0.05, 0.12),
        freq2: rand(2.5, 4.0),
        phase2: rand(0, Math.PI * 2),
        rise: rand(0.8, 1.4),
        drift: rand(-0.05, 0.3),
        scaleEnd: rand(1.2, 1.6),
        peakAlpha: rand(0.7, 0.9)
      });
    }

    setTimeout(spawn, rand(800, 1500));
  }

  let last = 0;
  function tick(now) {
    if (!last) last = now;
    const dt = Math.min((now - last) / 1000, 0.1);
    last = now;

    for (let i = particles.length - 1; i >= 0; i--) {
      const p = particles[i];
      p.age += dt * 1000;
      const t = p.age / p.lifetime;

      if (t >= 1) {
        p.el.remove();
        particles.splice(i, 1);
        continue;
      }

      const y = p.rise * (p.age / 1000);
      const wobble1 = Math.sin(p.freq1 * t * Math.PI * 2 + p.phase1) * p.amp1;
      const wobble2 = Math.sin(p.freq2 * t * Math.PI * 2 + p.phase2) * p.amp2;
      const x = p.x0 + (wobble1 + wobble2) * t + p.drift * t;
      const scale = 1 + (p.scaleEnd - 1) * t;
      let alpha;
      if (t < 0.12) {
        alpha = (t / 0.12) * p.peakAlpha;
      } else {
        alpha = p.peakAlpha * (1 - (t - 0.12) / 0.88);
      }

      p.el.style.transform =
        `translate(${x.toFixed(2)}em, -${y.toFixed(2)}em) scale(${scale.toFixed(2)})`;
      p.el.style.opacity = alpha.toFixed(3);
    }

    requestAnimationFrame(tick);
  }

  spawn();
  requestAnimationFrame(tick);
})();
