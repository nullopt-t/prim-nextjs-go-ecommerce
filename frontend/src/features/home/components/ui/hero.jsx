import Button from "@/components/ui/button";

export default function Hero() {
	return (
		<div className="p-15">
			<p className="text-accent-brand mb-5">New Season — 2026</p>
			<p className="mb-2.5 text-white font-black text-title-sm md:text-title-md lg:text-title-lg">
				<span className="">Everything you need, </span>
				<span className="text-accent-brand">delivered.</span>
			</p>
			<p className="text-muted-foreground mb-10">
				Discover thousands of products from trusted brands, all in one
				place. Fast shipping, easy returns, and unbeatable prices.
			</p>
			<div className="flex gap-5">
				<div className="w-32 h-12 bg-accent-brand hover:scale-90 text-white rounded-md">
					<Button text="Shop now" />
				</div>
				<div className="bg-secondary text-secondary-foreground hover:scale-90 rounded-md w-32 h-12">
					<Button text="View deals" />
				</div>
			</div>
		</div>
	);
}
