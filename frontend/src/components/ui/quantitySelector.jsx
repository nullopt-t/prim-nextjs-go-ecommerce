function QuantitySelector() {
	return (
		<div className="text-foreground flex border border-border rounded-md w-24 md:w-36 text-txt-sm md:text-txt-md lg:text-txt-lg">
			<button className="pt-1 pb-1 flex-1 text-center border-r border-border hover:bg-accent hover:text-accent-foreground">
				-
			</button>
			<p className="pt-1 pb-1 flex-2 text-center">1</p>
			<button className="pt-1 pb-1 flex-1 text-center border-l border-border hover:bg-accent hover:text-accent-foreground">
				+
			</button>
		</div>
	);
}

export default QuantitySelector;
