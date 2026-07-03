function CustomInput({ type, placeholder, handle }) {
	return (
		<input
			type={type}
			placeholder={placeholder}
			onClick={handle}
			className="truncate w-full text-txt-sm md:text-txt-md lg:text-txt-lg text-foreground placeholder:text-muted-foreground"
		/>
	);
}

export default CustomInput;
