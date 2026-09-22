"use client";

import { useState } from "react";
import { useCart } from "@/hooks/useCart";
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
	Pencil
} from "lucide-react";
import { toast } from "sonner";
import { motion, AnimatePresence } from "motion/react";

function generateOrderCode() {
	return Math.floor(10000 + Math.random() * 90000).toString();
}

function resolveShippingAddress(isCustom, custom, addresses, effectiveId) {
	if (isCustom || !addresses || addresses.length === 0) {
		return `${custom.name ? custom.name + " - " : ""}${custom.street}, ${custom.city}`;
	}
	const addr = addresses.find((a) => a.id === effectiveId);
	return addr ? `${addr.name} - ${addr.street}, ${addr.city}, ${addr.country}` : "";
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

export function Checkout() {
	const { cartItems, refreshCart, appliedCoupon } = useCart();
	const { checkout, loading: placingOrder } = useCheckout();
	const { addresses } = useAddresses();
	const { cards } = usePaymentMethods();
	const router = useRouter();

	// Step management: 1: Delivery Address, 2: Delivery Speed, 3: Payment & Confirm
	const [currentStep, setCurrentStep] = useState(1);

	// Selection states
	const [selectedAddressId, setSelectedAddressId] = useState("");
	const [isCustomAddress, setIsCustomAddress] = useState(false);
	const [customAddress, setCustomAddress] = useState({
		name: "",
		street: "",
		city: "",
		phone: "",
	});

	const [shippingSpeed, setShippingSpeed] = useState("standard"); // 'standard' | 'express'
	const [deliveryNotes, setDeliveryNotes] = useState("");
	const [paymentType, setPaymentType] = useState("saved-card"); // 'saved-card' | 'new-card' | 'cod' | 'apple-pay'
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

	// Calculations
	const subtotal = cartItems?.reduce((acc: number, item: any) => acc + (parseFloat(String(item.productPrice || "").replace('$', '') || "0") * (item.quantity || 1)), 0) || 0;
	
	let discount = 0;
	if (appliedCoupon?.discountPercent) {
		discount = (subtotal * appliedCoupon.discountPercent) / 100;
	} else if (appliedCoupon?.fixedDiscount) {
		discount = appliedCoupon.fixedDiscount;
	}

	const isStandardFree = subtotal >= 150 || appliedCoupon?.freeShipping;
	const shippingCost = shippingSpeed === "express" ? 15.00 : (isStandardFree ? 0 : 9.99);
	const finalTotal = Math.max(0, subtotal - discount + shippingCost);

	const resolvedAddressString = resolveShippingAddress(isCustomAddress, customAddress, addresses, effectiveAddressId);

	// Validation helpers
	const validateStep1 = () => {
		if (!cartItems || cartItems.length === 0) {
			toast.error("Your cart is empty");
			router.push("/cart");
			return false;
		}

		if (isCustomAddress || !addresses || addresses.length === 0) {
			if (!customAddress.name.trim()) {
				toast.error("Please enter recipient name");
				return false;
			}
			if (!customAddress.street.trim() || !customAddress.city.trim()) {
				toast.error("Please enter complete delivery address (street and city)");
				return false;
			}
		} else if (!effectiveAddressId) {
			toast.error("Please select a delivery address");
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
				toast.error("Please enter the name on your card");
				return;
			}
			if (!newCardData.number.trim() || newCardData.number.replace(/\s/g, "").length < 13) {
				toast.error("Please enter a valid card number");
				return;
			}
			if (!newCardData.expiry.trim()) {
				toast.error("Please enter card expiration date");
				return;
			}
			if (!newCardData.cvc.trim() || newCardData.cvc.length < 3) {
				toast.error("Please enter card security code (CVC)");
				return;
			}
		} else if (paymentType === "saved-card" && (!cards || cards.length === 0)) {
			setPaymentType("new-card");
			toast.error("Please enter your card details");
			return;
		}

		const finalShippingAddress = resolvedAddressString;
		const paymentMethodLabel = resolvePaymentLabel(paymentType, cards, effectiveCardId, newCardData);

		const orderPayload = {
			items: cartItems,
			address: finalShippingAddress,
			shippingMethod: shippingSpeed === "express" ? "Express Priority ($15.00)" : "Standard Delivery (Free)",
			shipping: shippingCost === 0 ? "Free" : `$${shippingCost.toFixed(2)}`,
			paymentMethod: paymentMethodLabel,
			deliveryNotes: deliveryNotes.trim() || undefined,
			subtotal: `$${subtotal.toFixed(2)}`,
			discount: discount > 0 ? `-$${discount.toFixed(2)}` : null,
			total: `$${finalTotal.toFixed(2)}`
		};

		const result = await checkout(orderPayload);

		if (result.success) {
			if (refreshCart) await refreshCart();
			const randomOrderCode = generateOrderCode();
			const fallbackOrder = {
				id: `ORD-${randomOrderCode}`,
				trackingNumber: `PRM-${randomOrderCode}`,
				total: `$${finalTotal.toFixed(2)}`,
				shippingAddress: finalShippingAddress,
				date: "Today",
			};
			setCompletedOrder(result.data || fallbackOrder);
			toast.success("Order confirmed successfully!");
			window.scrollTo({ top: 0, behavior: "smooth" });
		} else {
			toast.error(result.error || "Failed to place order");
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
							Order Placed Successfully
						</span>
						<h1 className="text-3xl sm:text-4xl font-bold text-foreground mb-3">
							Thank You for Your Order!
						</h1>
						<p className="text-muted-foreground text-sm sm:text-base max-w-md mb-8">
							We've received your order and are preparing it for shipment. A confirmation receipt has been sent to your email.
						</p>

						<div className="w-full bg-secondary/40 border border-border/80 rounded-2xl p-6 text-left mb-8 flex flex-col gap-4">
							<div className="flex justify-between items-center pb-3 border-b border-border/50 text-sm">
								<span className="text-muted-foreground font-medium">Order Number:</span>
								<span className="font-mono font-bold text-foreground">{completedOrder.id}</span>
							</div>
							<div className="flex justify-between items-center pb-3 border-b border-border/50 text-sm">
								<span className="text-muted-foreground font-medium">Tracking Number:</span>
								<span className="font-mono font-semibold text-accent-brand">{completedOrder.trackingNumber || "PRM-829104"}</span>
							</div>
							<div className="flex justify-between items-center pb-3 border-b border-border/50 text-sm">
								<span className="text-muted-foreground font-medium">Total Paid:</span>
								<span className="font-bold text-foreground text-base">{completedOrder.total}</span>
							</div>
							<div className="flex flex-col gap-1 text-sm pt-1">
								<span className="text-muted-foreground font-medium">Delivery Destination:</span>
								<span className="text-foreground text-xs sm:text-sm font-medium">{completedOrder.shippingAddress}</span>
							</div>
						</div>

						<div className="flex flex-col sm:flex-row gap-4 w-full justify-center">
							<button
								onClick={() => router.push("/user/orders")}
								className="flex-1 py-3.5 px-6 rounded-xl bg-primary text-primary-foreground font-semibold text-sm hover:bg-primary/90 transition-colors shadow-sm flex items-center justify-center gap-2"
							>
								<span>Track in My Orders</span>
								<ArrowRight className="size-4" />
							</button>
							<button
								onClick={() => router.push("/products")}
								className="flex-1 py-3.5 px-6 rounded-xl bg-secondary text-foreground font-semibold text-sm hover:bg-secondary/80 border border-border transition-colors flex items-center justify-center"
							>
								Continue Shopping
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
						<h2 className="text-2xl font-bold text-foreground mb-2">Your cart is empty</h2>
						<p className="text-muted-foreground text-sm max-w-sm mb-6">
							You need to add items to your cart before proceeding to checkout.
						</p>
						<button
							onClick={() => router.push("/products")}
							className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-primary text-primary-foreground font-medium text-sm hover:bg-primary/90 transition-colors shadow-sm"
						>
							<span>Discover Products</span>
							<ArrowRight className="size-4" />
						</button>
					</div>
				</div>
		);
	}

	const stepsConfig = [
		{ id: 1, label: "Delivery Address", shortLabel: "Address", icon: MapPin },
		{ id: 2, label: "Delivery Speed", shortLabel: "Shipping", icon: Truck },
		{ id: 3, label: "Payment & Confirm", shortLabel: "Payment", icon: CreditCard },
	];

	return (
		<div className="max-w-6xl mx-auto py-8 sm:py-12 px-4">
				
				{/* Step Progress Bar */}
				<div className="mb-10">
					<div className="max-w-xl mx-auto">
						<nav aria-label="Checkout Progress">
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
												<h2 className="text-lg font-bold text-foreground">1. Delivery Address</h2>
												<p className="text-xs text-muted-foreground">Select your saved destination or enter a new address.</p>
											</div>
										</div>
										{addresses.length > 0 && (
											<button
												type="button"
												onClick={() => setIsCustomAddress(!isCustomAddress)}
												className="text-xs font-semibold text-accent-brand hover:underline px-3 py-1.5 rounded-lg hover:bg-accent-brand/5 transition-colors"
											>
												{isCustomAddress ? "Use saved address" : "+ Enter new address"}
											</button>
										)}
									</div>

									{!isCustomAddress && addresses.length > 0 ? (
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
																<span className="font-semibold text-foreground text-sm flex items-center gap-1.5">
																	<MapPin className="size-3.5 text-accent-brand" />
																	{addr.type}
																</span>
																{isSelected ? (
																	<span className="size-2.5 rounded-full bg-accent-brand"></span>
																) : (
																	addr.isDefault && (
																		<span className="text-[10px] font-medium text-muted-foreground bg-secondary px-2 py-0.5 rounded-md">Default</span>
																	)
																)}
															</div>
															<p className="text-xs font-semibold text-foreground mb-1">{addr.name}</p>
															<p className="text-xs text-muted-foreground leading-relaxed">{addr.street}</p>
															<p className="text-xs text-muted-foreground">{addr.city}, {addr.country}</p>
														</div>
														<span className="text-[11px] text-muted-foreground mt-3 font-medium">{addr.phone}</span>
													</div>
												);
											})}
										</div>
									) : (
										<div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
											<div className="flex flex-col gap-1.5 sm:col-span-2">
												<label className="text-xs font-medium text-foreground">Recipient Full Name *</label>
												<input
													type="text"
													className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
													placeholder="e.g. Mohamed Mahmoud"
													value={customAddress.name}
													onChange={(e) => setCustomAddress({ ...customAddress, name: e.target.value })}
													required
												/>
											</div>
											<div className="flex flex-col gap-1.5 sm:col-span-2">
												<label className="text-xs font-medium text-foreground">Street Address / Apartment *</label>
												<input
													type="text"
													className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
													placeholder="Street name, building number, suite/apt"
													value={customAddress.street}
													onChange={(e) => setCustomAddress({ ...customAddress, street: e.target.value })}
													required
												/>
											</div>
											<div className="flex flex-col gap-1.5">
												<label className="text-xs font-medium text-foreground">City *</label>
												<input
													type="text"
													className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
													placeholder="Cairo"
													value={customAddress.city}
													onChange={(e) => setCustomAddress({ ...customAddress, city: e.target.value })}
													required
												/>
											</div>
											<div className="flex flex-col gap-1.5">
												<label className="text-xs font-medium text-foreground">Phone Number</label>
												<input
													type="text"
													className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
													placeholder="+20 100 123 4567"
													value={customAddress.phone}
													onChange={(e) => setCustomAddress({ ...customAddress, phone: e.target.value })}
												/>
											</div>
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
											<span>Return to Cart</span>
										</button>
										<button
											type="button"
											onClick={handleGoToShipping}
											className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-semibold text-sm rounded-xl shadow-sm hover:bg-accent-brand/90 transition-all"
										>
											<span>Continue to Delivery Speed</span>
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
												<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">Ship To</span>
												<p className="text-xs sm:text-sm font-semibold text-foreground line-clamp-1">{resolvedAddressString}</p>
											</div>
										</div>
										<button
											type="button"
											onClick={() => setCurrentStep(1)}
											className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-3 py-1.5 rounded-lg hover:bg-accent-brand/5 shrink-0"
										>
											<Pencil className="size-3.5" />
											<span>Change</span>
										</button>
									</div>

									{/* Main Delivery Speed Selector */}
									<div className="bg-card border border-border rounded-2xl p-6 sm:p-7 shadow-xs">
										<div className="flex items-center gap-3 pb-5 border-b border-border/60 mb-6">
											<div className="size-9 rounded-xl bg-accent-brand/10 text-accent-brand flex items-center justify-center font-bold">
												<Truck className="size-5" />
											</div>
											<div>
												<h2 className="text-lg font-bold text-foreground">2. Delivery Speed</h2>
												<p className="text-xs text-muted-foreground">Select your preferred courier fulfillment schedule.</p>
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
																<span className="font-bold text-sm text-foreground block">Standard Delivery</span>
																<span className="text-[11px] text-muted-foreground">3 to 5 business days</span>
															</div>
														</div>
														<span className="text-xs font-extrabold text-emerald-600 dark:text-emerald-400">
															{isStandardFree ? "FREE" : "$9.99"}
														</span>
													</div>
													<p className="text-xs text-muted-foreground leading-relaxed">
														Standard ground delivery via regular postal service. Free for orders over $150.
													</p>
												</div>
												{shippingSpeed === "standard" && (
													<div className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand mt-4 pt-3 border-t border-accent-brand/20">
														<CheckCircle2 className="size-4" />
														<span>Selected method</span>
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
																<span className="font-bold text-sm text-foreground block">Express Priority</span>
																<span className="text-[11px] text-muted-foreground">1 to 2 business days</span>
															</div>
														</div>
														<span className="text-xs font-extrabold text-foreground">$15.00</span>
													</div>
													<p className="text-xs text-muted-foreground leading-relaxed">
														Guaranteed air express courier with real-time GPS tracking and signature on delivery.
													</p>
												</div>
												{shippingSpeed === "express" && (
													<div className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand mt-4 pt-3 border-t border-accent-brand/20">
														<CheckCircle2 className="size-4" />
														<span>Selected method</span>
													</div>
												)}
											</div>
										</div>

										{/* Optional Delivery Instructions */}
										<div className="space-y-2 pt-4 border-t border-border/60">
											<label className="text-xs font-semibold text-foreground">Delivery Instructions (Optional)</label>
											<input
												type="text"
												placeholder="e.g. Leave package at front door, gate code #4492"
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
												<span>Back to Address</span>
											</button>
											<button
												type="button"
												onClick={handleGoToPayment}
												className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-semibold text-sm rounded-xl shadow-sm hover:bg-accent-brand/90 transition-all"
											>
												<span>Continue to Payment</span>
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
													<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">Ship To</span>
													<p className="text-xs sm:text-sm font-semibold text-foreground line-clamp-1">{resolvedAddressString}</p>
												</div>
											</div>
											<button
												type="button"
												onClick={() => setCurrentStep(1)}
												className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-2.5 py-1 rounded-lg hover:bg-accent-brand/5 shrink-0"
											>
												<Pencil className="size-3" />
												<span>Edit</span>
											</button>
										</div>

										<div className="flex items-center justify-between pt-3">
											<div className="flex items-start gap-3">
												<div className="p-2 rounded-xl bg-secondary text-foreground shrink-0 mt-0.5">
													<Truck className="size-4 text-accent-brand" />
												</div>
												<div>
													<span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">Delivery Speed</span>
													<p className="text-xs sm:text-sm font-semibold text-foreground">
														{shippingSpeed === "express" ? "Express Priority (1-2 days) • $15.00" : `Standard Delivery (3-5 days) • ${isStandardFree ? "Free" : "$9.99"}`}
													</p>
												</div>
											</div>
											<button
												type="button"
												onClick={() => setCurrentStep(2)}
												className="flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline px-2.5 py-1 rounded-lg hover:bg-accent-brand/5 shrink-0"
											>
												<Pencil className="size-3" />
												<span>Edit</span>
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
												<h2 className="text-lg font-bold text-foreground">3. Payment Method</h2>
												<p className="text-xs text-muted-foreground">All transactions are encrypted with 256-bit bank-grade security.</p>
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
													<span>Saved Card</span>
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
												<span>Credit Card</span>
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
												<span>Digital Pay</span>
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
												<span>Cash On Delivery</span>
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
																<p className="text-sm font-semibold text-foreground">{c.type} ending in {c.last4}</p>
																<p className="text-xs text-muted-foreground">Expires {c.expiry} • {c.name}</p>
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
													<label className="text-xs font-medium text-foreground">Name on Card *</label>
													<input
														type="text"
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder="Mohamed Mahmoud"
														value={newCardData.name}
														onChange={(e) => setNewCardData({ ...newCardData, name: e.target.value })}
														required
													/>
												</div>
												<div className="flex flex-col gap-1.5 sm:col-span-2">
													<label className="text-xs font-medium text-foreground">Card Number *</label>
													<input
														type="text"
														maxLength={19}
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder="4242 •••• •••• 4242"
														value={newCardData.number}
														onChange={(e) => setNewCardData({ ...newCardData, number: e.target.value })}
														required
													/>
												</div>
												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">Expiration Date *</label>
													<input
														type="text"
														maxLength={5}
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder="MM/YY"
														value={newCardData.expiry}
														onChange={(e) => setNewCardData({ ...newCardData, expiry: e.target.value })}
														required
													/>
												</div>
												<div className="flex flex-col gap-1.5">
													<label className="text-xs font-medium text-foreground">Security Code (CVC) *</label>
													<input
														type="password"
														maxLength={4}
														className="px-3.5 py-2.5 border border-border rounded-xl bg-background text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent-brand"
														placeholder="123"
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
												<h4 className="font-semibold text-foreground text-sm mb-1">Instant 1-Click Digital Payment</h4>
												<p className="text-xs text-muted-foreground max-w-sm">
													Authenticate with Touch ID, Face ID, or your biometric wallet when you confirm the order below.
												</p>
											</div>
										)}

										{paymentType === "cod" && (
											<div className="p-5 rounded-xl bg-secondary/30 border border-border flex items-start gap-3">
												<Banknote className="size-5 text-accent-brand shrink-0 mt-0.5" />
												<div className="flex flex-col gap-1 text-xs">
													<span className="font-semibold text-foreground">Cash on Delivery</span>
													<span className="text-muted-foreground leading-relaxed">
														Pay with cash directly to the courier upon delivery. Please ensure exact change is available at your address.
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
												<span>Back to Delivery</span>
											</button>
											<button
												type="button"
												onClick={handlePlaceOrder}
												disabled={placingOrder}
												className="flex items-center gap-2 px-6 py-3 bg-accent-brand text-white font-bold text-sm rounded-xl shadow-md hover:bg-accent-brand/90 transition-all disabled:opacity-50"
											>
												<span>{placingOrder ? "Placing Order..." : `Place Order • $${finalTotal.toFixed(2)}`}</span>
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
								<h3 className="font-semibold text-lg text-foreground">Order Summary</h3>
								<span className="text-xs font-medium text-muted-foreground bg-secondary px-2.5 py-1 rounded-full">
									{cartItems?.length || 0} {cartItems?.length === 1 ? "item" : "items"}
								</span>
							</div>

							<div className="flex flex-col gap-3.5 max-h-[240px] overflow-y-auto pr-1">
								{cartItems?.map((item) => (
									<div key={item.id} className="flex items-center justify-between gap-3 text-sm py-2 border-b border-border/40 last:border-0">
										<div className="flex items-center gap-3">
											<div className="w-12 h-12 rounded-lg bg-secondary overflow-hidden shrink-0 border border-border/50">
												<img
													src={item.img || "/placeholder-product.png"}
													alt={(typeof item.productName === "object" ? item.productName?.en : item.productName) || "Product"}
													className="w-full h-full object-cover"
												/>
											</div>
											<div className="flex flex-col">
												<span className="font-medium text-foreground line-clamp-1">{(typeof item.productName === "object" ? item.productName?.en : item.productName) || "Product"}</span>
												<span className="text-xs text-muted-foreground">Qty: {item.quantity || 1}</span>
											</div>
										</div>
										<span className="font-semibold text-foreground">{item.productPrice}</span>
									</div>
								))}
							</div>

							<div className="flex flex-col gap-2.5 pt-4 border-t border-border/60 text-sm">
								<div className="flex justify-between text-muted-foreground">
									<span>Subtotal</span>
									<span className="font-medium text-foreground">${subtotal.toFixed(2)}</span>
								</div>

								{discount > 0 && (
									<div className="flex justify-between text-emerald-600 dark:text-emerald-400 font-medium">
										<span className="flex items-center gap-1">
											<Sparkles className="size-3.5" />
											Discount ({appliedCoupon?.code})
										</span>
										<span>-${discount.toFixed(2)}</span>
									</div>
								)}

								<div className="flex justify-between text-muted-foreground">
									<span>Shipping ({shippingSpeed === "express" ? "Express" : "Standard"})</span>
									<span className="font-medium text-foreground">
										{shippingCost === 0 ? (
											<span className="text-emerald-600 font-medium">FREE</span>
										) : (
											`$${shippingCost.toFixed(2)}`
										)}
									</span>
								</div>

								<div className="flex justify-between text-muted-foreground">
									<span>Estimated Taxes</span>
									<span className="font-medium text-foreground">$0.00</span>
								</div>

								<div className="flex justify-between items-center text-lg font-bold text-foreground pt-3 border-t border-border">
									<span>Total to Pay</span>
									<span className="text-2xl font-extrabold text-foreground">${finalTotal.toFixed(2)}</span>
								</div>
							</div>

							{/* Context-aware primary button */}
							{currentStep === 1 && (
								<button
									type="button"
									onClick={handleGoToShipping}
									className="w-full py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm shadow-sm transition-all flex items-center justify-center gap-2"
								>
									<span>Continue to Delivery Speed</span>
									<ArrowRight className="size-4" />
								</button>
							)}

							{currentStep === 2 && (
								<button
									type="button"
									onClick={handleGoToPayment}
									className="w-full py-3.5 px-6 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm shadow-sm transition-all flex items-center justify-center gap-2"
								>
									<span>Continue to Payment</span>
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
									<span>{placingOrder ? "Processing Order..." : `Place Order • $${finalTotal.toFixed(2)}`}</span>
									<CheckCircle2 className="size-4" />
								</button>
							)}

							<div className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
								<ShieldCheck className="size-4 text-emerald-600" />
								<span>Guaranteed Safe & Encrypted Checkout</span>
							</div>
						</div>
					</div>
			</div>
		</div>
	);
}
