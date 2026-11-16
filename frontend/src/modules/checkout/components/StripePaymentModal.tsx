"use client";

import { useState, type FormEvent } from "react";
import { loadStripe, type StripeElementLocale } from "@stripe/stripe-js";
import {
	Elements,
	PaymentElement,
	useStripe,
	useElements,
} from "@stripe/react-stripe-js";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
} from "@/shared/ui/form/dialog";
import {
	Lock,
	AlertCircle,
	Loader2,
	CreditCard,
	CheckCircle2,
	RefreshCw,
} from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { orderService } from "@/modules/orders/api/orderService";
import { apiErrorMessage } from "@/shared/lib/notify";

export const isStripeConfigured = Boolean(
	process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY,
);
const stripePromise = isStripeConfigured
	? loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || "")
	: null;

export interface PendingStripeOrder {
	clientSecret: string;
	orderId: string;
	totalFormatted: string;
	completedOrderInfo: {
		id: string;
		trackingNumber: string;
		total: string;
		shippingAddress: string;
		date: string;
	};
}

interface StripePaymentModalProps {
	clientSecret: string;
	orderId: string;
	totalFormatted: string;
	onClose: () => void;
	onSuccess: () => void;
	onPending: () => void;
}

function PaymentForm({
	orderId,
	totalFormatted,
	isProcessing,
	setIsProcessing,
	errorMessage,
	setErrorMessage,
	onConfirm,
	onPending,
}: {
	orderId: string;
	totalFormatted: string;
	isProcessing: boolean;
	setIsProcessing: (value: boolean) => void;
	errorMessage: string | null;
	setErrorMessage: (value: string | null) => void;
	onConfirm: (paymentIntentId: string) => Promise<void>;
	onPending: () => void;
}) {
	const stripe = useStripe();
	const elements = useElements();
	const t = useTranslations("checkout.payment");

	const handleSubmit = async (e: FormEvent) => {
		e.preventDefault();
		if (!stripe || !elements || isProcessing) return;

		setIsProcessing(true);
		setErrorMessage(null);

		// Keep the locale prefix so an Arabic shopper returns to /ar/checkout.
		const returnUrl = `${window.location.origin}${window.location.pathname}?order_id=${encodeURIComponent(orderId)}`;

		const { error, paymentIntent } = await stripe.confirmPayment({
			elements,
			confirmParams: {
				return_url: returnUrl,
			},
			redirect: "if_required",
		});

		if (error) {
			setErrorMessage(error.message || t("paymentFailed"));
			setIsProcessing(false);
			return;
		}

		// "processing" is not settled yet; the webhook finalizes the order.
		if (paymentIntent?.status === "processing") {
			onPending();
			return;
		}

		if (paymentIntent?.status !== "succeeded") {
			setErrorMessage(t("paymentFailed"));
			setIsProcessing(false);
			return;
		}

		await onConfirm(paymentIntent.id);
	};

	return (
		<form onSubmit={handleSubmit} className="space-y-5">
			<div className="rounded-xl border border-border/80 bg-secondary/30 p-4">
				<PaymentElement />
			</div>

			{errorMessage && (
				<div
					role="alert"
					className="rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-xs font-semibold text-destructive flex items-center gap-2"
				>
					<AlertCircle className="size-4 shrink-0" aria-hidden="true" />
					<span>{errorMessage}</span>
				</div>
			)}

			<button
				type="submit"
				disabled={!stripe || !elements || isProcessing}
				className="w-full py-3.5 px-6 rounded-xl bg-accent-brand text-white font-bold text-sm hover:bg-accent-brand/90 transition-all flex items-center justify-center gap-2 disabled:opacity-50 cursor-pointer shadow-md"
			>
				{isProcessing ? (
					<>
						<Loader2 className="size-4 animate-spin" />
						<span>{t("processingPayment")}</span>
					</>
				) : (
					<>
						<Lock className="size-4" />
						<span>{t("payNow", { amount: totalFormatted })}</span>
					</>
				)}
			</button>

			<p className="text-[11px] text-muted-foreground text-center flex items-center justify-center gap-1.5">
				<Lock className="size-3 shrink-0" />
				<span>{t("security")}</span>
			</p>
		</form>
	);
}

export function StripePaymentModal({
	clientSecret,
	orderId,
	totalFormatted,
	onClose,
	onSuccess,
	onPending,
}: StripePaymentModalProps) {
	const t = useTranslations("checkout.payment");
	const locale = useLocale() as StripeElementLocale;
	const [isProcessing, setIsProcessing] = useState(false);
	const [errorMessage, setErrorMessage] = useState<string | null>(null);
	const [unrecordedIntent, setUnrecordedIntent] = useState<string | null>(null);

	// Mock mode is strictly restricted to non-production environments when backend mints pi_mock_
	const isDevMock =
		process.env.NODE_ENV !== "production" && clientSecret.startsWith("pi_mock_");

	// The card was already authorized; only the order confirmation is retried,
	// so this can never charge the customer twice.
	const confirmOrder = async (paymentIntentId: string) => {
		try {
			await orderService.confirmPayment({ orderId, paymentIntentId });
			onSuccess();
		} catch (err) {
			setUnrecordedIntent(paymentIntentId);
			setErrorMessage(apiErrorMessage(err, t("confirmFailed")));
			setIsProcessing(false);
		}
	};

	const handleOpenChange = (open: boolean) => {
		if (open) return;
		// Block dismissal while Stripe or the confirm call is in flight.
		if (isProcessing) return;
		// A charged-but-unrecorded payment must not send the shopper back to the
		// checkout form, where "Place Order" would create a second order.
		if (unrecordedIntent) {
			onPending();
			return;
		}
		onClose();
	};

	return (
		<Dialog open onOpenChange={handleOpenChange}>
			<DialogContent className="max-w-md p-6">
				<DialogHeader className="mb-4">
					<DialogTitle className="text-lg font-bold text-foreground flex items-center gap-2">
						<CreditCard className="size-5 text-accent-brand" />
						<span>{t("modalTitle")}</span>
					</DialogTitle>
					<DialogDescription className="text-xs text-muted-foreground">
						{t("modalSubtitle", { orderId, amount: totalFormatted })}
					</DialogDescription>
				</DialogHeader>

				{unrecordedIntent ? (
					<div className="space-y-4">
						<div
							role="alert"
							className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-3.5 text-xs font-medium text-amber-800 dark:text-amber-300 flex items-start gap-2"
						>
							<AlertCircle className="size-4 shrink-0 mt-0.5" aria-hidden="true" />
							<span>{errorMessage || t("confirmFailed")}</span>
						</div>
						<button
							type="button"
							disabled={isProcessing}
							onClick={() => {
								setIsProcessing(true);
								setErrorMessage(null);
								void confirmOrder(unrecordedIntent);
							}}
							className="w-full py-3.5 px-6 rounded-xl bg-accent-brand text-white font-bold text-sm hover:bg-accent-brand/90 transition-all flex items-center justify-center gap-2 disabled:opacity-50 cursor-pointer shadow-md"
						>
							{isProcessing ? (
								<Loader2 className="size-4 animate-spin" />
							) : (
								<RefreshCw className="size-4" />
							)}
							<span>{t("retryConfirm")}</span>
						</button>
						<button
							type="button"
							disabled={isProcessing}
							onClick={onPending}
							className="w-full py-2.5 px-4 rounded-xl border border-border text-foreground font-semibold text-xs hover:bg-secondary transition-all cursor-pointer disabled:opacity-50"
						>
							{t("dismissPending")}
						</button>
					</div>
				) : isDevMock ? (
					<div className="space-y-4">
						<div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-3.5 text-xs font-medium text-amber-800 dark:text-amber-300">
							{t("devModeNotice")}
						</div>
						<button
							type="button"
							disabled={isProcessing}
							onClick={() => {
								setIsProcessing(true);
								setErrorMessage(null);
								void confirmOrder(clientSecret.replace("_secret_mock", ""));
							}}
							className="w-full py-3.5 px-6 rounded-xl bg-accent-brand text-white font-bold text-sm hover:bg-accent-brand/90 transition-all flex items-center justify-center gap-2 disabled:opacity-50 cursor-pointer shadow-md"
						>
							{isProcessing ? (
								<Loader2 className="size-4 animate-spin" />
							) : (
								<CheckCircle2 className="size-4" />
							)}
							<span>{t("simulateTestPayment")}</span>
						</button>
					</div>
				) : stripePromise ? (
					<Elements stripe={stripePromise} options={{ clientSecret, locale }}>
						<PaymentForm
							orderId={orderId}
							totalFormatted={totalFormatted}
							isProcessing={isProcessing}
							setIsProcessing={setIsProcessing}
							errorMessage={errorMessage}
							setErrorMessage={setErrorMessage}
							onConfirm={confirmOrder}
							onPending={onPending}
						/>
					</Elements>
				) : (
					<div className="space-y-4">
						<div
							role="alert"
							className="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-xs font-semibold text-destructive flex items-center gap-2"
						>
							<AlertCircle className="size-4 shrink-0" />
							<span>{t("stripeLoadError")}</span>
						</div>
						<button
							type="button"
							onClick={onClose}
							className="w-full py-2.5 px-4 rounded-xl border border-border text-foreground font-semibold text-xs hover:bg-secondary transition-all"
						>
							{t("dismissPending")}
						</button>
					</div>
				)}
			</DialogContent>
		</Dialog>
	);
}
