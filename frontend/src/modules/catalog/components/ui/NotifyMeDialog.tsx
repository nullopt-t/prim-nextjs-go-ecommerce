"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Bell, X } from "lucide-react";
import {
	EMAIL_RE,
	getStockAlert,
	subscribeStockAlert,
} from "@/modules/catalog/lib/notifyMeStore";
import { notify } from "@/shared/lib/notify";

interface NotifyMeDialogProps {
	isOpen: boolean;
	slug?: string | null;
	productName?: string;
	onClose: () => void;
}

/**
 * Notify-me-when-available consent dialog for out-of-stock PDP variants.
 *
 * Frontend coordination stub: subscriptions are recorded locally (per product
 * slug) until the backend subscribe endpoint + Kafka consumer land. The store
 * helper is the only place to swap when wiring the real endpoint — keep this
 * dialog's copy and validation untouched.
 */
export function NotifyMeDialog({ isOpen, slug, productName, onClose }: NotifyMeDialogProps) {
	const t = useTranslations("products");
	const [email, setEmail] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [subscribed, setSubscribed] = useState(false);

	useEffect(() => {
		if (!isOpen) return;
		setEmail("");
		setError(null);
		setSubscribed(Boolean(getStockAlert(slug)));
	}, [isOpen, slug]);

	if (!isOpen) return null;

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!slug) return;
		if (!EMAIL_RE.test(email.trim())) {
			setError(t("notifyMe.emailInvalid"));
			return;
		}
		const result = subscribeStockAlert(slug, email.trim());
		if (result.ok) {
			setSubscribed(true);
			setError(null);
			notify.success(t("notifyMe.success", { name: productName || slug }));
			return;
		}
		if (result.reason === "cooldown") setError(t("notifyMe.cooldown"));
		else if (result.reason === "already-subscribed") setSubscribed(true);
		else setError(t("notifyMe.emailInvalid"));
	};

	return (
		<div
			className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150"
			onClick={onClose}
		>
			<div
				role="dialog"
				aria-modal="true"
				aria-label={t("notifyMe.title")}
				className="w-full max-w-md rounded-2xl border border-border bg-card shadow-2xl"
				onClick={(e) => e.stopPropagation()}
			>
				<div className="flex items-start justify-between gap-3 p-6 pb-0">
					<div className="flex items-center gap-3">
						<span className="flex size-10 shrink-0 items-center justify-center rounded-xl border border-border bg-secondary text-foreground">
							<Bell className="size-5" aria-hidden="true" />
						</span>
						<div>
							<h2 className="text-base font-bold text-foreground">{t("notifyMe.title")}</h2>
							<p className="mt-0.5 text-sm text-muted-foreground">
								{t("notifyMe.description", { name: productName || "" })}
							</p>
						</div>
					</div>
					<button
						type="button"
						onClick={onClose}
						aria-label={t("notifyMe.close")}
						className="shrink-0 rounded-lg p-1 text-muted-foreground hover:bg-secondary hover:text-foreground"
					>
						<X className="size-5" />
					</button>
				</div>

				{subscribed ? (
					<div className="p-6 pt-4">
						<p className="rounded-xl bg-emerald-600/10 px-4 py-3 text-sm font-medium text-emerald-700 dark:text-emerald-400">
							{t("notifyMe.subscribed")}
						</p>
						<button
							type="button"
							onClick={onClose}
							className="mt-4 w-full rounded-lg bg-accent-brand py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90"
						>
							{t("notifyMe.done")}
						</button>
					</div>
				) : (
					<form onSubmit={handleSubmit} className="flex flex-col gap-3 p-6 pt-4" noValidate>
						<label htmlFor="notify-me-email" className="text-sm font-semibold text-foreground">
							{t("notifyMe.emailLabel")}
						</label>
						<input
							id="notify-me-email"
							type="email"
							autoComplete="email"
							value={email}
							onChange={(e) => setEmail(e.target.value)}
							placeholder="customer@example.com"
							className="rounded-lg border border-border bg-background px-3.5 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-accent-brand focus:outline-none"
						/>
						{error && (
							<p role="alert" className="text-xs text-destructive">
								{error}
							</p>
						)}
						<div className="flex gap-3">
							<button
								type="button"
								onClick={onClose}
								className="flex-1 rounded-lg border border-border bg-background px-4 py-2.5 text-xs font-medium text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground"
							>
								{t("notifyMe.cancel")}
							</button>
							<button
								type="submit"
								className="flex-1 rounded-lg bg-accent-brand px-4 py-2.5 text-xs font-semibold text-white transition-opacity hover:opacity-90"
							>
								{t("notifyMe.submit")}
							</button>
						</div>
					</form>
				)}
			</div>
		</div>
	);
}
