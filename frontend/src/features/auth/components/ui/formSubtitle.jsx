function FormSubtitle({ type }) {
	const subTitle =
		type === "login"
			? "Enter your email to sign in or create a new account."
			: type === "verify"
				? "We sent a verification code to your email. Enter it below to proceed."
				: "";

	return (
		<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
			{subTitle}
		</p>
	);
}

export default FormSubtitle;
