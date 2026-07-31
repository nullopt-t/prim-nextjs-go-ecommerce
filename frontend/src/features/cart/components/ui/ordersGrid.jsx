import OrderBox from "@/features/cart/components/ui/orderBox";

export default function OrdersGrid() {
	const products = [
		{
			id: "ordsl-2lkfl",
			productName: {
				ar: "ساعة ذكية",
				en: "Smart watch",
			},
			productBrand: "Huawi",
			productPrice: "$499",
		},
		{
			id: "ordsl-sjkh4dl",
			productName: {
				ar: "ساعة ذكية",
				en: "Smart watch",
			},
			productBrand: "Huawi",
			productPrice: "$499",
		},
	];

	return (
		<div className="grid grid-cols-1 gap-5">
			{products.map((prod) => (
				<OrderBox key={prod.id} orderDetails={prod} />
			))}
		</div>
	);
}
