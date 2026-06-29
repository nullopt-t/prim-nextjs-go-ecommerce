import UserSideBar from "@/features/user/components/ui/sideBar";

function UserLayout({ content }) {
	return (
		<div className="flex gap-2.5 md:gap-5">
			<UserSideBar />
			<div className="flex-1 pl-2.5 pr-2.5">{content}</div>
		</div>
	);
}

export default UserLayout;
