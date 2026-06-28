import UserSideBar from "@/features/user/components/ui/sideBar";

function UserContent({ content }) {
	return (
		<div className="flex">
			<UserSideBar />
			{content}
		</div>
	);
}

export default UserContent;
