import SideBar from "@/features/products/ui/sideBar";
import Content from "@/features/products/ui/content";

function ProductsLayout() {
	return (
		<div className="">
			<div className="">
				<SideBar />
			</div>
			<div className="">
				<Content />
			</div>
		</div>
	);
}

export default ProductsLayout;
