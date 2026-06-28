import MainLayout from "@/components/layouts/mainLayout";
import UserContent from "@/features/user/components/ui/userContent";
import { Outlet } from "react-router-dom";

export function User() {
	return (
		<MainLayout>
			<UserContent content={<Outlet />} />
		</MainLayout>
	);
}
