export function Toggle({ isEnabled, onChange }) {
	return (
		<label
			className={`
		relative
		h-6
		w-12
		cursor-pointer
		overflow-hidden
		rounded-full
		transition-colors
		duration-300
		${isEnabled ? "bg-accent-brand" : "bg-switch-background"}
		
		before:absolute
		before:left-0.5
		before:top-0.5
		before:h-5
		before:w-5
		before:rounded-full
		before:bg-white
		before:shadow
		before:content-['']
		before:transition-transform
		before:duration-300
		${isEnabled ? "before:translate-x-6" : ""}
	`}
		>
			<input
				type="checkbox"
				checked={isEnabled}
				onChange={onChange}
				className="sr-only"
			/>
		</label>
	);
}
