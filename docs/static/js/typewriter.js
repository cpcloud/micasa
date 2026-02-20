// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

(() => {
  const el = document.getElementById('typewriter');
  const words = [
    { text: 'panicked',    cls: 'accent' },
    { text: 'daydreamed',  cls: 'accent-alt' },
    { text: 'avoided',        cls: 'accent' },
    { text: 'workshopped',  cls: 'accent-alt' },
    { text: 'entertained', cls: 'accent-alt' }
  ];
  let i = 0;
  const TYPE = 100, DELETE = 60, PAUSE = 2800, GAP = 500;

  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (reduced) return;

  function type(word, cb) {
    let j = 0;
    el.className = `typewriter ${word.cls}`;
    (function next() {
      el.textContent = word.text.slice(0, ++j);
      if (j < word.text.length) setTimeout(next, TYPE);
      else setTimeout(cb, PAUSE);
    })();
  }

  function erase(cb) {
    let text = el.textContent;
    (function next() {
      text = text.slice(0, -1);
      el.textContent = text;
      if (text.length > 0) setTimeout(next, DELETE);
      else setTimeout(cb, GAP);
    })();
  }

  function cycle() {
    type(words[i], () => {
      erase(() => {
        i = (i + 1) % words.length;
        cycle();
      });
    });
  }

  setTimeout(() => {
    erase(() => {
      i = 1;
      cycle();
    });
  }, PAUSE);
})();
