import { refreshStagedFileIndicators } from "./staged-files";

export const MODAL_ID = "rubrical-modal";
export const MODAL_IFRAME_ID = "rubrical-modal-iframe";

function onEscape(event: KeyboardEvent): void {
  if (event.key === "Escape") {
    closeAssignmentModal();
  }
}

export function closeAssignmentModal(): void {
  const modal = document.getElementById(MODAL_ID);
  if (!modal) {
    return;
  }

  modal.remove();
  document.documentElement.style.overflow = "";
  document.removeEventListener("keydown", onEscape);
  void refreshStagedFileIndicators();
}

export function openAssignmentModal(base: string, path: string, title: string): void {
  closeAssignmentModal();

  const overlay = document.createElement("div");
  overlay.id = MODAL_ID;
  overlay.setAttribute("role", "dialog");
  overlay.setAttribute("aria-modal", "true");
  overlay.setAttribute("aria-label", "Rubrical assignment review");
  overlay.style.cssText = [
    "position:fixed",
    "inset:0",
    "z-index:2147483647",
    "display:flex",
    "align-items:center",
    "justify-content:center",
    "padding:16px",
    "background:rgba(15,23,42,0.55)",
    "backdrop-filter:blur(2px)",
  ].join(";");

  const panel = document.createElement("div");
  panel.style.cssText = [
    "display:flex",
    "flex-direction:column",
    "width:min(1100px,100%)",
    "height:min(90vh,900px)",
    "background:#fafaf9",
    "border-radius:12px",
    "overflow:hidden",
    "box-shadow:0 25px 50px rgba(0,0,0,0.25)",
  ].join(";");

  const header = document.createElement("div");
  header.style.cssText = [
    "display:flex",
    "align-items:center",
    "justify-content:space-between",
    "gap:12px",
    "padding:12px 16px",
    "border-bottom:1px solid #e7e5e4",
    "background:#fff",
  ].join(";");

  const heading = document.createElement("div");
  heading.style.cssText = "min-width:0";

  const brand = document.createElement("p");
  brand.textContent = "Rubrical";
  brand.style.cssText = "margin:0;font-size:14px;font-weight:600;color:#312e81";

  const subtitle = document.createElement("p");
  subtitle.textContent = title;
  subtitle.style.cssText =
    "margin:2px 0 0; font-size:13px; color:#57534e; white-space:nowrap; overflow:hidden; text-overflow:ellipsis";

  heading.append(brand, subtitle);

  const closeButton = document.createElement("button");
  closeButton.type = "button";
  closeButton.setAttribute("aria-label", "Close Rubrical");
  closeButton.textContent = "×";
  closeButton.style.cssText = [
    "flex-shrink:0",
    "width:36px",
    "height:36px",
    "border:none",
    "border-radius:8px",
    "background:#f5f5f4",
    "color:#44403c",
    "font-size:24px",
    "line-height:1",
    "cursor:pointer",
  ].join(";");
  closeButton.addEventListener("click", () => closeAssignmentModal());

  const iframe = document.createElement("iframe");
  iframe.id = MODAL_IFRAME_ID;
  iframe.title = `Rubrical: ${title}`;
  const embedUrl = new URL(path, base);
  embedUrl.searchParams.set("embed", "1");
  embedUrl.searchParams.set("_", String(Date.now()));
  iframe.src = embedUrl.toString();
  iframe.style.cssText = "flex:1;width:100%;border:none;background:#fafaf9";

  header.append(heading, closeButton);
  panel.append(header, iframe);
  overlay.append(panel);

  overlay.addEventListener("click", (event) => {
    if (event.target === overlay) {
      closeAssignmentModal();
    }
  });

  document.body.append(overlay);
  document.documentElement.style.overflow = "hidden";
  document.addEventListener("keydown", onEscape);
  closeButton.focus();
}
