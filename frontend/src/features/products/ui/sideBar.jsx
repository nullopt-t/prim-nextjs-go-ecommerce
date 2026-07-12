import SideBarCategories from "@/features/products/ui/sideBarCategories";
import SideBarDiscount from "@/features/products/ui/sideBarDiscount";
import SideBarRatings from "@/features/products/ui/sideBarRatings";
import SideBarFilters from "@/features/products/ui/sideBarFilters";

function SideBar() {
	return (
		<div className="">
			<div className="text-foreground">
				<p className="font-bold text-title-sm md:text-title-md lg:text-title-lg">
					Headphones
				</p>
				<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
					1200 products
				</p>
			</div>
			<SideBarFilters />
			<SideBarCategories />
			<SideBarDiscount />
			<SideBarRatings />
		</div>
	);
}

export default SideBar;
