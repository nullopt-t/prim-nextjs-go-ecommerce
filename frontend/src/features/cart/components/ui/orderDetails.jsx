import ProductImage from "@/assets/product.jpeg";

function OrderDetails() {
	return (
		<div className="flex gap-2.5 md:gap-5">
			<img
				src={ProductImage}
				alt="Order Image"
				className="rounded-md object-cover object-center aspect-square w-20 md:w-24"
			/>
			<div className="">
				<p className="font-medium text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg mb-0.5">
					Smart watch
				</p>
				<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg mb-2.5">
					Huawi
				</p>
				<p className="font-medium text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
					$200
				</p>
			</div>
		</div>
	);
}

export default OrderDetails;
