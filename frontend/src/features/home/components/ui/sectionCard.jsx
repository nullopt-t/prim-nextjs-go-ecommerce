export default function SectionCard({ category }) {
	return (
		<div className="cursor-pointer p-2.5 pt-5 pb-5 capitalize text-black hover:text-accent-foreground hover:border-accent-brand bg-[#f9f9f9] font-medium flex items-center justify-center rounded-sm border-2 border-border hover:bg-accent">
			{category}
		</div>
	);
}
