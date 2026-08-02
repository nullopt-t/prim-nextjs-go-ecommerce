import SideBar from "@/features/products/components/ui/sideBar";
import Content from "@/features/products/components/ui/content";
import Title from "@/components/ui/title";
import Sort from "@/components/ui/sort";

function ProductsLayout() {
	return (
		<div className="">
			<div className="flex">
				<Title title="Headphones" subtitle="1,240 results" />
				<Sort />
			</div>
			<div className="relative flex gap-2.5 md:gap-5 lg:gap-10">
				<div className="">
					<SideBar />
				</div>
				<div className="flex-1">
					<Content />
				</div>
			</div>
		</div>
	);
}

export default ProductsLayout;
