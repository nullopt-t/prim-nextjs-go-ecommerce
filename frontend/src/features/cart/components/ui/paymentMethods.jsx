function PaymentMethods() {
	return (
		<p className="flex justify-center gap-2.5 text-muted-foreground text-txt-sm">
			<span className="bg-[#0000001a] p-1 pl-2 pr-2 rounded-md">
				Visa
			</span>
			<span className="bg-[#0000001a] p-1 pl-2 pr-2 rounded-md">MC</span>
			<span className="bg-[#0000001a] p-1 pl-2 pr-2 rounded-md">
				PayPal
			</span>
			<span className="bg-[#0000001a] p-1 pl-2 pr-2 rounded-md">
				Apple Pay
			</span>
		</p>
	);
}

export default PaymentMethods;
