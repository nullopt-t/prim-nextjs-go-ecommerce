import OrderDetails from "@/features/cart/components/ui/orderDetails";
import OrderActions from "@/features/cart/components/ui/orderActions";

function OrderBox() {
	return (
		<div className="flex justify-between border border-border rounded-md p-2.5 md:p-5">
			<OrderDetails />
			<OrderActions />
		</div>
	);
}

export default OrderBox;
