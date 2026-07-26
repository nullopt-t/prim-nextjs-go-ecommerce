import { HiMiniArrowLongRight } from "react-icons/hi2";

export default function SectionTitle({ title }) {
	return (
		<div className="flex justify-between">
			<p className="capitalize font-medium text-title-sm md:text-title-md text-foreground">
				{title}
			</p>
			<p className="flex gap-5 items-center text-txt-sm md:text-txt-md lg:text-txt-lg cursor-pointer text-accent-brand hover:underline hover:underline-offset-4">
				<span className="">See all</span>
				<span className="">
					<HiMiniArrowLongRight />
				</span>
			</p>
		</div>
	);
}
