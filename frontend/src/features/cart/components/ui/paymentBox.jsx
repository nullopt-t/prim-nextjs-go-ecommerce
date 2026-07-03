import CustomButton from "@/components/ui/button";
import PaymentMethods from "@/features/cart/components/ui/paymentMethods";

function PaymentBox() {
	return (
		<div className="border border-border p-2.5 md:p-5 rounded-md">
			<p className="mb-5 font-medium">Order summary</p>
			<div className="flex flex-col gap-2.5 mb-5">
				<p className="flex justify-between text-txt-sm md:text-txt-sm lg:text-txt-sm">
					<span className="text-muted-foreground">Subtotal</span>
					<span className="text-foreground font-medium">$578.98</span>
				</p>
				<p className="flex justify-between text-txt-sm md:text-txt-sm lg:text-txt-sm">
					<span className="text-muted-foreground">Shipping</span>
					<span className="text-foreground font-medium">Free</span>
				</p>
				<p className="flex justify-between text-txt-sm md:text-txt-sm lg:text-txt-sm">
					<span className="text-muted-foreground">
						Tax &#40;8%&#41;
					</span>
					<span className="text-foreground font-medium">$46.32</span>
				</p>
			</div>
			<hr className="text-border" />
			<div className="mt-5">
				<p className="flex justify-between text-title-sm md:text-title-sm">
					<span className="text-foreground font-medium">Total</span>
					<span className="text-foreground font-medium">$625.50</span>
				</p>
			</div>
			<div className="mt-5 h-12">
				<CustomButton text="Proceed to checkout" />
			</div>
			<div className="mt-5">
				<PaymentMethods />
			</div>
		</div>
	);
}

export default PaymentBox;
