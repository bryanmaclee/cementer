// Client theme preference (dark default / light). Stored per-laptop in
// localStorage — a personal preference, not synced. Applied as data-theme on
// <html>; the CSS defines both palettes.

export type Theme = "dark" | "light";

const KEY = "cementer.theme";

function read(): Theme {
  return localStorage.getItem(KEY) === "light" ? "light" : "dark";
}

function apply(theme: Theme): void {
  document.documentElement.dataset.theme = theme;
}

export function initTheme(): Theme {
  const t = read();
  apply(t);
  return t;
}

export function currentTheme(): Theme {
  return read();
}

export function toggleTheme(): Theme {
  const next: Theme = read() === "dark" ? "light" : "dark";
  localStorage.setItem(KEY, next);
  apply(next);
  return next;
}
