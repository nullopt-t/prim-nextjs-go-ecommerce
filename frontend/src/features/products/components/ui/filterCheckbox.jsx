export default function FilterCheckbox({ labelTxt }) {
	return (
		<label className="flex gap-2.5 items-center group">
			<input type="checkbox" className="size-4" />
			<p
				className="text-muted-foreground group-hover:text-accent-brand"
				checked={true}
			>
				{labelTxt}
			</p>
		</label>
	);
}
