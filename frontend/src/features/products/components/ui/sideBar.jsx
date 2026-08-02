import SideBarCategories from "@/features/products/components/ui/sideBarCategories";
import SideBarDiscount from "@/features/products/components/ui/sideBarDiscount";
import SideBarRatings from "@/features/products/components/ui/sideBarRatings";
import SideBarFilters from "@/features/products/components/ui/sideBarFilters";
import SideBarAvailability from "@/features/products/components/ui/sideBarAvailability";
import SideBarBrands from "@/features/products/components/ui/sideBarBrands";
import { PanelLeftOpen, PanelLeftClose } from "lucide-react";
import { useState } from "react";

function SideBar() {
	const [isOpen, setIsOpen] = useState(true);

	return (
		<div className="p-2.5 lg:p-5 border-r border-r-border">
			<button
				className="block ml-auto text-muted-foreground mb-5 hover:text-accent-brand"
				onClick={() => setIsOpen((prev) => !prev)}
			>
				{isOpen ? <PanelLeftClose /> : <PanelLeftOpen />}
			</button>

			<div
				className={`flex flex-col gap-5 overflow-hidden ${isOpen ? "w-52" : "w-0"}`}
			>
				<SideBarFilters />
				<SideBarCategories />
				<SideBarBrands />
				<SideBarRatings />
				<SideBarAvailability />
				<SideBarDiscount />
			</div>
		</div>
	);
}

export default SideBar;
