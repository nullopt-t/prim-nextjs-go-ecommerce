export function Title({ title, subtitle, textColor }) {
	return (
		<p className="flex flex-col mb-5">
			<span
				className={`${textColor || "text-foreground"} capitalize font-medium text-title-lg md:text-title-md lg:text-title-lg`}
			>
				{title}
			</span>
			<span className="text-txt-lg md:text-txt-md lg:text-txt-lg text-muted-foreground">
				{subtitle}
			</span>
		</p>
	);
}
