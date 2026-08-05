export function Title({ title, subtitle }) {
	return (
		<p className="flex flex-col mb-5">
			<span className="capitalize font-medium text-title-lg md:text-title-md lg:text-title-lg text-foreground">
				{title}
			</span>
			<span className="text-txt-lg md:text-txt-md lg:text-txt-lg text-muted-foreground">
				{subtitle}
			</span>
		</p>
	);
}
