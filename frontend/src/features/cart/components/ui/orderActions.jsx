import QuantitySelector from "@/components/ui/quantitySelector";

function OrderActions() {
	return (
		<div className="flex flex-col gap-2.5">
			<QuantitySelector />
			<p className="ml-auto w-fit text-foreground font-medium text-txt-sm md:text-txt-md lg:text-txt-lg">
				$45
			</p>
			<button className="block ml-auto text-orange-400 text-txt-sm md:text-txt-md lg:text-txt-lg">
				Remove
			</button>
		</div>
	);
}

export default OrderActions;
