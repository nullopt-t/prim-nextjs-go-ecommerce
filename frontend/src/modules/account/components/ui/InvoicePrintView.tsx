"use client";

import { useLocale, useTranslations } from "next-intl";

interface InvoiceOrderItem {
	id: string;
	variantId: string;
	quantity: number;
	priceAtPurchase: number;
	productSnapshot?: string;
}

interface InvoiceOrder {
	id: string;
	customerEmail: string;
	shippingAddress?: {
		street?: string;
		city?: string;
		state?: string;
		postalCode?: string;
		country?: string;
	};
	status: string;
	totalAmount: number;
	discountAmount?: number;
	shippingAmount?: number;
	taxAmount?: number;
	codFee?: number;
	carrier?: string | null;
	trackingNumber?: string | null;
	couponCode?: string | null;
	currency?: string;
	items?: InvoiceOrderItem[];
	createdAt?: string;
}

function parseSnapshotTitle(snapshot?: string): string {
	if (!snapshot) return "Product";
	try {
		const parsed = JSON.parse(snapshot) as { title?: string; name?: string };
		return parsed.title || parsed.name || "Product";
	} catch {
		return "Product";
	}
}

/** Amounts in the order payload are cents, matching OrderItem.priceAtPurchase. */
function toMajor(cents: number): number {
	return cents / 100;
}

function formatAddress(addr?: InvoiceOrder["shippingAddress"]): string {
	if (!addr) return "";
	const parts = [addr.street, addr.city, addr.state, addr.postalCode, addr.country].filter(Boolean);
	return parts.join(", ");
}

/**
 * Totally non-visual, print-only invoice. It never renders on screen
 * (`.print-only` hides it); the print stylesheet swaps it in and hides the
 * app chrome, so the printed page contains only this document.
 */
export default function InvoicePrintView({ order }: { order: InvoiceOrder }) {
	const t = useTranslations("orderDetails");
	const tOrders = useTranslations("orders");
	const locale = useLocale();

	const currency = (order.currency || "USD").toUpperCase();
	const items = order.items ?? [];
	const subtotal = items.reduce(
		(acc, item) => acc + (item.priceAtPurchase || 0) * (item.quantity || 1),
		0,
	);
	const subtotalVal = subtotal > 0
		? toMajor(subtotal)
		: toMajor((order.totalAmount || 0) + (order.discountAmount || 0));
	const discount = toMajor(order.discountAmount || 0);
	const shipping = toMajor(order.shippingAmount || 0);
	const tax = toMajor(order.taxAmount || 0);
	const cod = toMajor(order.codFee || 0);
	const total = toMajor(order.totalAmount || 0);
	const placedDate = order.createdAt
		? new Date(order.createdAt).toLocaleDateString(locale, {
				year: "numeric",
				month: "long",
				day: "numeric",
		  })
		: "—";

	return (
		<div className="print-only" data-invoice-print="true">
			<h1>{t("invoice.title")}</h1>
			<p>{t("orderTitle", { id: order.id.slice(0, 8).toUpperCase() })}</p>

			<table>
				<tbody>
					<tr>
						<th>{t("customer")}</th>
						<td>{order.customerEmail}</td>
					</tr>
					<tr>
						<th>{t("deliveryAddress")}</th>
						<td>{formatAddress(order.shippingAddress) || "—"}</td>
					</tr>
					<tr>
						<th>{t("invoice.date")}</th>
						<td>{placedDate}</td>
					</tr>
					<tr>
						<th>{t("status")}</th>
						<td>{tOrders(`status.${order.status}`)}</td>
					</tr>
					{(order.carrier || order.trackingNumber) && (
						<tr>
							<th>{t("carrier")}</th>
							<td>
								{[order.carrier, order.trackingNumber].filter(Boolean).join(" - ")}
							</td>
						</tr>
					)}
				</tbody>
			</table>

			<table>
				<thead>
					<tr>
						<th>{t("invoice.item")}</th>
						<th>{t("invoice.quantity")}</th>
						<th>{t("invoice.unitPrice")}</th>
						<th>{t("invoice.amount")}</th>
					</tr>
				</thead>
				<tbody>
					{items.map((item) => {
						const lineTotal = toMajor((item.priceAtPurchase || 0) * (item.quantity || 1));
						return (
							<tr key={item.id}>
								<td>{parseSnapshotTitle(item.productSnapshot)}</td>
								<td>{item.quantity}</td>
								<td>{toMajor(item.priceAtPurchase || 0).toFixed(2)} {currency}</td>
								<td>{lineTotal.toFixed(2)} {currency}</td>
							</tr>
						);
					})}
				</tbody>
			</table>

			<table>
				<tbody>
					<tr>
						<th>{t("subtotal")}</th>
						<td>{subtotalVal.toFixed(2)} {currency}</td>
					</tr>
					{discount > 0 && (
						<tr>
							<th>
								{t("discount")}
								{order.couponCode ? ` (${order.couponCode})` : ""}
							</th>
							<td>-{discount.toFixed(2)} {currency}</td>
						</tr>
					)}
					<tr>
						<th>{t("shipping")}</th>
						<td>{shipping > 0 ? `${shipping.toFixed(2)} ${currency}` : t("freeShipping")}</td>
					</tr>
					{tax > 0 && (
						<tr>
							<th>{t("invoice.tax")}</th>
							<td>{tax.toFixed(2)} {currency}</td>
						</tr>
					)}
					{cod > 0 && (
						<tr>
							<th>{t("invoice.codFee")}</th>
							<td>{cod.toFixed(2)} {currency}</td>
						</tr>
					)}
					<tr>
						<th>{t("total")}</th>
						<td>{total.toFixed(2)} {currency}</td>
					</tr>
				</tbody>
			</table>
		</div>
	);
}
