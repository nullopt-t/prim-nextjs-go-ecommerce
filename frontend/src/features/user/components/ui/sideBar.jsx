import UserSideBarProfile from "@/features/user/components/ui/sideBarProfile";
import UserSideBarLinks from "@/features/user/components/ui/sideBarLinks";

function UserSideBar() {
	return (
		<div className="border-2 border-sidebar-border rounded-md bg-sidebar text-sidebar-foreground p-2.5">
			<UserSideBarProfile />
			<UserSideBarLinks />
		</div>
	);
}

export default UserSideBar;
