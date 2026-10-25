"use client";

import { AlertTriangle, Loader2, X } from "lucide-react";
import { useTranslations } from "next-intl";

interface ConfirmDialogProps {
  isOpen: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Visual intent of the confirm action. Defaults to a destructive (red) action. */
  variant?: "danger" | "default";
  /** Disables both actions and shows a spinner on the confirm button. */
  loading?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

/**
 * In-app confirmation dialog to replace native `window.confirm`.
 *
 * Keeps destructive prompts inside the design system (themed, localized,
 * keyboard/focus friendly) instead of the browser's unstyled modal.
 */
export function ConfirmDialog({
  isOpen,
  title,
  description,
  confirmLabel,
  cancelLabel,
  variant = "danger",
  loading = false,
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  const t = useTranslations("common");

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        role="alertdialog"
        aria-modal="true"
        aria-label={title}
        className="bg-card border border-border w-full max-w-sm rounded-2xl shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6 space-y-5">
          <div className="flex items-start gap-3">
            <div
              className={`p-2.5 rounded-xl border shrink-0 ${
                variant === "danger"
                  ? "bg-destructive/10 text-destructive border-destructive/20"
                  : "bg-secondary text-foreground border-border"
              }`}
            >
              <AlertTriangle className="size-5" />
            </div>
            <div className="space-y-1 flex-1 min-w-0">
              <h2 className="text-sm font-bold text-foreground">{title}</h2>
              {description && (
                <p className="text-xs text-muted-foreground leading-relaxed">{description}</p>
              )}
            </div>
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors disabled:opacity-50"
            >
              <X className="size-4" />
            </button>
          </div>

          <div className="flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2.5 rounded-xl border border-border text-xs font-semibold text-foreground hover:bg-secondary transition-colors disabled:opacity-50"
            >
              {cancelLabel ?? t("cancel")}
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={loading}
              className={`px-5 py-2.5 rounded-xl text-xs font-bold transition-colors flex items-center gap-2 shadow-xs disabled:opacity-60 ${
                variant === "danger"
                  ? "bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  : "bg-accent-brand text-white hover:bg-accent-brand/90"
              }`}
            >
              {loading && <Loader2 className="size-3.5 animate-spin" />}
              <span>{confirmLabel ?? t("confirm")}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
