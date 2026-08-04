export function CustomButton({ text, onClick }) {
	return (
		<button
			className="border border-border w-full h-full rounded-md font-medium text-txt-sm md:text-txt-md lg:text-txt-lg"
			onClick={onClick}
		>
			{text}
		</button>
	);
}
