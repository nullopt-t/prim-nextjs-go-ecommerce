import Form from "@/features/auth/components/ui/form";
import PostData from "@/api/post";
import AuthEndpoints from "@/features/auth/api/endpoints/authEndpoints";

export function Login() {
	const handleSubmit = async (type, payload) => {
		try {
			let result;
			if (type === "login" || type === "register") {
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
			<Form formType="login" handleSubmit={handleSubmit} />
		</div>
	);
}
