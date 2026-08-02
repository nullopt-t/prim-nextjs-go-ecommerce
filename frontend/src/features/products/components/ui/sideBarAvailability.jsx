import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import { useState } from "react";

export default function SideBarAvailability() {
	const [inStock, setInStock] = useState(false);

	return (
		<div className="border-b border-border pb-5">
			<SideBarTitle title="availability" />
			<div className="flex justify-between items-center">
				<p className="text-muted-foreground">In stock only</p>
				<label
					className={`
		relative
		h-6
		w-12
		cursor-pointer
		overflow-hidden
		rounded-full
		transition-colors
		duration-300
		${inStock ? "bg-accent-brand" : "bg-switch-background"}
		
		before:absolute
		before:left-0.5
		before:top-0.5
		before:h-5
		before:w-5
		before:rounded-full
		before:bg-white
		before:shadow
		before:content-['']
		before:transition-transform
		before:duration-300
		${inStock ? "before:translate-x-6" : ""}
	`}
				>
					<input
						type="checkbox"
						checked={inStock}
						onChange={(e) => setInStock(e.target.checked)}
						className="sr-only"
					/>
				</label>
			</div>
		</div>
	);
}
