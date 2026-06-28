import {
	Blocks,
	Box,
	Heart,
	MapPin,
	CreditCard,
	Star,
	Settings,
	LogOut,
} from "lucide-react";

function UserSideBarLinks() {
	const links = [
		{
			id: "user-sidbar-links-1",
			icon: Blocks,
			link: "Overview",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-2",
			icon: Box,
			link: "My orders",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-3",
			icon: Heart,
			link: "Wishlist",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-4",
			icon: MapPin,
			link: "Addresses",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-5",
			icon: CreditCard,
			link: "Payment methods",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-6",
			icon: Star,
			link: "Reviews",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-7",
			icon: Settings,
			link: "Settings",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-8",
			icon: LogOut,
			link: "Logout",
			isLogout: true,
		},
	];

	return (
		<ul className="">
			{links.map((value) => (
				<li
					key={value.id}
					className={`hover:bg-sidebar-accent p-2.5 rounded-md cursor-pointer flex gap-2.5 items-center${value.isLogout ? "text-red-500" : ""}`}
				>
					<span className="">
						<value.icon className="" />
					</span>
					<span className="hidden md:block">{value.link}</span>
				</li>
			))}
		</ul>
	);
}

export default UserSideBarLinks;
