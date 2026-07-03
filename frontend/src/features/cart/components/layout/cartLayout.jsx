import CartTitle from "@/features/cart/components/ui/cartTitle";
import PaymentBox from "@/features/cart/components/ui/paymentBox";
import OrderBox from "@/features/cart/components/ui/orderBox";
import Copoun from "@/features/cart/components/ui/copoun";

function CartLayout() {
	return (
		<div className="">
			<CartTitle />
			<div className="flex flex-col md:flex-row mt-5 gap-5 md:gap-7">
				<div className="flex-4 flex flex-col justify-between">
					<OrderBox />
					<Copoun />
				</div>
				<div className="flex-1">
					<PaymentBox />
				</div>
			</div>
		</div>
	);
}

export default CartLayout;
