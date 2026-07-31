import OrderDetails from "@/features/cart/components/ui/orderDetails";
import OrderActions from "@/features/cart/components/ui/orderActions";

function OrderBox({ orderDetails }) {
	return (
		<div className="flex justify-between border border-border rounded-md p-2.5 md:p-5">
			<OrderDetails details={orderDetails} />
			<OrderActions />
		</div>
	);
}

export default OrderBox;
