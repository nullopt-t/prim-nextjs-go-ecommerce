import Form from "@/features/auth/components/ui/form";
import PostData from "@/api/post";
import AuthEndpoints from "@/features/auth/api/endpoints/authEndpoints";

export function Verify() {
	const validateEmail = (email) => {
		if (!email.trim()) {
			return "Email is required.";
		}
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		if (!emailRegex.test(email)) {
			return "Please enter a valid email address.";
		}
		return null;
	};

	const validateCode = (code) => {
		if (!code.trim()) {
			return "Verification code is required.";
		}
		const codeRegex = /^\d{6}$/;
		if (!codeRegex.test(code)) {
			return "Verification code must contain 6 digits.";
		}
		return null;
	};

	const handleSubmit = async (type, payload) => {
		if (type === "login") {
			const emailError = validateEmail(payload.email);

			if (emailError) {
				console.log(emailError);
				return;
			}
		}

		if (type === "verify") {
			const codeError = validateCode(payload.code);

			if (codeError) {
				console.log(codeError);
				return;
			}
		}

		try {
			let result;

			if (type === "login") {
				result = await PostData(AuthEndpoints.start, {
					identifier: payload.email,
				});

				sessionStorage.setItem("identifier", payload.email);
			} else if (type === "verify") {
				result = await PostData(AuthEndpoints.verifyCode, {
					code: payload.code,
					identifier: sessionStorage.getItem("identifier"),
				});
			}

			console.log(result);
		} catch (err) {
			console.log(err);
		}
	};

	return (
		<div className="overflow-y-auto max-h-full">
			<Form formType="verify" handleSubmit={handleSubmit} />
		</div>
	);
}
