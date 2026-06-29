import { mountAuthCard, type AuthMode } from "./auth-ui";
import type { RubricalSession } from "./auth-api";

export const AUTH_MODAL_ID = "rubrical-auth-modal";

function onEscape(event: KeyboardEvent): void {
  if (event.key === "Escape") {
    closeAuthModal();
  }
}

export function closeAuthModal(): void {
  const modal = document.getElementById(AUTH_MODAL_ID);
  if (!modal) {
    return;
  }
  modal.remove();
  document.documentElement.style.overflow = "";
  document.removeEventListener("keydown", onEscape);
}

export type AuthModalOptions = {
  initialMode?: AuthMode;
  onSignedIn?: (session: RubricalSession) => void;
};

export function showAuthModal(options: AuthModalOptions = {}): void {
  closeAuthModal();

  const overlay = document.createElement("div");
  overlay.id = AUTH_MODAL_ID;
  overlay.setAttribute("role", "dialog");
  overlay.setAttribute("aria-modal", "true");
  overlay.setAttribute("aria-label", "Sign in to Rubrical");
  overlay.style.cssText = [
    "position:fixed",
    "inset:0",
    "z-index:2147483647",
    "display:flex",
    "flex-direction:column",
    "align-items:center",
    "justify-content:safe center",
    "padding:16px",
    "background:rgba(15,23,42,0.55)",
    "backdrop-filter:blur(2px)",
    "overflow-y:auto",
    "overscroll-behavior:contain",
    "scrollbar-width:none",
    "-ms-overflow-style:none",
  ].join(";");

  const scrollStyle = document.createElement("style");
  scrollStyle.textContent = `
    #${AUTH_MODAL_ID}::-webkit-scrollbar {
      display: none;
    }
  `;
  overlay.append(scrollStyle);

  const authRoot = document.createElement("div");
  authRoot.style.cssText = "width:min(400px,100%);flex-shrink:0;";

  void mountAuthCard(authRoot, {
    initialMode: options.initialMode,
    onClose: closeAuthModal,
    onSignedIn: (session) => {
      closeAuthModal();
      options.onSignedIn?.(session);
    },
  });

  overlay.append(authRoot);

  overlay.addEventListener("click", (event) => {
    if (event.target === overlay) {
      closeAuthModal();
    }
  });

  document.body.append(overlay);
  document.documentElement.style.overflow = "hidden";
  document.addEventListener("keydown", onEscape);
}
