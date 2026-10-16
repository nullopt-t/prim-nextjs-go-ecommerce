"use client";

import { useState } from "react";
import { useCart } from "@/hooks/useCart";
import { calculateCartTotals } from "@/features/cart/utils/cartCalculations";
import { useCheckout } from "@/hooks/useOrders";
import { useAddresses, usePaymentMethods } from "@/hooks/useUserAccount";
import { useRouter } from "@/i18n/navigation";
import MainLayout from "@/components/layouts/mainLayout";
import { CustomButton } from "@/components/ui";
import { 
	CheckCircle2, 
	Check,
	Truck, 
	CreditCard, 
	ShieldCheck, 
	MapPin, 
	ArrowRight, 
	ArrowLeft,
	Clock, 
	Banknote, 
	Smartphone,
	Sparkles,
	PackageCheck,
	Pencil,
	Plus,
	X,
	Tag,
	Loader2
} from "lucide-react";
import { CountrySelect } from "@/components/ui/CountrySelect";
import { toast } from "sonner";
import { motion, AnimatePresence } from "motion/react";
import { useAuthContext } from "@/context/AuthContext";
import { useTranslations } from "next-intl";

function generateOrderCode() {
	return Math.floor(10000 + Math.random() * 90000).toString();
}

function resolveShippingAddress(addresses, effectiveId) {
	if (!addresses || addresses.length === 0) return "";
	const addr = addresses.find((a) => a.id === effectiveId);
	if (!addr) return "";
	const name = addr.fullName || addr.name || "";
	const street = addr.line1 ? `${addr.line1}${addr.line2 ? ", " + addr.line2 : ""}` : addr.street || "";
	return `${name ? name + " - " : ""}${street}, ${addr.city}, ${addr.country}`;
}

function resolvePaymentLabel(paymentType, cards, effectiveCardId, newCardData) {
	if (paymentType === "saved-card") {
		const card = cards?.find((c) => c.id === effectiveCardId);
		return card ? `${card.type} ending in ${card.last4}` : "Saved Card";
	}
	if (paymentType === "cod") return "Cash on Delivery (COD)";
	if (paymentType === "apple-pay") return "Apple Pay / Digital Wallet";
	return `Card ending in ${newCardData.number.slice(-4) || "0000"}`;
}

const INITIAL_NEW_ADDRESS = {
	label: "Home",
	fullName: "",
	phone: "",
	line1: "",
	line2: "",
	city: "",
	state: "",
	postalCode: "",
	country: "Egypt",
	isDefault: true,
};

export function Checkout() {
	const t = useTranslations("checkout");
	const tCart = useTranslations("cart");
	const tAddresses = useTranslations("addresses");
	const { user } = useAuthContext();
	const { cartItems, refreshCart, resetCart, appliedCoupon, applyCoupon, removeCoupon } = useCart();
	const { checkout, loading: placingOrder } = useCheckout();
	const [couponInput, setCouponInput] = useState("");
	const [applyingCoupon, setApplyingCoupon] = useState(false);
	const { addresses, addAddress, refreshAddresses, loading: addressesLoading } = useAddresses();
	const { cards } = usePaymentMethods();
	const router = useRouter();

	// Step management: 1: Delivery Address, 2: Delivery Speed, 3: Payment & Confirm
	const [currentStep, setCurrentStep] = useState(1);

	// Address states
	const [selectedAddressId, setSelectedAddressId] = useState("");
	const [isAddingAddress, setIsAddingAddress] = useState(false);
	const [savingAddress, setSavingAddress] = useState(false);
	const [addressForm, setAddressForm] = useState(INITIAL_NEW_ADDRESS);

	const [shippingSpeed, setShippingSpeed] = useState("standard"); // 'standard' | 'express'
	const [deliveryNotes, setDeliveryNotes] = useState("");
	const [paymentType, setPaymentType] = useState("cod"); // default to COD
	const [selectedCardId, setSelectedCardId] = useState("");

	const [newCardData, setNewCardData] = useState({
		name: "",
		number: "",
		expiry: "",
		cvc: "",
	});

	// Order confirmation modal state
	const [completedOrder, setCompletedOrder] = useState<any>(null);

	// Derived selection IDs without setState in effect
	const effectiveAddressId = selectedAddressId || addresses?.find((a: any) => a.isDefault)?.id || addresses?.[0]?.id || "";
	const effectiveCardId = selectedCardId || cards?.find((c: any) => c.isDefault)?.id || cards?.[0]?.id || "";

	// Calculations using standard cartCalculations utility
	const {
		subtotal,
		discount,
		isFreeShipping: isStandardFree,
	} = calculateCartTotals(cartItems, appliedCoupon);

	const shippingCost = shippingSpeed === "express" ? 15.00 : (isStandardFree ? 0 : 9.99);
	const finalTotal = Math.max(0, subtotal - discount + shippingCost);

	const resolvedAddressString = resolveShippingAddress(addresses, effectiveAddressId);

	const handleOpenAddAddress = () => {
		setAddressForm({
			...INITIAL_NEW_ADDRESS,
			fullName: user?.name || "",
			phone: user?.phone || "",
			isDefault: !addresses || addresses.length === 0,
		});
		setIsAddingAddress(true);
	};

	const handleSaveNewAddress = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!addressForm.fullName.trim() || !addressForm.line1.trim() || !addressForm.city.trim() || !addressForm.country.trim()) {
			toast.error(t("validation.fillRequired"));
			return;
		}

		setSavingAddress(true);
		try {
			const res = await addAddress(addressForm);
			if (res.success) {
				toast.success(tAddresses("addedSuccess"));
				if (res.data?.id) {
					setSelectedAddressId(res.data.id);
				}
				setIsAddingAddress(false);
				if (refreshAddresses) await refreshAddresses();
			} else {
				toast.error(res.error || tAddresses("failedToAdd"));
			}
		} catch (err: any) {
			toast.error(err?.message || tAddresses("failedToAdd"));
		} finally {
			setSavingAddress(false);
		}
	};

	const handleApplyCoupon = async (codeToApply?: string) => {
		const targetCode = (codeToApply || couponInput).trim();
		if (!targetCode) {
			toast.error(t("validation.enterCoupon"));
			return;
		}
		setApplyingCoupon(true);
		try {
			const res = await applyCoupon(targetCode);
			if (res.success) {
				toast.success(res.coupon?.message || t("validation.couponApplied", { code: targetCode.toUpperCase() }));
				setCouponInput("");
			} else {
				toast.error(res.error || t("validation.couponFailed"));
			}
		} catch (err: any) {
			toast.error(err?.message || t("validation.invalidCoupon"));
		} finally {
			setApplyingCoupon(false);
		}
	};

	const handleRemoveCoupon = () => {
		removeCoupon();
		toast.info(t("validation.couponRemoved"));
	};

	// Validation helpers
	const validateStep1 = () => {
		if (!cartItems || cartItems.length === 0) {
			toast.error(tCart("empty.title"));
			router.push("/cart");
			return false;
		}

		if (!addresses || addresses.length === 0) {
			toast.error(t("validation.addAddress"));
			return false;
		}

		if (!effectiveAddressId) {
			toast.error(t("validation.selectAddress"));
			return false;
		}

		return true;
	};

	const handleGoToShipping = (e) => {
		if (e) e.preventDefault();
		if (validateStep1()) {
			setCurrentStep(2);
			window.scrollTo({ top: 0, behavior: "smooth" });
		}
	};

	const handleGoToPayment = (e) => {
		if (e) e.preventDefault();
		setCurrentStep(3);
		window.scrollTo({ top: 0, behavior: "smooth" });
	};

	const handlePlaceOrder = async (e) => {
		if (e) e.preventDefault();

		if (!validateStep1()) {
			setCurrentStep(1);
			return;
		}

		if (paymentType === "new-card") {
			if (!newCardData.name.trim()) {
				toast.error(t("validation.cardName"));
				return;
			}
			if (!newCardData.number.trim() || newCardData.number.replace(/\s/g, "").length < 13) {
				toast.error(t("validation.cardNumber"));
				return;
			}
			if (!newCardData.expiry.trim()) {
				toast.error(t("validation.cardExpiry"));
				return;
			}
			if (!newCardData.cvc.trim() || newCardData.cvc.length < 3) {
				toast.error(t("validation.cardCvc"));
				return;
			}
		} else if (paymentType === "saved-card" && (!cards || cards.length === 0)) {
			setPaymentType("new-card");
			toast.error(t("validation.cardDetails"));
			return;
		}

		const finalShippingAddress = resolvedAddressString;

		const orderPayload = {
			shippingAddressId: effectiveAddressId,
			couponId: appliedCoupon?.id || undefined,
			paymentMethod: paymentType === "cod" ? "cash_on_delivery" : paymentType,
		};

		const result = await checkout(orderPayload);

		if (result.success) {
			if (resetCart) resetCart();
			if (refreshCart) await refreshCart();
			const orderData = result.data || {};
			const randomOrderCode = generateOrderCode();
			const completedOrderInfo = {
				id: orderData.orderId || `ORD-${randomOrderCode}`,
				trackingNumber: `PRM-${orderData.orderId ? orderData.orderId.slice(0, 8).toUpperCase() : randomOrderCode}`,
				total: `EGP ${finalTotal.toFixed(2)}`,
				shippingAddress: finalShippingAddress,
				date: t("success.today"),
			};
			setCompletedOrder(completedOrderInfo);
			toast.success(t("validation.orderConfirmed"));
			window.scrollTo({ top: 0, behavior: "smooth" });
		} else {
			const errorMsg = typeof result.error === "string" ? result.error : result.error?.message || t("validation.orderFailed");
			toast.error(errorMsg);
		}
	};

	if (completedOrder) {
		return (
			<div className="max-w-2xl mx-auto py-12 px-4">
				<div className="bg-card border border-border rounded-3xl p-8 sm:p-12 text-center shadow-lg flex flex-col items-center">
						<div className="w-20 h-20 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center mb-6 shadow-inner">
							<PackageCheck className="size-10" />
						</div>

						<span className="px-3.5 py-1 rounded-full bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 text-xs font-semibold uppercase tracking-wider mb-3">
							{t("success.badge")}
						</span>
						<h1 className="text-3xl sm:text-4xl font-bold text-foreground mb-3">
							{t("success.title")}
						</h1>
						<p className="text-muted-foreground text-sm sm:text-base max-w-md mb-8">
							{t("success.desc")}
						</p>

						<div className="w-full bg-secondary/40 border border-border/80 rounded-2xl p-6 text-left mb-8 flex flex-col gap-4">
							<div className="flex justify-between items-center pb-3 border-b border-border/50 text-sm">
								<span className="text-muted-foreground font-medium">{t("success.orderNumber")}</span>
								<span className="font-mono font-bold text-foreground">{completedOrder.id}</span>
							</div>
							<div className="flex justify-between items-center pb-3 border-b border-border/50 text-sm">
								<span className="text-muted-foreground font-medium">{t("success.trackingNumber")}</span>
								<span className="font-mono font-semibold text-accent-brand">{completedOrder.trackingNumber || "PRM-829104"}</span>
							</div>
							<div className="flex justify-between items-center pb-3 border-b border-border/50 text-sm">
								<span className="text-muted-foreground font-medium">{t("success.totalPaid")}</span>
								<span className="font-bold text-foreground text-base">{completedOrder.total}</span>
							</div>
							<div className="flex flex-col gap-1 text-sm pt-1">
								<span className="text-muted-foreground font-medium">{t("success.destination")}</span>
								<span className="text-foreground text-xs sm:text-sm font-medium">{completedOrder.shippingAddress}</span>
							</div>
						</div>

						<div className="flex flex-col sm:flex-row gap-4 w-full justify-center">
							<button
								onClick={() => router.push("/user/orders")}
								className="flex-1 py-3.5 px-6 rounded-xl bg-primary text-primary-foreground font-semibold text-sm hover:bg-primary/90 transition-colors shadow-sm flex items-center justify-center gap-2"
							>
								<span>{t("success.trackOrders")}</span>
								<ArrowRight className="size-4" />
							</button>
							<button
								onClick={() => router.push("/products")}
								className="flex-1 py-3.5 px-6 rounded-xl bg-secondary text-foreground font-semibold text-sm hover:bg-secondary/80 border border-border transition-colors flex items-center justify-center"
							>
								{t("success.continueShopping")}
							</button>
						</div>
					</div>
				</div>
		);
	}

	if (!cartItems || cartItems.length === 0) {
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
							<ArrowRight className="size-4" />
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
				
				{/* Step Progress Bar */}
				<div className="mb-10">
					<div className="max-w-xl mx-auto">
						<nav aria-label={t("steps.progressAria")}>
							<ol className="flex items-center justify-between relative">
								{/* Connecting background line */}
								<div className="absolute left-0 top-1/2 -translate-y-1/2 w-full h-0.5 bg-border -z-10" />
								{/* Active progress colored line */}
								<div
									className="absolute left-0 top-1/2 -translate-y-1/2 h-0.5 bg-accent-brand transition-all duration-300 -z-10"
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

				<div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
					{/* Left Column: Multi-step Active View */}
					<div className="lg:col-span-7 flex flex-col gap-6">
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
										<form onSubmit={handleSaveNewAddress} className="space-y-4">
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
															onClick={() => setAddressForm({ ...addressForm, label: lbl })}
															className={`py-2 px-3 rounded-xl border text-xs font-medium cursor-pointer transition-all ${
																addressForm.label.toLowerCase() === lbl.toLowerCase()
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
												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">{tAddresses("labels.fullName")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("address.recipientPlaceholder")}
														value={addressForm.fullName}
														onChange={(e) => setAddressForm({ ...addressForm, fullName: e.target.value })}
														required
													/>
												</div>

												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">{t("address.street")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("address.streetPlaceholder")}
														value={addressForm.line1}
														onChange={(e) => setAddressForm({ ...addressForm, line1: e.target.value })}
														required
													/>
												</div>

												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">{t("address.apartment")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("address.apartmentPlaceholder")}
														value={addressForm.line2}
														onChange={(e) => setAddressForm({ ...addressForm, line2: e.target.value })}
													/>
												</div>

												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">{tAddresses("labels.city")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={tAddresses("placeholders.city")}
														value={addressForm.city}
														onChange={(e) => setAddressForm({ ...addressForm, city: e.target.value })}
														required
													/>
												</div>

												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">{t("address.state")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={tAddresses("placeholders.state")}
														value={addressForm.state}
														onChange={(e) => setAddressForm({ ...addressForm, state: e.target.value })}
													/>
												</div>

												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">{tAddresses("labels.postalCode")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={tAddresses("placeholders.postalCode")}
														value={addressForm.postalCode}
														onChange={(e) => setAddressForm({ ...addressForm, postalCode: e.target.value })}
													/>
												</div>

												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">{t("address.phone")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("address.phonePlaceholder")}
														value={addressForm.phone}
														onChange={(e) => setAddressForm({ ...addressForm, phone: e.target.value })}
													/>
												</div>

												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">{tAddresses("labels.country")}</label>
													<CountrySelect
														value={addressForm.country}
														onChange={(countryName) =>
															setAddressForm({ ...addressForm, country: countryName })
														}
													/>
												</div>
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
									) : addresses && addresses.length > 0 ? (
										<div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
											{addresses.map((addr) => {
												const isSelected = effectiveAddressId === addr.id;
												return (
													<div
														key={addr.id}
														onClick={() => setSelectedAddressId(addr.id)}
														className={`p-4 rounded-xl border cursor-pointer transition-all flex flex-col justify-between ${
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
															<p className="text-xs font-semibold text-foreground mb-1">{addr.fullName || addr.name}</p>
															<p className="text-xs text-muted-foreground leading-relaxed">
																{addr.line1 || addr.street}
																{addr.line2 ? `, ${addr.line2}` : ""}
															</p>
															<p className="text-xs text-muted-foreground">{addr.city}{addr.state ? `, ${addr.state}` : ""}, {addr.country}</p>
														</div>
														{addr.phone && (
															<span className="text-[11px] text-muted-foreground mt-3 font-medium">{addr.phone}</span>
														)}
													</div>
												);
											})}

											{/* Add Address Card Option */}
											<div
												onClick={handleOpenAddAddress}
												className="p-5 rounded-xl border border-dashed border-border hover:border-accent-brand/60 bg-secondary/10 hover:bg-accent-brand/5 cursor-pointer transition-all flex flex-col items-center justify-center text-center gap-2 min-h-[140px]"
											>
												<div className="size-8 rounded-full bg-accent-brand/10 text-accent-brand flex items-center justify-center">
													<Plus className="size-4" />
												</div>
												<span className="text-xs font-semibold text-foreground">{t("address.addAnother")}</span>
												<span className="text-[11px] text-muted-foreground">{t("address.addAnotherDesc")}</span>
											</div>
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
									<div className="flex items-center justify-between pt-6 border-t border-border mt-8">
										<button
											type="button"
											onClick={() => router.push("/cart")}
											className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors px-3 py-2 rounded-xl hover:bg-secondary"
										>
											<ArrowLeft className="size-4" />
											<span>{t("actions.returnToCart")}</span>
										</button>
										<button
											type="button"
											onClick={handleGoToShipping}
											className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-semibold text-sm rounded-xl shadow-sm hover:bg-accent-brand/90 transition-all"
										>
											<span>{t("actions.continueToDelivery")}</span>
											<ArrowRight className="size-4" />
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
										<div className="flex items-start gap-3">
											<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
												<MapPin className="size-4 text-accent-brand" />
											</div>
											<div>
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
											<div
												onClick={() => setShippingSpeed("standard")}
												className={`p-5 rounded-2xl border cursor-pointer transition-all flex flex-col justify-between ${
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
															{isStandardFree ? tCart("orderSummary.free") : "EGP 9.99"}
														</span>
													</div>
													<p className="text-xs text-muted-foreground leading-relaxed">
														{t("delivery.standardDesc")}
													</p>
												</div>
												{shippingSpeed === "standard" && (
													<div className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand mt-4 pt-3 border-t border-accent-brand/20">
														<CheckCircle2 className="size-4" />
														<span>{t("delivery.selected")}</span>
													</div>
												)}
											</div>

											<div
												onClick={() => setShippingSpeed("express")}
												className={`p-5 rounded-2xl border cursor-pointer transition-all flex flex-col justify-between ${
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
														<span className="text-xs font-extrabold text-foreground">EGP 15.00</span>
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
											</div>
										</div>

										{/* Optional Delivery Instructions */}
										<div className="space-y-2 pt-4 border-t border-border/60">
											<label className="text-xs font-semibold text-foreground">{t("delivery.instructions")}</label>
											<input
												type="text"
												placeholder={t("delivery.instructionsPlaceholder")}
												value={deliveryNotes}
												onChange={(e) => setDeliveryNotes(e.target.value)}
												className="w-full px-3.5 py-2.5 bg-background border border-border rounded-xl text-xs text-foreground focus:outline-none focus:ring-2 focus:ring-accent-brand"
											/>
										</div>

										{/* Step 2 Actions */}
										<div className="flex items-center justify-between pt-6 border-t border-border mt-8">
											<button
												type="button"
												onClick={() => setCurrentStep(1)}
												className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors px-3 py-2 rounded-xl hover:bg-secondary"
											>
												<ArrowLeft className="size-4" />
												<span>{t("actions.backToAddress")}</span>
											</button>
											<button
												type="button"
												onClick={handleGoToPayment}
												className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-semibold text-sm rounded-xl shadow-sm hover:bg-accent-brand/90 transition-all"
											>
												<span>{t("actions.continueToPayment")}</span>
												<ArrowRight className="size-4" />
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
											<div className="flex items-start gap-3">
												<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
													<MapPin className="size-4 text-accent-brand" />
												</div>
												<div>
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
											<div className="flex items-start gap-3">
												<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
													<Truck className="size-4 text-accent-brand" />
												</div>
												<div>
													<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{t("steps.shipping")}</span>
													<p className="text-xs sm:text-sm font-semibold text-foreground">
														{shippingSpeed === "express" ? t("delivery.expressRecap") : t("delivery.standardRecap", { price: isStandardFree ? tCart("orderSummary.free") : "EGP 9.99" })}
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

										{/* Payment Tabs */}
										<div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 mb-6">
											{cards.length > 0 && (
												<button
													type="button"
													onClick={() => setPaymentType("saved-card")}
													className={`p-3 rounded-xl border text-xs font-medium flex flex-col items-center gap-1.5 transition-all ${
														paymentType === "saved-card"
															? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold ring-1 ring-accent-brand"
															: "border-border hover:bg-secondary text-muted-foreground hover:text-foreground"
													}`}
												>
													<CreditCard className="size-4" />
													<span>{t("payment.savedCard")}</span>
												</button>
											)}
											<button
												type="button"
												onClick={() => setPaymentType("new-card")}
												className={`p-3 rounded-xl border text-xs font-medium flex flex-col items-center gap-1.5 transition-all ${
													paymentType === "new-card"
														? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold ring-1 ring-accent-brand"
														: "border-border hover:bg-secondary text-muted-foreground hover:text-foreground"
												}`}
											>
												<CreditCard className="size-4" />
												<span>{t("payment.creditCard")}</span>
											</button>
											<button
												type="button"
												onClick={() => setPaymentType("apple-pay")}
												className={`p-3 rounded-xl border text-xs font-medium flex flex-col items-center gap-1.5 transition-all ${
													paymentType === "apple-pay"
														? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold ring-1 ring-accent-brand"
														: "border-border hover:bg-secondary text-muted-foreground hover:text-foreground"
												}`}
											>
												<Smartphone className="size-4" />
												<span>{t("payment.digital")}</span>
											</button>
											<button
												type="button"
												onClick={() => setPaymentType("cod")}
												className={`p-3 rounded-xl border text-xs font-medium flex flex-col items-center gap-1.5 transition-all ${
													paymentType === "cod"
														? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold ring-1 ring-accent-brand"
														: "border-border hover:bg-secondary text-muted-foreground hover:text-foreground"
												}`}
											>
												<Banknote className="size-4" />
												<span>{t("payment.cod")}</span>
											</button>
										</div>

										{/* Payment method content */}
										{paymentType === "saved-card" && cards.length > 0 && (
											<div className="flex flex-col gap-3">
												{cards.map((c) => (
													<div
														key={c.id}
														onClick={() => setSelectedCardId(c.id)}
														className={`p-4 rounded-xl border cursor-pointer flex items-center justify-between transition-all ${
															effectiveCardId === c.id
																? "border-accent-brand bg-accent-brand/5 ring-1 ring-accent-brand"
																: "border-border hover:border-border/80 bg-background"
														}`}
													>
														<div className="flex items-center gap-3">
															<div className="p-2 rounded-lg bg-secondary">
																<CreditCard className="size-4 text-foreground" />
															</div>
															<div>
																<p className="text-sm font-semibold text-foreground">{t("payment.cardEnding", { type: c.type, last4: c.last4 })}</p>
																<p className="text-xs text-muted-foreground">{t("payment.expires", { expiry: c.expiry, name: c.name })}</p>
															</div>
														</div>
														{effectiveCardId === c.id && (
															<CheckCircle2 className="size-4 text-accent-brand" />
														)}
													</div>
												))}
											</div>
										)}

										{paymentType === "new-card" && (
											<div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">{t("payment.nameOnCard")}</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("payment.namePlaceholder")}
														value={newCardData.name}
														onChange={(e) => setNewCardData({ ...newCardData, name: e.target.value })}
														required
													/>
												</div>
												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">{t("payment.cardNumber")}</label>
													<input
														type="text"
														maxLength={19}
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("payment.cardNumberPlaceholder")}
														value={newCardData.number}
														onChange={(e) => setNewCardData({ ...newCardData, number: e.target.value })}
														required
													/>
												</div>
												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">{t("payment.expiry")}</label>
													<input
														type="text"
														maxLength={5}
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("payment.expiryPlaceholder")}
														value={newCardData.expiry}
														onChange={(e) => setNewCardData({ ...newCardData, expiry: e.target.value })}
														required
													/>
												</div>
												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">{t("payment.cvc")}</label>
													<input
														type="password"
														maxLength={4}
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder={t("payment.cvcPlaceholder")}
														value={newCardData.cvc}
														onChange={(e) => setNewCardData({ ...newCardData, cvc: e.target.value })}
														required
													/>
												</div>
											</div>
										)}

										{paymentType === "apple-pay" && (
											<div className="p-6 rounded-xl bg-secondary/30 border border-border text-center flex flex-col items-center">
												<div className="w-12 h-12 rounded-full bg-foreground text-background flex items-center justify-center mb-3">
													<Smartphone className="size-6" />
												</div>
												<h4 className="font-semibold text-foreground text-sm mb-1">{t("payment.digitalTitle")}</h4>
												<p className="text-xs text-muted-foreground max-w-sm">
													{t("payment.digitalDesc")}
												</p>
											</div>
										)}

										{paymentType === "cod" && (
											<div className="p-5 rounded-xl bg-secondary/30 border border-border flex items-start gap-3">
												<Banknote className="size-5 text-accent-brand shrink-0 mt-0.5" />
												<div className="flex flex-col gap-1 text-xs">
													<span className="font-semibold text-foreground">{t("payment.codTitle")}</span>
													<span className="text-muted-foreground leading-relaxed">
														{t("payment.codDesc")}
													</span>
												</div>
											</div>
										)}

										{/* Step 3 Actions */}
										<div className="flex items-center justify-between pt-6 border-t border-border mt-8">
											<button
												type="button"
												onClick={() => setCurrentStep(2)}
												className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors px-3 py-2 rounded-xl hover:bg-secondary"
											>
												<ArrowLeft className="size-4" />
												<span>{t("actions.backToDelivery")}</span>
											</button>
											<button
												type="button"
												onClick={handlePlaceOrder}
												disabled={placingOrder}
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
					<div className="lg:col-span-5 sticky top-24">
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
										<div className="flex items-center gap-3">
											<div className="w-12 h-12 rounded-lg bg-secondary overflow-hidden shrink-0 border border-border/50">
												<img
													src={item.img || "/placeholder-product.png"}
													alt={(typeof item.productName === "object" ? item.productName?.en : item.productName) || t("summary.product")}
													className="w-full h-full object-cover"
												/>
											</div>
											<div className="flex flex-col">
												<span className="font-medium text-foreground line-clamp-1">{(typeof item.productName === "object" ? item.productName?.en : item.productName) || t("summary.product")}</span>
												<span className="text-xs text-muted-foreground">{t("summary.qty", { count: item.quantity || 1 })}</span>
											</div>
										</div>
										<span className="font-semibold text-foreground">{item.productPrice}</span>
									</div>
								))}
							</div>

							{/* Clean Promo Code Input */}
							<div className="pt-2 border-t border-border/60">
								{appliedCoupon ? (
									<div className="flex items-center justify-between p-2.5 bg-emerald-500/10 border border-emerald-500/25 rounded-xl text-xs">
										<div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium">
											<Tag className="size-3.5 text-emerald-600 shrink-0" />
											<span className="font-bold tracking-wider font-mono uppercase">{appliedCoupon.code}</span>
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
											className="p-1 rounded-md text-muted-foreground hover:text-destructive hover:bg-background/80 transition-colors"
											title={t("summary.removePromo")}
										>
											<X className="size-3.5" />
										</button>
									</div>
								) : (
									<div className="flex gap-2">
										<div className="flex-1 bg-secondary/30 border border-border focus-within:border-accent-brand rounded-xl overflow-hidden flex items-center px-3 h-9 transition-colors">
											<Tag className="size-3.5 text-muted-foreground shrink-0 mr-2" />
											<input
												type="text"
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
								)}
							</div>

							<div className="flex flex-col gap-2.5 pt-4 border-t border-border/60 text-sm">
								<div className="flex justify-between text-muted-foreground">
									<span>{tCart("orderSummary.subtotal")}</span>
									<span className="font-medium text-foreground">EGP {subtotal.toFixed(2)}</span>
								</div>

								{discount > 0 && (
									<div className="flex justify-between text-emerald-600 dark:text-emerald-400 font-medium">
										<span className="flex items-center gap-1">
											<Sparkles className="size-3.5" />
											{t("summary.discount", { code: appliedCoupon?.code })}
										</span>
										<span>-EGP {discount.toFixed(2)}</span>
									</div>
								)}

								<div className="flex justify-between text-muted-foreground">
									<span>{t("summary.shipping", { speed: shippingSpeed === "express" ? t("summary.express") : t("summary.standard") })}</span>
									<span className="font-medium text-foreground">
										{shippingCost === 0 ? (
											<span className="text-emerald-600 font-medium">{tCart("orderSummary.free")}</span>
										) : (
											`EGP ${shippingCost.toFixed(2)}`
										)}
									</span>
								</div>

								<div className="flex justify-between text-muted-foreground">
									<span>{t("summary.taxes")}</span>
									<span className="font-medium text-foreground">EGP 0.00</span>
								</div>

								<div className="flex justify-between items-center text-lg font-bold text-foreground pt-3 border-t border-border">
									<span>{t("summary.total")}</span>
									<span className="text-2xl font-extrabold text-foreground">EGP {finalTotal.toFixed(2)}</span>
								</div>
							</div>

							{/* Context-aware primary button */}
							{currentStep === 1 && (
								<button
									type="button"
									onClick={handleGoToShipping}
									className="w-full py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm shadow-sm transition-all flex items-center justify-center gap-2"
								>
									<span>{t("actions.continueToDelivery")}</span>
									<ArrowRight className="size-4" />
								</button>
							)}

							{currentStep === 2 && (
								<button
									type="button"
									onClick={handleGoToPayment}
									className="w-full py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm shadow-sm transition-all flex items-center justify-center gap-2"
								>
									<span>{t("actions.continueToPayment")}</span>
									<ArrowRight className="size-4" />
								</button>
							)}

							{currentStep === 3 && (
								<button
									type="button"
									onClick={handlePlaceOrder}
									disabled={placingOrder || !cartItems || cartItems.length === 0}
									className="w-full py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-bold text-base shadow-sm transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
								>
									<span>{placingOrder ? t("actions.processingOrder") : t("actions.placeOrder", { total: finalTotal.toFixed(2) })}</span>
									<CheckCircle2 className="size-4" />
								</button>
							)}

							<div className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
								<ShieldCheck className="size-4 text-emerald-600" />
								<span>{t("summary.secure")}</span>
							</div>
						</div>
					</div>
			</div>
		</div>
	);
}
