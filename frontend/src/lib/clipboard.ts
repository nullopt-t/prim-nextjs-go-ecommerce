/**
 * Clipboard helpers.
 *
 * `navigator.clipboard` is only exposed in a secure context (HTTPS or
 * `localhost`). When the app is served over plain HTTP on a LAN IP — the normal
 * local/dev setup here — it is `undefined`, so a direct `navigator.clipboard
 * .writeText(...)` throws `TypeError: can't access property "writeText"`.
 *
 * `copyTextToClipboard` therefore never throws: it prefers the async Clipboard
 * API when available, falls back to the legacy `execCommand("copy")` path, and
 * resolves to whether the copy actually succeeded so callers can decide what
 * toast to show.
 */
export async function copyTextToClipboard(text: string): Promise<boolean> {
  if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // The API can reject even when present (document not focused, permission
      // denied, insecure context). Fall through to the legacy path.
    }
  }

  return legacyCopy(text);
}

function legacyCopy(text: string): boolean {
  if (typeof document === "undefined") return false;

  let textarea: HTMLTextAreaElement | null = null;
  try {
    textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.setAttribute("readonly", "");
    // Keep it invisible and out of the layout/scroll flow while still
    // selectable, which execCommand("copy") requires.
    textarea.style.position = "fixed";
    textarea.style.top = "0";
    textarea.style.left = "0";
    textarea.style.width = "1px";
    textarea.style.height = "1px";
    textarea.style.padding = "0";
    textarea.style.border = "none";
    textarea.style.outline = "none";
    textarea.style.boxShadow = "none";
    textarea.style.background = "transparent";
    textarea.style.opacity = "0";

    document.body.appendChild(textarea);

    const selection = document.getSelection();
    const previousRange =
      selection && selection.rangeCount > 0 ? selection.getRangeAt(0) : null;

    textarea.select();
    textarea.setSelectionRange(0, text.length);

    const copied = document.execCommand("copy");

    if (previousRange && selection) {
      selection.removeAllRanges();
      selection.addRange(previousRange);
    }

    return copied;
  } catch {
    return false;
  } finally {
    if (textarea && textarea.parentNode) {
      textarea.parentNode.removeChild(textarea);
    }
  }
}
