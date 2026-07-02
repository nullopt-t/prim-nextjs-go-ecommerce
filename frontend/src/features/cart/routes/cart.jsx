import MainLayout from "@/components/layouts/mainLayout";
import CartTitle from "@/features/cart/components/ui/cartTitle";
import PaymentBox from "@/features/cart/components/ui/paymentBox";

export function Cart() {
	return (
		<MainLayout>
			<CartTitle />
			<PaymentBox />
		</MainLayout>
	);
}
