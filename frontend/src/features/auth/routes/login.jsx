import Form from "@/features/auth/components/ui/form";
import PostData from "@/api/post";
import useFetch from "@/hooks/useFetch";
import AuthEndpoints from "@/features/auth/api/endpoints/authEndpoints";
import { useEffect } from "react";

export function Login() {
	return (
		<div className="overflow-y-auto max-h-full">
			<Form formType="login" />
		</div>
	);
}
