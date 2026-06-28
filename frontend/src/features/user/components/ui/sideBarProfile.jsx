// import Image from "@/assets/auth.png";

function UserSideBarProfile() {
	return (
		<div className="p-2.5 hidden md:block">
			{/* <img
				className="rounded-full w-10 h-10"
				src={Image}
				alt="Profile image"
			/> */}
			<p className="flex flex-col">
				<span className="font-medium">Mohamed Mahmoud</span>
				<span className="text-muted-foreground">prim@gmail.com</span>
			</p>
		</div>
	);
}

export default UserSideBarProfile;
