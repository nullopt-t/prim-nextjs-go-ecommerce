function Button({ type, payload, handle }) {
	const buttonTxt =
		type === "login" ? "enter" : type === "verify" ? "Verify OTP Code" : "";

	return (
		<input
			className="bg-primary text-primary-foreground rounded-sm pt-2.5 pb-2.5 font-medium capitalize hover:bg-accent hover:text-accent-foreground"
			onClick={(e) => {
				e.preventDefault();
				handle(type, payload);
			}}
			type="submit"
			value={buttonTxt}
		/>
	);
}

export default Button;
