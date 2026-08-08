export function CustomInput({
	type,
	value,
	isDisabled,
	placeholder,
	onChange,
}) {
	return (
		<input
			type={type}
			placeholder={placeholder}
			onChange={onChange}
			value={value}
			disabled={isDisabled}
			className="p-2.5 truncate w-full text-txt-sm md:text-txt-md lg:text-txt-lg placeholder:text-muted-foreground disabled:text-muted-foreground"
		/>
	);
}
