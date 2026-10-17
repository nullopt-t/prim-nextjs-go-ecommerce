"use client";

import { useEffect, useRef, useState } from "react";
import { Copy, Download, Loader2, Share2, X } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { notify } from "@/lib/notify";

const CARD_W = 1200;
const CARD_H = 630;
const SCALE = 2;

export interface PromotionShareContent {
  code: string;
  /** Formatted discount magnitude, e.g. "20%" or "EGP 50.00". */
  value: string;
  /** Localized discount type, e.g. "Percentage". */
  typeLabel: string;
  /** Localized validity line, e.g. "Valid until Oct 9, 2026, 3:45 PM". */
  validityLabel: string;
}

interface PromotionShareDialogProps {
  isOpen: boolean;
  onClose: () => void;
  content: PromotionShareContent | null;
}

function hexToRgb(hex: string): [number, number, number] {
  const h = hex.replace("#", "").trim();
  const full = h.length === 3 ? h.split("").map((c) => c + c).join("") : h;
  const n = parseInt(full, 16);
  if (full.length !== 6 || Number.isNaN(n)) return [217, 119, 87];
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

function roundedRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number,
) {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}

/** Shrinks the font size until the text fits within maxWidth. */
function fitFont(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  startSize: number,
  weight: number,
  family: string,
): void {
  let size = startSize;
  ctx.font = `${weight} ${size}px ${family}`;
  while (size > 36 && ctx.measureText(text).width > maxWidth) {
    size -= 4;
    ctx.font = `${weight} ${size}px ${family}`;
  }
}

function drawCard(
  canvas: HTMLCanvasElement,
  content: PromotionShareContent,
  labels: { promoCode: string; cta: string },
  dir: "ltr" | "rtl",
) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  canvas.width = CARD_W * SCALE;
  canvas.height = CARD_H * SCALE;
  canvas.setAttribute("dir", dir);
  ctx.setTransform(SCALE, 0, 0, SCALE, 0, 0);
  ctx.clearRect(0, 0, CARD_W, CARD_H);

  // Follow the active locale's text direction (RTL for Arabic) so mixed
  // Arabic/Latin lines (e.g. "Valid until <date>") lay out correctly.
  ctx.direction = dir;

  const styles = getComputedStyle(document.documentElement);
  const accentHex = (styles.getPropertyValue("--accent-brand") || "#d97757").trim() || "#d97757";
  const [ar, ag, ab] = hexToRgb(accentHex);
  const accent = `rgb(${ar}, ${ag}, ${ab})`;
  const accentSoft = `rgba(${ar}, ${ag}, ${ab}, 0.12)`;
  const accentFaint = `rgba(${ar}, ${ag}, ${ab}, 0.07)`;
  const ink = "#1b1b1f";
  const muted = "#717182";
  const bodyFont = getComputedStyle(document.body).fontFamily || "system-ui, sans-serif";
  const monoFont = "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace";

  // Background gradient
  const bg = ctx.createLinearGradient(0, 0, CARD_W, CARD_H);
  bg.addColorStop(0, "#ffffff");
  bg.addColorStop(1, "#fff6f2");
  ctx.fillStyle = bg;
  ctx.fillRect(0, 0, CARD_W, CARD_H);

  // Decorative accents
  ctx.fillStyle = accentFaint;
  ctx.beginPath();
  ctx.arc(CARD_W - 90, 40, 240, 0, Math.PI * 2);
  ctx.fill();
  ctx.beginPath();
  ctx.arc(40, CARD_H + 20, 180, 0, Math.PI * 2);
  ctx.fill();

  ctx.textBaseline = "alphabetic";

  // Brand (PRI + accent M), centered
  ctx.font = `900 40px ${bodyFont}`;
  const brandParts: Array<[string, string]> = [["PRI", ink], ["M", accent]];
  const brandWidths = brandParts.map(([txt]) => ctx.measureText(txt).width);
  let bx = CARD_W / 2 - brandWidths.reduce((a, b) => a + b, 0) / 2;
  ctx.textAlign = "left";
  brandParts.forEach(([txt, color], i) => {
    ctx.fillStyle = color;
    ctx.fillText(txt, bx, 78);
    bx += brandWidths[i];
  });

  ctx.textAlign = "center";

  // Discount type pill
  const pillText = (content.typeLabel || "").toUpperCase();
  ctx.font = `700 22px ${bodyFont}`;
  const pillW = ctx.measureText(pillText).width + 60;
  const pillH = 52;
  const pillY = 116;
  ctx.fillStyle = accentSoft;
  roundedRect(ctx, CARD_W / 2 - pillW / 2, pillY, pillW, pillH, pillH / 2);
  ctx.fill();
  ctx.fillStyle = accent;
  ctx.fillText(pillText, CARD_W / 2, pillY + pillH / 2 + 8);

  // Big discount value
  fitFont(ctx, content.value || "", CARD_W - 180, 150, 900, bodyFont);
  ctx.fillStyle = accent;
  ctx.fillText(content.value || "", CARD_W / 2, 342);

  // "PROMO CODE" label (letter spacing is a Latin-typography device — skip it for RTL)
  ctx.font = `700 22px ${bodyFont}`;
  ctx.fillStyle = muted;
  if (dir === "ltr") {
    (ctx as CanvasRenderingContext2D & { letterSpacing?: string }).letterSpacing = "6px";
  }
  ctx.fillText(labels.promoCode, CARD_W / 2, 404);
  (ctx as CanvasRenderingContext2D & { letterSpacing?: string }).letterSpacing = "0px";

  // Code pill (dashed)
  fitFont(ctx, content.code || "", CARD_W - 220, 56, 800, monoFont);
  const codeW = ctx.measureText(content.code || "").width + 96;
  const codeH = 92;
  const codeX = CARD_W / 2 - codeW / 2;
  const codeY = 428;
  ctx.save();
  ctx.setLineDash([12, 10]);
  ctx.lineWidth = 3;
  ctx.strokeStyle = accent;
  ctx.fillStyle = accentSoft;
  roundedRect(ctx, codeX, codeY, codeW, codeH, 20);
  ctx.fill();
  ctx.stroke();
  ctx.restore();
  ctx.fillStyle = ink;
  ctx.fillText(content.code || "", CARD_W / 2, codeY + codeH / 2 + 20);

  // Validity + call to action
  ctx.font = `500 24px ${bodyFont}`;
  ctx.fillStyle = muted;
  ctx.fillText(content.validityLabel || "", CARD_W / 2, 578);
  ctx.font = `700 22px ${bodyFont}`;
  ctx.fillStyle = ink;
  ctx.fillText(labels.cta, CARD_W / 2, 612);
}

/**
 * Renders a promotion as a shareable image card and copies it to the
 * clipboard (PNG), with a download fallback for browsers that don't support
 * writing images to the clipboard.
 */
export function PromotionShareDialog({ isOpen, onClose, content }: PromotionShareDialogProps) {
  const t = useTranslations("admin");
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [busy, setBusy] = useState(false);
  const locale = useLocale();
  const dir: "ltr" | "rtl" = locale === "ar" ? "rtl" : "ltr";

  useEffect(() => {
    if (!isOpen || !content) return;
    let cancelled = false;

    const render = async () => {
      try {
        await document.fonts?.ready;
      } catch {
        // Fonts API unavailable — draw with whatever is loaded.
      }
      if (cancelled || !canvasRef.current) return;
      drawCard(
        canvasRef.current,
        content,
        {
          promoCode: t("promotions.share.promoCode"),
          cta: t("promotions.share.cta"),
        },
        dir,
      );
    };

    render();
    return () => {
      cancelled = true;
    };
  }, [isOpen, content, t, dir]);

  const toBlob = () =>
    new Promise<Blob | null>((resolve) => {
      const canvas = canvasRef.current;
      if (!canvas) return resolve(null);
      canvas.toBlob(resolve, "image/png");
    });

  const handleCopy = async () => {
    setBusy(true);
    try {
      const blob = await toBlob();
      const ClipboardItemCtor = (window as unknown as { ClipboardItem?: typeof ClipboardItem })
        .ClipboardItem;
      if (!blob || !navigator.clipboard || !ClipboardItemCtor) throw new Error("unsupported");
      await navigator.clipboard.write([new ClipboardItemCtor({ "image/png": blob })]);
      notify.success(t("promotions.share.copied"));
    } catch {
      notify.error(t("promotions.share.copyFailed"));
    } finally {
      setBusy(false);
    }
  };

  const handleDownload = async () => {
    const blob = await toBlob();
    if (!blob) return;
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `promo-${content?.code || "card"}.png`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    notify.info(t("promotions.share.downloaded"));
  };

  if (!isOpen || !content) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-background border border-border rounded-2xl w-full max-w-2xl p-6 shadow-xl space-y-4">
        <div className="flex justify-between items-center border-b border-border pb-3">
          <h3 className="font-bold text-base text-foreground flex items-center gap-2">
            <Share2 className="size-4 text-accent-brand" />
            {t("promotions.share.title")}
          </h3>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors">
            <X className="size-5" />
          </button>
        </div>

        <canvas
          ref={canvasRef}
          dir={dir}
          aria-label={t("promotions.share.title")}
          className="w-full h-auto rounded-xl border border-border shadow-xs"
        />

        <p className="text-xs text-muted-foreground">{t("promotions.share.hint")}</p>

        <div className="flex flex-wrap justify-end gap-3 pt-1">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-xs border border-border rounded-xl hover:bg-secondary transition-colors"
          >
            {t("actions.cancel")}
          </button>
          <button
            type="button"
            onClick={handleDownload}
            className="inline-flex items-center gap-2 px-4 py-2 text-xs border border-border rounded-xl hover:bg-secondary transition-colors"
          >
            <Download className="size-3.5" />
            {t("promotions.share.download")}
          </button>
          <button
            type="button"
            onClick={handleCopy}
            disabled={busy}
            className="inline-flex items-center gap-2 px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl hover:bg-accent-brand/90 transition-colors disabled:opacity-50"
          >
            {busy ? <Loader2 className="size-3.5 animate-spin" /> : <Copy className="size-3.5" />}
            {t("promotions.share.copyImage")}
          </button>
        </div>
      </div>
    </div>
  );
}
