import { ChevronUp } from "lucide-react";
import { useState, useRef, useEffect } from "react";
import SideBarTitle from "@/features/products/components/ui/sideBarTitle";

function SideBarCategories() {
	const catRef = useRef(null);

	const [height, setHeight] = useState(0);
	const [isOpen, setIsOpen] = useState(true);

	const categories = [
		{
			id: "ctg-1",
			name: "All categories",
		},
		{
			id: "ctg-2",
			name: "Headphones",
		},
		{
			id: "ctg-3",
			name: "Speakers",
		},
	];

	const handleArrowClick = () => {
		setIsOpen((prev) => !prev);
	};

	useEffect(() => {
		if (catRef.current) {
			setHeight(isOpen ? catRef.current.scrollHeight : 0);
		}
	}, [isOpen]);

	return (
		<div className="border-b border-border pb-5">
			<p className="flex justify-between items-center">
				<SideBarTitle title="category" />

				<button
					type="button"
					onClick={handleArrowClick}
					className="cursor-pointer"
				>
					<ChevronUp
						className={`size-4 transition-transform duration-300 ${
							isOpen ? "rotate-0" : "rotate-180"
						}`}
					/>
				</button>
			</p>

			<div
				style={{ height: `${height}px` }}
				className="transition-[height] duration-300 ease-in-out"
			>
				<ul ref={catRef} className="mt-3 space-y-2">
					{categories.map((category) => (
						<li
							key={category.id}
							className="text-txt-sm md:text-txt-md lg:text-txt-lg text-muted-foreground"
						>
							{category.name}
						</li>
					))}
				</ul>
			</div>
		</div>
	);
}

export default SideBarCategories;
