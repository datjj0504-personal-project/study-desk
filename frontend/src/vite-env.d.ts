/// <reference types="vite/client" />

interface Window {
  // Injected by the Wails runtime; absent in a plain browser.
  go?: { main?: { App?: unknown } };
}
