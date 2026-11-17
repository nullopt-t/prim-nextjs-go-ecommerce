"use client";

import { useState, useMemo, useEffect, useRef } from "react";
import { useCart } from "@/modules/cart/model/useCart";
import { calculateCartTotals } from "@/modules/cart/utils/cartCalculations";
import { formatMoney } from "@/shared/lib/money";
import { useCheckout } from "@/modules/orders/model/useOrders";
import { orderService } from "@/modules/orders/api/orderService";
import type { CheckoutResult } from "@/modules/orders/types";
import type { SellerShippingBreakdown } from "@/modules/checkout/types";
import { useAddresses } from "@/modules/account/model/useUserAccount";
import type { Address } from "@/modules/account/types";
import { usePathname, useRouter } from "@/shared/i18n/navigation";
import { 
	CheckCircle2,
	Check,
	Copy,
	Receipt,
	Truck,
	CreditCard, 
	ShieldCheck, 
	MapPin, 
	ArrowRight, 
	ArrowLeft,
	Clock, 
	Banknote, 
	Sparkles,
	PackageCheck,
	Pencil,
	Plus,
	X,
	Tag,
	Loader2,
	AlertCircle
} from "lucide-react";
import { CountrySelect } from "@/shared/ui/CountrySelect";
import { ErrorState } from "@/shared/ui/errorState";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { motion, AnimatePresence } from "motion/react";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
	createAddressSchema,
	INITIAL_ADDRESS,
	type AddressFormValues,
} from "@/modules/account/model/addressSchema";
import {
	Form,
	FormControl,
	FormField,
	FormInput,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/shared/ui/form/form";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { cn } from "@/shared/lib/cn";
import { StripePaymentModal, isStripeConfigured, type PendingStripeOrder } from "../components/StripePaymentModal";
import type { FormEvent } from "react";

interface CheckoutAddress {
	id: string;
	fullName?: string;
	name?: string;
	line1?: string;
	line2?: string;
	street?: string;
	city?: string;
	country?: string;
}

interface CheckoutOrder {
	id: string;
	trackingNumber: string;
	total: string;
	shippingAddress: string;
	date: string;
	pending?: boolean;
}

const PENDING_ORDER_KEY = "prim:pending-stripe-order";
// shortcut: 15 min matches the backend stock reservation TTL; bump it if that setting changes.
const PENDING_ORDER_TTL = 15 * 60 * 1000;

function readPendingOrder(): PendingStripeOrder | null {
	if (typeof window === "undefined") return null;
	try {
		const raw = sessionStorage.getItem(PENDING_ORDER_KEY);
		if (!raw) return null;
		const saved = JSON.parse(raw) as { order?: PendingStripeOrder; savedAt?: number };
		if (
			!saved.order?.clientSecret ||
			!saved.order?.orderId ||
			Date.now() - (saved.savedAt ?? 0) > PENDING_ORDER_TTL
		) {
			sessionStorage.removeItem(PENDING_ORDER_KEY);
			return null;
		}
		return saved.order;
	} catch {
		return null;
	}
}

function writePendingOrder(order: PendingStripeOrder | null) {
	if (typeof window === "undefined") return;
	try {
		if (order) {
			sessionStorage.setItem(
				PENDING_ORDER_KEY,
				JSON.stringify({ order, savedAt: Date.now() }),
			);
		} else {
			sessionStorage.removeItem(PENDING_ORDER_KEY);
		}
	} catch {
		// Storage can be unavailable (private mode); the in-memory banner still works.
	}
}

function resolveShippingAddress(
	addresses: CheckoutAddress[] | undefined,
	effectiveId: string | null,
) {
	if (!addresses || addresses.length === 0) return "";
	const addr = addresses.find((a) => a.id === effectiveId);
	if (!addr) return "";
	const name = addr.fullName || addr.name || "";
	const street = addr.line1 ? `${addr.line1}${addr.line2 ? ", " + addr.line2 : ""}` : addr.street || "";
	return `${name ? name + " - " : ""}${street}, ${addr.city}, ${addr.country}`;
}

export function Checkout() {
	const t = useTranslations("checkout");
	const tCart = useTranslations("cart");
	const tAddresses = useTranslations("addresses");
	const { user } = useAuthContext();
	const { cartItems, refreshCart, resetCart, appliedCoupon, applyCoupon, removeCoupon, loading: cartLoading, errorMsg: cartError } = useCart();
	const { checkout, loading: placingOrder } = useCheckout();
	const [couponInput, setCouponInput] = useState("");
	const [couponError, setCouponError] = useState<string | null>(null);
	const [orderError, setOrderError] = useState<string | null>(null);
	const [addressModalError, setAddressModalError] = useState<string | null>(null);
	const [applyingCoupon, setApplyingCoupon] = useState(false);
	const { addresses, addAddress, refreshAddresses, loading: addressesLoading, error: addressesError } = useAddresses();
	const router = useRouter();
	const pathname = usePathname();
	const { settings: storeSettings } = usePublicSettings();

	// Step management: 1: Delivery Address, 2: Delivery Speed, 3: Payment & Confirm
	const [currentStep, setCurrentStep] = useState(1);
	// Checkout starts with a read-only order summary. The address, delivery and
	// payment inputs stay hidden until the customer confirms the summary.
	const [hasStartedCheckout, setHasStartedCheckout] = useState(false);

	// Address states
	const [selectedAddressId, setSelectedAddressId] = useState("");
	const [isAddingAddress, setIsAddingAddress] = useState(false);
	const [savingAddress, setSavingAddress] = useState(false);
	const addressSchema = useMemo(
		() => createAddressSchema(tAddresses),
		[tAddresses],
	);
	const addressForm = useForm<AddressFormValues>({
		resolver: zodResolver(addressSchema),
		defaultValues: { ...INITIAL_ADDRESS, isDefault: true },
	});

	const [shippingSpeed, setShippingSpeed] = useState("standard"); // 'standard' | 'express'
	const [deliveryNotes, setDeliveryNotes] = useState("");
	const [shippingBreakdown, setShippingBreakdown] = useState<SellerShippingBreakdown[]>([]);

	// Card checkout is supported via Stripe PaymentIntents when cards are allowed and key configured.
	const settingsLoaded = storeSettings !== null;
	const codAvailable = settingsLoaded && storeSettings?.allowCod !== false;
	const stripeKeyConfigured = isStripeConfigured;
	const cardsAvailable =
		settingsLoaded &&
		storeSettings?.allowCards !== false &&
		(stripeKeyConfigured || process.env.NODE_ENV !== "production");
	const canPlaceOrder = settingsLoaded && (codAvailable || cardsAvailable);

	const [userChosenMethod, setUserChosenMethod] = useState<"cash_on_delivery" | "card" | null>(null);

	// Derive the effective payment method based on live availability and user selection
	const selectedPaymentMethod: "cash_on_delivery" | "card" =
		userChosenMethod && ((userChosenMethod === "card" && cardsAvailable) || (userChosenMethod === "cash_on_delivery" && codAvailable))
			? userChosenMethod
			: cardsAvailable
				? "card"
				: "cash_on_delivery";

	// Order confirmation modal state
	const [completedOrder, setCompletedOrder] = useState<CheckoutOrder | null>(null);
	const [pendingStripeOrder, setPendingStripeOrder] = useState<PendingStripeOrder | null>(null);
	const [isPaymentModalOpen, setIsPaymentModalOpen] = useState(false);
	const [copiedOrderId, setCopiedOrderId] = useState(false);

	const handleCopyOrderId = (id: string) => {
		navigator.clipboard.writeText(id);
		setCopiedOrderId(true);
		setTimeout(() => setCopiedOrderId(false), 2000);
	};

	const setPersistedPendingOrder = (order: PendingStripeOrder | null) => {
		writePendingOrder(order);
		setPendingStripeOrder(order);
	};

	// Restore a card order whose payment was interrupted (close/refresh) so the
	// shopper can finish it instead of placing a duplicate order.
	useEffect(() => {
		const saved = readPendingOrder();
		if (saved) setPendingStripeOrder(saved);
	}, []);

	// Derived selection IDs without setState in effect
	const effectiveAddressId = selectedAddressId || addresses?.find((a: Address) => a.isDefault)?.id || addresses?.[0]?.id || "";

	// Calculations using standard cartCalculations utility. Shipping rules come
	// from public settings; when settings are unavailable the rate is unknown
	// and rendered as "—" instead of an invented price.
	const currency = storeSettings?.currency?.trim() ?? "";
	const shippingKnown = storeSettings !== null;
	const standardShippingCost = storeSettings?.flatRateShipping ?? 0;
	const freeShippingThreshold = storeSettings?.freeShippingThreshold ?? null;

	const {
		subtotal,
		discount,
		isFreeShipping: isStandardFree,
	} = calculateCartTotals(cartItems, appliedCoupon, freeShippingThreshold, standardShippingCost);

		// Express has no configured rate in the public settings contract yet, so
		// fall back to the standard rate instead of inventing a price. The backend
		// remains the authority on the final charge.
		const roundCents = (n: number) => Math.round(n * 100) / 100;
		const expressShippingCost = roundCents(storeSettings?.expressShippingRate ?? standardShippingCost);
		// The backend charges 0 at/above the free-shipping threshold regardless
		// of speed (it has no express rate); mirror that so the displayed
		// express total cannot drift from the charged total.
		const shippingCost = isStandardFree ? 0 : shippingSpeed === "express" ? expressShippingCost : roundCents(standardShippingCost);

	// Tax is only charged on top when prices are configured as tax-exclusive.
	const taxRate = storeSettings?.taxRatePercent ?? 0;
	const taxAmount =
		!storeSettings?.pricesIncludeTax && taxRate > 0
			? Math.round(Math.max(0, (subtotal - discount) * (taxRate / 100)) * 100) / 100
			: 0;

	// Cash-on-delivery surcharge, applied only when COD is the selected method.
	const codFee = (codAvailable && selectedPaymentMethod === "cash_on_delivery") ? roundCents(storeSettings?.codFee ?? 0) : 0;

	const finalTotal = Math.round(Math.max(0, subtotal - discount + shippingCost + taxAmount + codFee) * 100) / 100;

	const resolvedAddressString = resolveShippingAddress(addresses, effectiveAddressId);

	// Reconcile a return from 3D-Secure. Must run once: the query string is user
	// input, and re-running would repeat the confirm call and wipe the cart.
	const handledReturnRef = useRef(false);
	useEffect(() => {
		if (handledReturnRef.current || typeof window === "undefined") return;
		const params = new URLSearchParams(window.location.search);
		const redirectStatus = params.get("redirect_status");
		const orderIdParam = params.get("order_id");
		if (!redirectStatus) return;
		handledReturnRef.current = true;

		const cleanUrl = () => router.replace(pathname, { scroll: false });

		if (redirectStatus !== "succeeded" || !orderIdParam) {
			const msg = t("payment.paymentFailed");
			setOrderError(msg);
			notify.error(msg);
			cleanUrl();
			return;
		}

		void (async () => {
			try {
				await orderService.confirmPayment({
					orderId: orderIdParam,
					paymentIntentId: params.get("payment_intent") || undefined,
				});
			} catch (err) {
				// Keep the cart and report the failure; never claim a settled order.
				const msg = apiErrorMessage(err, t("payment.confirmFailed"));
				setOrderError(msg);
				notify.error(msg);
				cleanUrl();
				return;
			}

			const order = await orderService
				.getOrderById(orderIdParam)
				.then((res) => res.data)
				.catch(() => null);
			const address = order?.shippingAddress;
			setPersistedPendingOrder(null);
			if (resetCart) resetCart();
			if (refreshCart) void refreshCart();
			setCompletedOrder({
				id: orderIdParam,
				trackingNumber: order?.trackingNumber || "",
				total: formatMoney(order ? order.totalAmount / 100 : undefined, order?.currency || currency),
				shippingAddress: address
					? [address.street, address.city, address.state, address.country].filter(Boolean).join(", ")
					: resolvedAddressString,
				date: t("success.today"),
				pending: order?.status === "pending",
			});
			window.scrollTo({ top: 0, behavior: "smooth" });
			cleanUrl();
		})();
	}, [pathname, router, t, resetCart, refreshCart, currency, resolvedAddressString]);

	const handleOpenAddAddress = () => {
		setAddressModalError(null);
		addressForm.reset({
			...INITIAL_ADDRESS,
			fullName: user?.name || "",
			phone: user?.phone || "",
			isDefault: !addresses || addresses.length === 0,
		});
		setIsAddingAddress(true);
	};

	const handleSaveNewAddress = async (values: AddressFormValues) => {
		setSavingAddress(true);
		setAddressModalError(null);
		try {
			const res = await addAddress(values);
			if (res.success) {
				setAddressModalError(null);
				if (res.data?.id) {
					setSelectedAddressId(res.data.id);
				}
				setIsAddingAddress(false);
				if (refreshAddresses) await refreshAddresses();
			} else {
				const msg = apiErrorMessage(res, tAddresses("failedToAdd"));
				setAddressModalError(msg);
				notify.error(msg);
			}
		} catch (err) {
			const msg = apiErrorMessage(err, tAddresses("failedToAdd"));
			setAddressModalError(msg);
			notify.error(msg);
		} finally {
			setSavingAddress(false);
		}
	};

	const handleApplyCoupon = async (codeToApply?: string) => {
		const targetCode = (codeToApply || couponInput).trim();
		setCouponError(null);
		if (!targetCode) {
			const msg = t("validation.enterCoupon");
			setCouponError(msg);
			notify.error(msg);
			return;
		}
		setApplyingCoupon(true);
		try {
			const res = await applyCoupon(targetCode);
			if (res.success) {
				setCouponError(null);
				setCouponInput("");
			} else {
				const msg = apiErrorMessage(res, t("validation.couponFailed"));
				setCouponError(msg);
				notify.error(msg);
			}
		} catch (err) {
			const msg = apiErrorMessage(err, t("validation.invalidCoupon"));
			setCouponError(msg);
			notify.error(msg);
		} finally {
			setApplyingCoupon(false);
		}
	};

	const handleRemoveCoupon = () => {
		removeCoupon();
		notify.info(t("validation.couponRemoved"));
	};

	// Validation helpers
	const validateStep1 = () => {
		if (cartLoading && (!cartItems || cartItems.length === 0)) {
		return (
			<div role="status" aria-live="polite" aria-busy="true" className="max-w-2xl mx-auto py-16 px-4">
				<div className="bg-card border border-border rounded-3xl p-8 sm:p-12 shadow-sm space-y-4">
					<div className="h-6 w-40 animate-pulse rounded bg-secondary/70" />
					<div className="h-4 w-full animate-pulse rounded bg-secondary/50" />
					<div className="h-4 w-2/3 animate-pulse rounded bg-secondary/50" />
					<div className="h-12 w-full animate-pulse rounded-xl bg-secondary/50" />
				</div>
			</div>
		);
	}

	if (cartError && (!cartItems || cartItems.length === 0)) {
		return (
			<div className="max-w-2xl mx-auto py-16 px-4">
				<ErrorState onRetry={() => void refreshCart?.()} />
			</div>
		);
	}

	if (!cartItems || cartItems.length === 0) {
			notify.error(tCart("empty.title"));
			router.push("/cart");
			return false;
		}

		if (!addresses || addresses.length === 0) {
			notify.error(t("validation.addAddress"));
			return false;
		}

		if (!effectiveAddressId) {
			notify.error(t("validation.selectAddress"));
			return false;
		}

		return true;
	};

	const handleStartCheckout = () => {
		if (!cartItems || cartItems.length === 0) {
			notify.error(tCart("empty.title"));
			router.push("/cart");
			return;
		}
		setHasStartedCheckout(true);
		window.scrollTo({ top: 0, behavior: "smooth" });
	};

	const handleGoToShipping = (e?: FormEvent) => {
		if (e) e.preventDefault();
		if (validateStep1()) {
			setCurrentStep(2);
			window.scrollTo({ top: 0, behavior: "smooth" });
		}
	};

	const handleGoToPayment = (e?: FormEvent) => {
		if (e) e.preventDefault();
		setCurrentStep(3);
		window.scrollTo({ top: 0, behavior: "smooth" });
	};

	const handlePlaceOrder = async (e?: FormEvent) => {
		if (e) e.preventDefault();
		setOrderError(null);

		if (!validateStep1()) {
			const msg = t("validation.selectAddress");
			setOrderError(msg);
			setCurrentStep(1);
			return;
		}

		if (!canPlaceOrder || (selectedPaymentMethod === "card" && !cardsAvailable) || (selectedPaymentMethod === "cash_on_delivery" && !codAvailable)) {
			const msg = t("validation.paymentUnavailable");
			setOrderError(msg);
			notify.error(msg);
			return;
		}

		const finalShippingAddress = resolvedAddressString;

		const orderPayload = {
			shippingAddressId: effectiveAddressId,
			couponId: appliedCoupon?.id || undefined,
			couponCode: appliedCoupon?.code || undefined,
			paymentMethod: selectedPaymentMethod,
		};

		const result = await checkout(orderPayload);

		if (result.success) {
			setOrderError(null);
			// Checkout returns { data: CheckoutResult }; the mutation wraps it once more.
			const data = (result.data?.data ?? result.data) as CheckoutResult | undefined;
			setShippingBreakdown(Array.isArray(data?.shippingBreakdown) ? data.shippingBreakdown : []);
			const clientSecret = data?.clientSecret;
			const completedOrderInfo = {
				id: data?.orderId || "",
				trackingNumber: "",
				total: formatMoney(finalTotal, currency),
				shippingAddress: finalShippingAddress,
				date: t("success.today"),
			};

			if (selectedPaymentMethod === "card") {
				if (!clientSecret) {
					const errorMsg = t("validation.cardInitFailed");
					setOrderError(errorMsg);
					notify.error(errorMsg);
					return;
				}

				setPersistedPendingOrder({
					clientSecret,
					orderId: completedOrderInfo.id,
					totalFormatted: completedOrderInfo.total,
					completedOrderInfo,
				});
				setIsPaymentModalOpen(true);
				return;
			}

			if (resetCart) resetCart();
			if (refreshCart) await refreshCart();
			setCompletedOrder(completedOrderInfo);
			window.scrollTo({ top: 0, behavior: "smooth" });
		} else {
			const errorMsg = typeof result.error === "string" ? result.error : t("validation.orderFailed");
			setOrderError(errorMsg);
			notify.error(errorMsg);
		}
	};

	// Kept as elements so the empty-cart screen (cart cleared by checkout) can
	// still offer the resume flow after a refresh.
	const resumeBanner = pendingStripeOrder && !isPaymentModalOpen ? (
		<div className="max-w-4xl mx-auto mb-6 rounded-2xl border border-accent-brand/30 bg-accent-brand/5 p-4 sm:p-5 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-sm">
			<div className="flex items-center gap-3">
				<CreditCard className="size-5 text-accent-brand shrink-0" />
				<div>
					<p className="text-sm font-bold text-foreground">
						{t("payment.modalTitle")}
					</p>
					<p className="text-xs text-muted-foreground">
						{t("payment.pendingNotice", { orderId: pendingStripeOrder.orderId })}
					</p>
				</div>
			</div>
			<div className="flex items-center gap-2 w-full sm:w-auto">
				<button
					type="button"
					onClick={() => setIsPaymentModalOpen(true)}
					className="flex-1 sm:flex-none px-4 py-2 rounded-xl bg-accent-brand text-white font-bold text-xs hover:bg-accent-brand/90 transition-all cursor-pointer shadow-sm"
				>
					{t("payment.resumePayment")}
				</button>
				<button
					type="button"
					onClick={() => setPersistedPendingOrder(null)}
					className="px-3 py-2 rounded-xl border border-border text-muted-foreground font-medium text-xs hover:bg-secondary transition-all cursor-pointer"
				>
					{t("payment.dismissPending")}
				</button>
			</div>
		</div>
	) : null;

	const paymentModal = pendingStripeOrder && isPaymentModalOpen ? (
		<StripePaymentModal
			clientSecret={pendingStripeOrder.clientSecret}
			orderId={pendingStripeOrder.orderId}
			totalFormatted={pendingStripeOrder.totalFormatted}
			onClose={() => setIsPaymentModalOpen(false)}
			onSuccess={() => {
				const info = pendingStripeOrder.completedOrderInfo;
				setPersistedPendingOrder(null);
				setIsPaymentModalOpen(false);
				if (resetCart) resetCart();
				void refreshCart?.();
				setCompletedOrder(info);
				window.scrollTo({ top: 0, behavior: "smooth" });
			}}
			onPending={() => {
				// Payment not settled yet: the order exists, so show it as pending
				// rather than as paid, and let the webhook finalize it.
				const info = pendingStripeOrder.completedOrderInfo;
				setPersistedPendingOrder(null);
				setIsPaymentModalOpen(false);
				if (resetCart) resetCart();
				void refreshCart?.();
				setCompletedOrder({ ...info, pending: true });
				window.scrollTo({ top: 0, behavior: "smooth" });
			}}
		/>
	) : null;

	if (completedOrder) {
		return (
			<div className="max-w-2xl mx-auto py-12 px-4">
				<div className="relative overflow-hidden bg-card border border-border/80 rounded-3xl p-8 sm:p-12 text-center shadow-xl flex flex-col items-center">
					{/* Subtle brand top accent gradient */}
					<div className="absolute top-0 inset-x-0 h-1.5 bg-gradient-to-r from-accent-brand/40 via-accent-brand to-accent-brand/40" />

					{/* Icon container */}
					<div className={cn(
						"size-20 rounded-2xl flex items-center justify-center mb-5 shadow-xs ring-8 transition-all",
						completedOrder.pending
							? "bg-amber-500/10 text-amber-600 ring-amber-500/5"
							: "bg-accent-brand/10 text-accent-brand ring-accent-brand/5"
					)}>
						{completedOrder.pending ? <Clock className="size-10 stroke-[2]" /> : <PackageCheck className="size-10 stroke-[2]" />}
					</div>

					{/* Status Badge */}
					<span className={cn(
						"inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border mb-4 shadow-2xs",
						completedOrder.pending
							? "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400"
							: "border-accent-brand/30 bg-accent-brand/10 text-accent-brand"
					)}>
						<span className={cn("size-1.5 rounded-full", completedOrder.pending ? "bg-amber-500 animate-pulse" : "bg-accent-brand")} />
						{completedOrder.pending ? t("success.pendingBadge") : t("success.badge")}
					</span>

					{/* Headings */}
					<h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground mb-2">
						{completedOrder.pending ? t("success.pendingTitle") : t("success.title")}
					</h1>
					<p className="text-muted-foreground text-xs sm:text-sm max-w-md mb-8 leading-relaxed">
						{completedOrder.pending ? t("success.pendingDesc") : t("success.desc")}
					</p>

					{/* Summary Box */}
					<div className="w-full bg-secondary/30 dark:bg-secondary/20 border border-border/80 rounded-2xl p-5 sm:p-6 text-start mb-8 divide-y divide-border/60 space-y-3.5">
						{completedOrder.id ? (
							<div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-3.5 text-xs sm:text-sm">
								<span className="flex items-center gap-2 text-muted-foreground font-medium">
									<Receipt className="size-4 text-accent-brand shrink-0" />
									{t("success.orderNumber")}
								</span>
								<div className="flex items-center gap-1.5">
									<span className="font-mono text-xs sm:text-sm font-bold text-foreground bg-background px-2.5 py-1 rounded-lg border border-border/80 shadow-2xs select-all">
										{completedOrder.id}
									</span>
									<button
										type="button"
										onClick={() => handleCopyOrderId(completedOrder.id!)}
										title={copiedOrderId ? "Copied" : "Copy"}
										aria-label="Copy order ID"
										className="p-1.5 rounded-lg border border-border/80 bg-background hover:bg-secondary text-muted-foreground hover:text-foreground transition-all cursor-pointer"
									>
										{copiedOrderId ? <Check className="size-3.5 text-accent-brand" /> : <Copy className="size-3.5" />}
									</button>
								</div>
							</div>
						) : null}

						{completedOrder.trackingNumber ? (
							<div className="flex items-center justify-between gap-2 py-3.5 text-xs sm:text-sm">
								<span className="flex items-center gap-2 text-muted-foreground font-medium">
									<Truck className="size-4 text-accent-brand shrink-0" />
									{t("success.trackingNumber")}
								</span>
								<span className="font-mono font-bold text-accent-brand bg-accent-brand/5 border border-accent-brand/20 px-2.5 py-1 rounded-lg">
									{completedOrder.trackingNumber}
								</span>
							</div>
						) : null}

						<div className="flex items-center justify-between gap-2 py-3.5 text-xs sm:text-sm">
							<span className="flex items-center gap-2 text-muted-foreground font-medium">
								<CreditCard className="size-4 text-accent-brand shrink-0" />
								{completedOrder.pending ? t("success.orderTotal") : t("success.totalPaid")}
							</span>
							<span className="font-extrabold text-foreground text-base sm:text-lg">
								{completedOrder.total}
							</span>
						</div>

						<div className="flex items-start justify-between gap-4 pt-3.5 text-xs sm:text-sm">
							<span className="flex items-center gap-2 text-muted-foreground font-medium shrink-0">
								<MapPin className="size-4 text-accent-brand shrink-0" />
								{t("success.destination")}
							</span>
							<span className="text-foreground font-medium text-end max-w-[260px] leading-relaxed">
								{completedOrder.shippingAddress}
							</span>
						</div>
					</div>

					{/* Action Buttons */}
					<div className="flex flex-col sm:flex-row gap-3 w-full justify-center">
						<button
							type="button"
							onClick={() => router.push("/user/orders")}
							className="flex-1 py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-bold text-sm shadow-md transition-all flex items-center justify-center gap-2 cursor-pointer"
						>
							<span>{t("success.trackOrders")}</span>
							<ArrowRight className="size-4 rtl:rotate-180" />
						</button>
						<button
							type="button"
							onClick={() => router.push("/products")}
							className="flex-1 py-3.5 px-6 rounded-xl bg-background hover:bg-secondary/80 border border-border text-foreground font-semibold text-sm transition-all flex items-center justify-center gap-2 cursor-pointer"
						>
							{t("success.continueShopping")}
						</button>
					</div>
				</div>
			</div>
		);
	}

	// While the cart is still loading, show a skeleton instead of flashing the
	// "cart empty" state (initial mount / hydration would otherwise show empty).
	if (cartLoading) {
		return (
			<div role="status" aria-live="polite" aria-busy="true" className="max-w-2xl mx-auto py-16 px-4">
				<div className="bg-card border border-border rounded-3xl p-8 sm:p-12 shadow-sm space-y-4">
					<div className="h-6 w-40 animate-pulse rounded bg-secondary/70" />
					<div className="h-4 w-full animate-pulse rounded bg-secondary/50" />
					<div className="h-4 w-2/3 animate-pulse rounded bg-secondary/50" />
					<div className="h-12 w-full animate-pulse rounded-xl bg-secondary/50" />
				</div>
			</div>
		);
	}

	if (!cartItems || cartItems.length === 0) {
		// Checkout already cleared the cart server-side; keep the pending card
		// order reachable after a refresh instead of showing an empty cart.
		if (pendingStripeOrder) {
			return (
				<div className="w-full py-4 sm:py-8">
					{resumeBanner}
					{paymentModal}
				</div>
			);
		}
		return (
				<div className="max-w-2xl mx-auto py-16 px-4">
					<div className="bg-card border border-border rounded-3xl p-8 sm:p-12 text-center shadow-sm flex flex-col items-center">
						<div className="w-16 h-16 rounded-2xl bg-secondary flex items-center justify-center text-muted-foreground mb-4">
							<PackageCheck className="size-8 stroke-[1.5]" />
						</div>
						<h2 className="text-2xl font-bold text-foreground mb-2">{tCart("empty.title")}</h2>
						<p className="text-muted-foreground text-sm max-w-sm mb-6">
							{t("empty.description")}
						</p>
						<button
							onClick={() => router.push("/products")}
							className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-primary text-primary-foreground font-medium text-sm hover:bg-primary/90 transition-colors shadow-sm"
						>
							<span>{t("empty.discover")}</span>
							<ArrowRight className="size-4 rtl:rotate-180" />
						</button>
					</div>
				</div>
		);
	}

	const stepsConfig = [
		{ id: 1, label: t("steps.address"), shortLabel: t("steps.addressShort"), icon: MapPin },
		{ id: 2, label: t("steps.shipping"), shortLabel: t("steps.shippingShort"), icon: Truck },
		{ id: 3, label: t("steps.payment"), shortLabel: t("steps.paymentShort"), icon: CreditCard },
	];

	return (
		<div className="w-full py-4 sm:py-8">
				{resumeBanner}

				{/* Step Progress Bar */}
				<div className={hasStartedCheckout ? "mb-10" : "hidden"}>
					<div className="max-w-xl mx-auto">
						<nav aria-label={t("steps.progressAria")}>
							<ol className="flex items-center justify-between relative">
								{/* Connecting background line */}
								<div className="absolute start-0 top-1/2 -translate-y-1/2 w-full h-0.5 bg-border -z-10" />
								{/* Active progress colored line */}
								<div
									className="absolute start-0 top-1/2 -translate-y-1/2 h-0.5 bg-accent-brand transition-all duration-300 -z-10"
									style={{
										width: currentStep === 1 ? "0%" : currentStep === 2 ? "50%" : "100%",
									}}
								/>

								{stepsConfig.map((step) => {
									const isCompleted = step.id < currentStep;
									const isActive = step.id === currentStep;
									const Icon = step.icon;

									return (
										<li key={step.id} className="flex flex-col items-center">
											<button
												type="button"
												aria-current={isActive ? "step" : undefined}
												disabled={step.id > currentStep}
												onClick={() => {
													if (step.id < currentStep) {
														setCurrentStep(step.id);
													}
												}}
												className={`size-10 sm:size-11 rounded-full flex items-center justify-center font-bold text-xs sm:text-sm transition-all shadow-xs ${
													isCompleted
														? "bg-accent-brand text-white hover:bg-accent-brand/90 cursor-pointer ring-4 ring-accent-brand/10"
														: isActive
														? "bg-primary text-primary-foreground ring-4 ring-primary/20 scale-105"
														: "bg-background border-2 border-border text-muted-foreground cursor-not-allowed"
												}`}
											>
												{isCompleted ? (
													<Check className="size-5" />
												) : (
													<Icon className="size-4 sm:size-5" />
												)}
											</button>
											<span
												className={`text-[11px] sm:text-xs font-semibold mt-2 text-center transition-colors ${
													isActive
														? "text-foreground font-bold"
														: isCompleted
														? "text-accent-brand cursor-pointer"
														: "text-muted-foreground"
												}`}
											>
												<span className="hidden sm:inline">{step.label}</span>
												<span className="sm:hidden">{step.shortLabel}</span>
											</span>
										</li>
									);
								})}
							</ol>
						</nav>
					</div>
				</div>

				<div
					className={
						hasStartedCheckout
							? "grid grid-cols-1 lg:grid-cols-12 gap-8 items-start"
							: "grid grid-cols-1 gap-8 max-w-2xl mx-auto"
					}
				>
					{/* Left Column: Multi-step Active View — hidden until the
					    customer confirms the order summary. */}
					<div className={hasStartedCheckout ? "lg:col-span-7 flex flex-col gap-6" : "hidden"}>
						<AnimatePresence mode="wait">
							
							{/* STEP 1: DELIVERY ADDRESS */}
							{currentStep === 1 && (
								<motion.div
									key="step-address"
									initial={{ opacity: 0, y: 8 }}
									animate={{ opacity: 1, y: 0 }}
									exit={{ opacity: 0, y: -8 }}
									transition={{ duration: 0.2 }}
									className="bg-card border border-border rounded-2xl p-6 sm:p-7 shadow-xs"
								>
									<div className="flex items-center justify-between pb-5 border-b border-border/60 mb-6">
										<div className="flex items-center gap-3">
											<div className="size-9 rounded-xl bg-accent-brand/10 text-accent-brand flex items-center justify-center font-bold">
												<MapPin className="size-5" />
											</div>
											<div>
												<h2 className="text-lg font-bold text-foreground">{t("address.heading")}</h2>
												<p className="text-xs text-muted-foreground">{t("address.subtitle")}</p>
											</div>
										</div>
										{!isAddingAddress && (
											<button
												type="button"
												onClick={handleOpenAddAddress}
												className="inline-flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-3 py-1.5 rounded-lg hover:bg-accent-brand/5 transition-colors"
											>
												<Plus className="size-3.5" />
												<span>{tAddresses("addAddress")}</span>
											</button>
										)}
									</div>

									{isAddingAddress ? (
																				<Form {...addressForm}>
											<form
												onSubmit={addressForm.handleSubmit(handleSaveNewAddress)}
												className="space-y-4"
												noValidate
											>
												<div className="flex items-center justify-between pb-2 border-b border-border/40">
													<span className="text-xs font-semibold text-foreground">{t("address.newAddress")}</span>
													{addresses && addresses.length > 0 && (
														<button
															type="button"
															onClick={() => setIsAddingAddress(false)}
															className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1"
														>
															<X className="size-3.5" />
															<span>{tAddresses("cancel")}</span>
														</button>
													)}
												</div>

												{addressModalError && (
													<div
														role="alert"
														className="rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-xs font-semibold text-destructive flex items-center gap-2"
													>
														<AlertCircle className="size-4 shrink-0" aria-hidden="true" />
														<span>{addressModalError}</span>
													</div>
												)}

												{/* Label Selector */}
												<div>
													<label className="block text-xs font-semibold text-foreground mb-1.5">
														{tAddresses("labels.addressLabel")}
													</label>
													<div className="grid grid-cols-3 gap-2">
														{["Home", "Office", "Other"].map((lbl) => (
															<button
																key={lbl}
																type="button"
																onClick={() =>
																	addressForm.setValue("label", lbl, {
																		shouldDirty: true,
																	})
																}
																className={`py-2 px-3 rounded-xl border text-xs font-medium cursor-pointer transition-all ${
																	addressForm.watch("label").toLowerCase() ===
																	lbl.toLowerCase()
																		? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold"
																		: "border-border/80 bg-secondary/40 text-muted-foreground hover:text-foreground"
																}`}
															>
																{tAddresses(`labels.${lbl.toLowerCase()}`)}
															</button>
														))}
													</div>
												</div>

												<div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
													<FormField
														control={addressForm.control}
														name="fullName"
														render={({ field }) => (
															<FormItem className="sm:col-span-2">
																<FormLabel>{tAddresses("labels.fullName")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={t("address.recipientPlaceholder")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="line1"
														render={({ field }) => (
															<FormItem className="sm:col-span-2">
																<FormLabel>{t("address.street")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={t("address.streetPlaceholder")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="line2"
														render={({ field }) => (
															<FormItem className="sm:col-span-2">
																<FormLabel>{t("address.apartment")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={t("address.apartmentPlaceholder")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="city"
														render={({ field }) => (
															<FormItem>
																<FormLabel>{tAddresses("labels.city")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={tAddresses("placeholders.city")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="state"
														render={({ field }) => (
															<FormItem>
																<FormLabel>{t("address.state")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={tAddresses("placeholders.state")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="postalCode"
														render={({ field }) => (
															<FormItem>
																<FormLabel>{tAddresses("labels.postalCode")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={tAddresses("placeholders.postalCode")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="phone"
														render={({ field }) => (
															<FormItem>
																<FormLabel>{t("address.phone")}</FormLabel>
																<FormControl>
																	<FormInput
																		{...field}
																		placeholder={t("address.phonePlaceholder")}
																	/>
																</FormControl>
																<FormMessage />
															</FormItem>
														)}
													/>

													<FormField
														control={addressForm.control}
														name="country"
														render={({ field }) => (
															<FormItem className="sm:col-span-2">
																<FormLabel>{tAddresses("labels.country")}</FormLabel>
																<CountrySelect
																	value={field.value}
																	onChange={field.onChange}
																/>
																<FormMessage />
															</FormItem>
														)}
													/>
												</div>

												<div className="flex items-center justify-end gap-3 pt-4 border-t border-border">
													{addresses && addresses.length > 0 && (
														<button
															type="button"
															onClick={() => setIsAddingAddress(false)}
															className="px-4 py-2.5 text-xs font-semibold text-muted-foreground hover:text-foreground border border-border rounded-xl hover:bg-secondary transition-colors"
														>
															{tAddresses("cancel")}
														</button>
													)}
													<button
														type="submit"
														disabled={savingAddress}
														className="px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all flex items-center gap-2 disabled:opacity-50"
													>
														{savingAddress && <Loader2 className="size-3.5 animate-spin" />}
														<span>{savingAddress ? t("address.saving") : t("address.save")}</span>
													</button>
												</div>
											</form>
										</Form>
									) : addressesError ? (
										<div className="py-8 px-4 border border-border rounded-2xl bg-background">
											<ErrorState onRetry={() => void refreshAddresses?.()} />
										</div>
									) : addressesLoading ? (
										<div role="status" aria-live="polite" aria-busy="true" className="py-8 px-4 border border-border rounded-2xl space-y-3">
											<div className="h-4 w-1/2 animate-pulse rounded bg-secondary/60" />
											<div className="h-4 w-2/3 animate-pulse rounded bg-secondary/60" />
										</div>
									) : addresses && addresses.length > 0 ? (
										<div className="grid grid-cols-1 sm:grid-cols-2 gap-4" role="radiogroup" aria-label={tAddresses("title")}>
											{addresses.map((addr) => {
												const isSelected = effectiveAddressId === addr.id;
												return (
													<button
														type="button"
														key={addr.id}
														onClick={() => setSelectedAddressId(addr.id)}
														role="radio"
														aria-checked={isSelected}
														className={`p-4 rounded-xl border cursor-pointer transition-all flex flex-col justify-between text-start ${
															isSelected
																? "border-accent-brand bg-accent-brand/5 ring-2 ring-accent-brand"
																: "border-border hover:border-border/80 bg-background"
														}`}
													>
														<div>
															<div className="flex items-center justify-between mb-2">
																<span className="font-semibold text-foreground text-sm flex items-center gap-1.5 capitalize">
																	<MapPin className="size-3.5 text-accent-brand" />
																	{addr.label || addr.type || t("address.fallbackLabel")}
																</span>
																{isSelected ? (
																	<span className="size-2.5 rounded-full bg-accent-brand"></span>
																) : (
																	addr.isDefault && (
																		<span className="text-[10px] font-medium text-accent-brand bg-accent-brand/10 border border-accent-brand/20 px-2 py-0.5 rounded-md">{tAddresses("defaultBadge")}</span>
																	)
																)}
															</div>
															<p className="text-xs font-semibold text-foreground mb-1">{addr.fullName || addr.name || user?.name}</p>
															<p className="text-xs text-muted-foreground leading-relaxed">
																{addr.line1 || addr.street}
																{addr.line2 ? `, ${addr.line2}` : ""}
															</p>
															<p className="text-xs text-muted-foreground">{addr.city}{addr.state ? `, ${addr.state}` : ""}, {addr.country}</p>
														</div>
														{addr.phone && (
															<span className="text-[11px] text-muted-foreground mt-3 font-medium">{addr.phone}</span>
														)}
													</button>
												);
											})}

											{/* Add Address Card Option */}
											<button
												type="button"
												onClick={handleOpenAddAddress}
												className="p-5 rounded-xl border border-dashed border-border hover:border-accent-brand/60 bg-secondary/10 hover:bg-accent-brand/5 cursor-pointer transition-all flex flex-col items-center justify-center text-center gap-2 min-h-[140px]"
											>
												<div className="size-8 rounded-full bg-accent-brand/10 text-accent-brand flex items-center justify-center">
													<Plus className="size-4" />
												</div>
												<span className="text-xs font-semibold text-foreground">{t("address.addAnother")}</span>
												<span className="text-[11px] text-muted-foreground">{t("address.addAnotherDesc")}</span>
											</button>
										</div>
									) : (
										<div className="py-8 px-4 border border-dashed border-border rounded-2xl bg-secondary/20 text-center flex flex-col items-center">
											<div className="size-12 rounded-2xl bg-accent-brand/10 text-accent-brand flex items-center justify-center mb-3">
												<MapPin className="size-6" />
											</div>
											<h3 className="text-sm font-bold text-foreground mb-1">{t("address.noneTitle")}</h3>
											<p className="text-xs text-muted-foreground max-w-sm mb-5 leading-relaxed">
												{t("address.noneDesc")}
											</p>
											<button
												type="button"
												onClick={handleOpenAddAddress}
												className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-accent-brand text-white font-semibold text-xs shadow-sm hover:bg-accent-brand/90 transition-all cursor-pointer"
											>
												<Plus className="size-4" />
												<span>{t("address.addShipping")}</span>
											</button>
										</div>
									)}

									{/* Step 1 Actions */}
									<div className="flex flex-wrap items-center justify-between gap-3 pt-6 border-t border-border mt-8">
										<button
											type="button"
											onClick={() => router.push("/cart")}
											className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors px-3 py-2 rounded-xl hover:bg-secondary"
										>
											<ArrowLeft className="size-4 rtl:rotate-180" />
											<span>{t("actions.returnToCart")}</span>
										</button>
										<button
											type="button"
											onClick={handleGoToShipping}
											className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-semibold text-sm rounded-xl shadow-sm hover:bg-accent-brand/90 transition-all"
										>
											<span>{t("actions.continueToDelivery")}</span>
											<ArrowRight className="size-4 rtl:rotate-180" />
										</button>
									</div>
								</motion.div>
							)}

							{/* STEP 2: DELIVERY SPEED */}
							{currentStep === 2 && (
								<motion.div
									key="step-shipping"
									initial={{ opacity: 0, y: 8 }}
									animate={{ opacity: 1, y: 0 }}
									exit={{ opacity: 0, y: -8 }}
									transition={{ duration: 0.2 }}
									className="space-y-6"
								>
									{/* Recap of Step 1 */}
									<div className="bg-secondary/30 border border-border/80 rounded-2xl p-4 flex items-center justify-between">
										<div className="flex items-start gap-3 min-w-0">
											<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
												<MapPin className="size-4 text-accent-brand" />
											</div>
											<div className="min-w-0">
												<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{t("summary.shipTo")}</span>
												<p className="text-xs sm:text-sm font-semibold text-foreground line-clamp-1">{resolvedAddressString}</p>
											</div>
										</div>
										<button
											type="button"
											onClick={() => setCurrentStep(1)}
											className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-3 py-1.5 rounded-lg hover:bg-accent-brand/5 shrink-0"
										>
											<Pencil className="size-3.5" />
											<span>{t("actions.change")}</span>
										</button>
									</div>

									{/* Main Delivery Speed Selector */}
									<div className="bg-card border border-border rounded-2xl p-6 sm:p-7 shadow-xs">
										<div className="flex items-center gap-3 pb-5 border-b border-border/60 mb-6">
											<div className="size-9 rounded-xl bg-accent-brand/10 text-accent-brand flex items-center justify-center font-bold">
												<Truck className="size-5" />
											</div>
											<div>
												<h2 className="text-lg font-bold text-foreground">{t("delivery.heading")}</h2>
												<p className="text-xs text-muted-foreground">{t("delivery.subtitle")}</p>
											</div>
										</div>

										<div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
											<button
												type="button"
												onClick={() => setShippingSpeed("standard")}
												aria-pressed={shippingSpeed === "standard"}
												className={`p-5 rounded-2xl border cursor-pointer transition-all flex flex-col justify-between text-start ${
													shippingSpeed === "standard"
														? "border-accent-brand bg-accent-brand/5 ring-2 ring-accent-brand"
														: "border-border hover:border-border/80 bg-background"
												}`}
											>
												<div>
													<div className="flex items-start justify-between mb-3">
														<div className="flex items-center gap-2">
															<div className="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400">
																<Truck className="size-4" />
															</div>
															<div>
																<span className="font-bold text-sm text-foreground block">{t("delivery.standard")}</span>
																<span className="text-[11px] text-muted-foreground">{t("delivery.standardTime")}</span>
															</div>
														</div>
														<span className="text-xs font-extrabold text-emerald-600 dark:text-emerald-400">
															{isStandardFree ? tCart("orderSummary.free") : shippingKnown ? formatMoney(standardShippingCost, currency) : "—"}
														</span>
													</div>
													<p className="text-xs text-muted-foreground leading-relaxed">
														{t("delivery.standardDesc", { threshold: freeShippingThreshold !== null ? formatMoney(freeShippingThreshold, currency) : "—" })}
													</p>
												</div>
												{shippingSpeed === "standard" && (
													<div className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand mt-4 pt-3 border-t border-accent-brand/20">
														<CheckCircle2 className="size-4" />
														<span>{t("delivery.selected")}</span>
													</div>
												)}
											</button>

											<button
												type="button"
												onClick={() => setShippingSpeed("express")}
												aria-pressed={shippingSpeed === "express"}
												className={`p-5 rounded-2xl border cursor-pointer transition-all flex flex-col justify-between text-start ${
													shippingSpeed === "express"
														? "border-accent-brand bg-accent-brand/5 ring-2 ring-accent-brand"
														: "border-border hover:border-border/80 bg-background"
												}`}
											>
												<div>
													<div className="flex items-start justify-between mb-3">
														<div className="flex items-center gap-2">
															<div className="p-2 rounded-lg bg-amber-50 dark:bg-amber-950/40 text-amber-500">
																<Clock className="size-4" />
															</div>
															<div>
																<span className="font-bold text-sm text-foreground block">{t("delivery.express")}</span>
																<span className="text-[11px] text-muted-foreground">{t("delivery.expressTime")}</span>
															</div>
														</div>
														<span className="text-xs font-extrabold text-foreground">{shippingKnown ? formatMoney(expressShippingCost, currency) : "—"}</span>
													</div>
													<p className="text-xs text-muted-foreground leading-relaxed">
														{t("delivery.expressDesc")}
													</p>
												</div>
												{shippingSpeed === "express" && (
													<div className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand mt-4 pt-3 border-t border-accent-brand/20">
														<CheckCircle2 className="size-4" />
														<span>{t("delivery.selected")}</span>
													</div>
												)}
											</button>
										</div>

										{/* Optional Delivery Instructions */}
										<div className="space-y-2 pt-4 border-t border-border/60">
											<label htmlFor="delivery-notes" className="text-xs font-semibold text-foreground">{t("delivery.instructions")}</label>
											<input
												id="delivery-notes"
												type="text"
												placeholder={t("delivery.instructionsPlaceholder")}
												value={deliveryNotes}
												onChange={(e) => setDeliveryNotes(e.target.value)}
												className="w-full px-3.5 py-2.5 bg-background border border-border rounded-xl text-xs text-foreground focus:outline-none focus:ring-2 focus:ring-accent-brand"
											/>
										</div>

										{/* Step 2 Actions */}
										<div className="flex flex-wrap items-center justify-between gap-3 pt-6 border-t border-border mt-8">
											<button
												type="button"
												onClick={() => setCurrentStep(1)}
												className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors px-3 py-2 rounded-xl hover:bg-secondary"
											>
												<ArrowLeft className="size-4 rtl:rotate-180" />
												<span>{t("actions.backToAddress")}</span>
											</button>
											<button
												type="button"
												onClick={handleGoToPayment}
												className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-semibold text-sm rounded-xl shadow-sm hover:bg-accent-brand/90 transition-all"
											>
												<span>{t("actions.continueToPayment")}</span>
												<ArrowRight className="size-4 rtl:rotate-180" />
											</button>
										</div>
									</div>
								</motion.div>
							)}

							{/* STEP 3: PAYMENT & CONFIRM */}
							{currentStep === 3 && (
								<motion.div
									key="step-payment"
									initial={{ opacity: 0, y: 8 }}
									animate={{ opacity: 1, y: 0 }}
									exit={{ opacity: 0, y: -8 }}
									transition={{ duration: 0.2 }}
									className="space-y-6"
								>
									{/* Recaps of Step 1 & Step 2 */}
									<div className="bg-secondary/30 border border-border/80 rounded-2xl p-4 divide-y divide-border/60 space-y-3">
										<div className="flex items-center justify-between">
											<div className="flex items-start gap-3 min-w-0">
												<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
													<MapPin className="size-4 text-accent-brand" />
												</div>
												<div className="min-w-0">
													<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{t("summary.shipTo")}</span>
													<p className="text-xs sm:text-sm font-semibold text-foreground line-clamp-1">{resolvedAddressString}</p>
												</div>
											</div>
											<button
												type="button"
												onClick={() => setCurrentStep(1)}
												className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-2.5 py-1 rounded-lg hover:bg-accent-brand/5 shrink-0"
											>
												<Pencil className="size-3" />
												<span>{tAddresses("edit")}</span>
											</button>
										</div>

										<div className="flex items-center justify-between pt-3">
											<div className="flex items-start gap-3 min-w-0">
												<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
													<Truck className="size-4 text-accent-brand" />
												</div>
												<div>
													<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{t("steps.shipping")}</span>
													<p className="text-xs sm:text-sm font-semibold text-foreground">
														{shippingSpeed === "express" ? t("delivery.expressRecap", { price: shippingKnown ? formatMoney(expressShippingCost, currency) : "—" }) : t("delivery.standardRecap", { price: isStandardFree ? tCart("orderSummary.free") : shippingKnown ? formatMoney(standardShippingCost, currency) : "—" })}
													</p>
												</div>
											</div>
											<button
												type="button"
												onClick={() => setCurrentStep(2)}
												className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-2.5 py-1 rounded-lg hover:bg-accent-brand/5 shrink-0"
											>
												<Pencil className="size-3" />
												<span>{tAddresses("edit")}</span>
											</button>
										</div>
									</div>

									{/* Main Payment Section */}
									<div className="bg-card border border-border rounded-2xl p-6 sm:p-7 shadow-xs">
										<div className="flex items-center gap-3 pb-5 border-b border-border/60 mb-6">
											<div className="size-9 rounded-xl bg-accent-brand/10 text-accent-brand flex items-center justify-center font-bold">
												<CreditCard className="size-5" />
											</div>
											<div>
												<h2 className="text-lg font-bold text-foreground">{t("payment.heading")}</h2>
												<p className="text-xs text-muted-foreground">{t("payment.security")}</p>
											</div>
										</div>

										{cardsAvailable || codAvailable ? (
											<div className="space-y-3 mb-6" role="radiogroup" aria-label={t("payment.heading")}>
												{cardsAvailable && (
													<button
														type="button"
														role="radio"
														aria-checked={selectedPaymentMethod === "card"}
														onClick={() => setUserChosenMethod("card")}
														className={cn(
															"w-full flex items-center justify-between gap-3 rounded-xl border p-4 text-xs font-semibold transition-all text-start cursor-pointer",
															selectedPaymentMethod === "card"
																? "border-accent-brand bg-accent-brand/10 text-accent-brand ring-1 ring-accent-brand"
																: "border-border bg-card text-foreground hover:bg-secondary/30"
														)}
													>
														<div className="flex items-center gap-3">
															<CreditCard className="size-5 shrink-0 text-accent-brand" />
															<div className="flex flex-col gap-0.5">
																<span className="font-semibold text-foreground">{t("payment.creditCard")}</span>
																<span className="text-[11px] font-normal text-muted-foreground">{t("payment.stripeCardDesc")}</span>
															</div>
														</div>
														{selectedPaymentMethod === "card" && <CheckCircle2 className="size-4 shrink-0 text-accent-brand" />}
													</button>
												)}

												{codAvailable && (
													<button
														type="button"
														role="radio"
														aria-checked={selectedPaymentMethod === "cash_on_delivery"}
														onClick={() => setUserChosenMethod("cash_on_delivery")}
														className={cn(
															"w-full flex items-center justify-between gap-3 rounded-xl border p-4 text-xs font-semibold transition-all text-start cursor-pointer",
															selectedPaymentMethod === "cash_on_delivery"
																? "border-accent-brand bg-accent-brand/10 text-accent-brand ring-1 ring-accent-brand"
																: "border-border bg-card text-foreground hover:bg-secondary/30"
														)}
													>
														<div className="flex items-center gap-3">
															<Banknote className="size-5 shrink-0 text-accent-brand" />
															<div className="flex flex-col gap-0.5">
																<span className="font-semibold text-foreground">{t("payment.codTitle")}</span>
																<span className="text-[11px] font-normal text-muted-foreground leading-relaxed">{t("payment.codDesc")}</span>
															</div>
														</div>
														{selectedPaymentMethod === "cash_on_delivery" && <CheckCircle2 className="size-4 shrink-0 text-accent-brand" />}
													</button>
												)}
											</div>
										) : (
											<div className="mb-6 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-xs text-amber-800 dark:text-amber-300">
												{t("payment.noPaymentMethods")}
											</div>
										)}

										{orderError && (
											<div
												role="alert"
												className="rounded-xl border border-destructive/30 bg-destructive/10 p-3.5 text-xs font-semibold text-destructive flex items-center gap-2 mt-4"
											>
												<AlertCircle className="size-4 shrink-0" aria-hidden="true" />
												<span>{orderError}</span>
											</div>
										)}

										{/* Step 3 Actions */}
										<div className="flex flex-wrap items-center justify-between gap-3 pt-6 border-t border-border mt-8">
											<button
												type="button"
												onClick={() => setCurrentStep(2)}
												className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors px-3 py-2 rounded-xl hover:bg-secondary"
											>
												<ArrowLeft className="size-4 rtl:rotate-180" />
												<span>{t("actions.backToDelivery")}</span>
											</button>
											<button
												type="button"
												onClick={handlePlaceOrder}
												disabled={placingOrder || !canPlaceOrder}
												className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-bold text-sm rounded-xl shadow-md hover:bg-accent-brand/90 transition-all disabled:opacity-50"
											>
												<span>{placingOrder ? t("actions.placingOrder") : t("actions.placeOrder", { total: finalTotal.toFixed(2) })}</span>
												<CheckCircle2 className="size-4" />
											</button>
										</div>
									</div>
								</motion.div>
							)}
						</AnimatePresence>
					</div>

					{/* Right Column: Order Summary & Quick Action */}
					<div className={hasStartedCheckout ? "lg:col-span-5 sticky top-24" : "w-full"}>
						{!hasStartedCheckout && (
							<div className="mb-5 text-center">
								<h1 className="text-2xl sm:text-3xl font-bold text-foreground">
									{t("review.heading")}
								</h1>
								<p className="mt-2 text-sm text-muted-foreground">
									{t("review.subtitle")}
								</p>
							</div>
						)}
						<div className="bg-card border border-border rounded-2xl p-6 shadow-sm flex flex-col gap-6">
							<div className="flex items-center justify-between">
								<h3 className="font-semibold text-lg text-foreground">{tCart("orderSummary.title")}</h3>
								<span className="text-xs font-medium text-muted-foreground bg-secondary px-2.5 py-1 rounded-full">
									{t("summary.itemCount", { count: cartItems?.length || 0 })}
								</span>
							</div>

							<div className="flex flex-col gap-3.5 max-h-[240px] overflow-y-auto pr-1">
								{cartItems?.map((item) => (
									<div key={item.id} className="flex items-center justify-between gap-3 text-sm py-2 border-b border-border/40 last:border-0">
										<div className="flex items-center gap-3 min-w-0">
											<div className="w-12 h-12 rounded-lg bg-secondary overflow-hidden shrink-0 border border-border/50">
												<img
													src={normalizeMediaUrl(item.img)}
													alt={(typeof item.productName === "object" ? item.productName?.en : item.productName) || t("summary.product")}
													className="w-full h-full object-cover"
												/>
											</div>
											<div className="flex flex-col min-w-0">
												<span className="font-medium text-foreground line-clamp-1">{(typeof item.productName === "object" ? item.productName?.en : item.productName) || t("summary.product")}</span>
												<span className="text-xs text-muted-foreground">{t("summary.qty", { count: item.quantity || 1 })}</span>
											</div>
										</div>
										<span className="font-semibold text-foreground shrink-0 tabular-nums whitespace-nowrap">{item.productPrice}</span>
									</div>
								))}
							</div>

							{/* Clean Promo Code Input */}
							<div className="pt-2 border-t border-border/60">
								{/* Input stays visible so an applied coupon can be replaced;
								    an invalid replacement keeps the current discount. */}
								{appliedCoupon && (
									<div className="flex items-center justify-between p-2.5 bg-emerald-500/10 border border-emerald-500/25 rounded-xl text-xs">
										<div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium min-w-0">
											<Tag className="size-3.5 text-emerald-600 shrink-0" />
											<span className="font-bold tracking-wider font-mono uppercase min-w-0 truncate">{appliedCoupon.code}</span>
											<span className="text-[11px] opacity-80">
												({appliedCoupon.discountPercent
													? t("summary.percentOff", { percent: appliedCoupon.discountPercent })
													: appliedCoupon.fixedDiscount
													? t("summary.fixedOff", { amount: appliedCoupon.fixedDiscount })
													: t("summary.applied")})
											</span>
										</div>
										<button
											type="button"
											onClick={handleRemoveCoupon}
											aria-label={t("summary.removePromo")}
											className="p-1 rounded-md text-muted-foreground hover:text-destructive hover:bg-background/80 transition-colors"
											title={t("summary.removePromo")}
										>
											<X className="size-3.5" />
										</button>
									</div>
								)}
									<div className="flex gap-2 mt-2">
										<div className="flex-1 min-w-0 bg-secondary/30 border border-border focus-within:border-accent-brand rounded-xl overflow-hidden flex items-center px-3 h-9 transition-colors">
											<Tag className="size-3.5 text-muted-foreground shrink-0 me-2" />
											<input
												type="text"
												aria-label={t("summary.promoPlaceholder")}
												placeholder={t("summary.promoPlaceholder")}
												value={couponInput}
												onChange={(e) => setCouponInput(e.target.value)}
												onKeyDown={(e) => {
													if (e.key === "Enter") {
														e.preventDefault();
														handleApplyCoupon();
													}
												}}
												className="w-full bg-transparent outline-none text-xs text-foreground placeholder:text-muted-foreground uppercase font-mono"
											/>
										</div>
										<button
											type="button"
											onClick={() => handleApplyCoupon()}
											disabled={applyingCoupon || !couponInput.trim()}
											className="px-3.5 h-9 rounded-xl bg-secondary hover:bg-secondary/80 disabled:opacity-40 text-foreground font-semibold text-xs transition-colors shrink-0 flex items-center justify-center cursor-pointer border border-border"
										>
											{applyingCoupon ? <Loader2 className="size-3.5 animate-spin" /> : tCart("coupon.apply")}
										</button>
									</div>
									{couponError && (
										<p className="text-[11px] text-destructive font-medium mt-1.5 leading-tight">
											{couponError}
										</p>
									)}
							</div>

							<div className="flex flex-col gap-2.5 pt-4 border-t border-border/60 text-sm">
								<div className="flex justify-between text-muted-foreground">
									<span>{tCart("orderSummary.subtotal")}</span>
									<span className="font-medium text-foreground tabular-nums whitespace-nowrap">{formatMoney(subtotal, currency)}</span>
								</div>

								{discount > 0 && (
									<div className="flex justify-between text-emerald-600 dark:text-emerald-400 font-medium">
										<span className="flex items-center gap-1 min-w-0">
											<Sparkles className="size-3.5 shrink-0" />
											<span className="truncate">{t("summary.discount", { code: appliedCoupon?.code ?? "" })}</span>
										</span>
										<span className="tabular-nums whitespace-nowrap">-{formatMoney(discount, currency)}</span>
									</div>
								)}

								<div className="flex justify-between text-muted-foreground">
									<span>{t("summary.shipping", { speed: shippingSpeed === "express" ? t("summary.express") : t("summary.standard") })}</span>
									<span className="font-medium text-foreground tabular-nums whitespace-nowrap">
										{shippingCost === 0 ? (
											<span className="text-emerald-600 font-medium">{tCart("orderSummary.free")}</span>
										) : (
											formatMoney(shippingCost, currency)
										)}
									</span>
								</div>

								{shippingBreakdown.length > 0 && (
									<div className="space-y-1.5 rounded-xl border border-border bg-secondary/20 p-3">
										<p className="text-xs font-semibold text-foreground">{t("summary.shippingBreakdownTitle")}</p>
										{shippingBreakdown.map((line, index) => (
											<p key={line.sellerId ?? `${line.sellerName}-${index}`} className="text-xs text-muted-foreground">
												{t("summary.vendorShipping", {
													seller: line.sellerName,
													fee: line.isFree
														? tCart("orderSummary.free")
														: `${formatMoney(line.shippingFee, currency)}${line.leadTimeDays > 0 ? ` (${t("summary.estDays", { days: line.leadTimeDays })})` : ""}`,
												})}
											</p>
										))}
									</div>
								)}

								<div className="flex justify-between text-muted-foreground">
									<span>{t("summary.taxes")}</span>
									<span className="font-medium text-foreground tabular-nums whitespace-nowrap">{formatMoney(taxAmount, currency)}</span>
								</div>

								{codFee > 0 && (
									<div className="flex justify-between text-muted-foreground">
										<span>{t("summary.codFee")}</span>
										<span className="font-medium text-foreground tabular-nums whitespace-nowrap">{formatMoney(codFee, currency)}</span>
									</div>
								)}

								<div className="flex justify-between items-center text-lg font-bold text-foreground pt-3 border-t border-border">
									<span>{t("summary.total")}</span>
									<span className="text-2xl font-extrabold text-foreground tabular-nums whitespace-nowrap">{formatMoney(finalTotal, currency)}</span>
								</div>

								{orderError && (
									<div
										role="alert"
										className="rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-xs font-semibold text-destructive flex items-center gap-2 mt-2"
									>
										<AlertCircle className="size-4 shrink-0" aria-hidden="true" />
										<span>{orderError}</span>
									</div>
								)}
							</div>

							{!hasStartedCheckout && (
								<div className="flex flex-col gap-3">
									<button
										type="button"
										onClick={handleStartCheckout}
										className="w-full py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm shadow-sm transition-all flex items-center justify-center gap-2"
									>
										<span>{t("actions.continueToCheckout")}</span>
										<ArrowRight className="size-4 rtl:rotate-180" />
									</button>
									<button
										type="button"
										onClick={() => router.push("/cart")}
										className="w-full py-2.5 px-6 rounded-xl text-muted-foreground hover:text-foreground font-medium text-xs transition-colors flex items-center justify-center gap-2"
									>
										<ArrowLeft className="size-3.5 rtl:rotate-180" />
										<span>{t("actions.returnToCart")}</span>
									</button>
								</div>
							)}

							<div className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
								<ShieldCheck className="size-4 text-emerald-600" />
								<span>{t("summary.secure")}</span>
							</div>
						</div>
					</div>
			</div>

			{paymentModal}
		</div>
	);
}
