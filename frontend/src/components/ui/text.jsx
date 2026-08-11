export function Text({ text, className }) {
	return (
		<p
			className={`${className} text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg`}
		>
			{text}
		</p>
	);
}
