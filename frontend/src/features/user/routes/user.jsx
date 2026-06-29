import MainLayout from "@/components/layouts/mainLayout";
import UserLayout from "@/features/user/components/layout/userLayout";
import { Outlet } from "react-router-dom";

export function User() {
	return (
		<MainLayout>
			<UserLayout content={<Outlet />} />
		</MainLayout>
	);
}
