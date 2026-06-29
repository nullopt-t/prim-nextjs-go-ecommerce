import { NavLink } from "react-router-dom";
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
			path: "overview",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-2",
			icon: Box,
			link: "My orders",
			path: "orders",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-3",
			icon: Heart,
			link: "Wishlist",
			path: "wishlist",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-4",
			icon: MapPin,
			link: "Addresses",
			path: "address",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-5",
			icon: CreditCard,
			link: "Payment methods",
			path: "payment",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-6",
			icon: Star,
			link: "Reviews",
			path: "reviews",
			isLogout: false,
		},
		{
			id: "user-sidbar-links-7",
			icon: Settings,
			link: "Settings",
			path: "settings",
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
					className={`hover:bg-sidebar-accent overflow-hidden rounded-md ${value.isLogout ? "text-red-500" : ""}`}
				>
					<NavLink
						to={value.isLogout ? "" : value.path}
						className={({ isActive }) =>
							`flex gap-2.5 items-center p-2.5 ${isActive ? "text-sidebar-accent-foreground bg-sidebar-accent" : ""}`
						}
					>
						<span>
							<value.icon className="" />
						</span>
						<span className="hidden md:block">{value.link}</span>
					</NavLink>
				</li>
			))}
		</ul>
	);
}

export default UserSideBarLinks;
