import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import { useState } from "react";
import { Toggle } from "@/components/ui/Toggle";

export default function SideBarAvailability() {
	const [inStock, setInStock] = useState(false);

	return (
		<div className="border-b border-border pb-5">
			<SideBarTitle title="availability" />
			<div className="flex justify-between items-center">
				<p className="text-muted-foreground">In stock only</p>
				<Toggle
					isEnabled={inStock}
					onChange={(e) => setInStock(e.target.checked)}
				/>
			</div>
		</div>
	);
}
