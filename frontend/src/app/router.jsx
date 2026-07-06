import { Routes, Route, Navigate } from "react-router-dom";
import { Cart } from "@/features/cart";
import { Home } from "@/features/home";
import { Auth, Login, Verify } from "@/features/auth";
import {
	User,
	Payment,
	Overview,
	Orders,
	Reviews,
	Settings,
	Wishlist,
	Address,
} from "@/features/user";

function Router() {
	return (
		<Routes>
			<Route path="/" element={<Navigate to="/home" />} />

			<Route path="/home" element={<Home />} />

			<Route path="/auth" element={<Auth />}>
				<Route path="login" element={<Login />} />
				<Route path="verify" element={<Verify />} />
			</Route>

			<Route path="user" element={<User />}>
				<Route path="overview" element={<Overview />} />
				<Route path="orders" element={<Orders />} />
				<Route path="payment" element={<Payment />} />
				<Route path="reviews" element={<Reviews />} />
				<Route path="settings" element={<Settings />} />
				<Route path="wishlist" element={<Wishlist />} />
				<Route path="address" element={<Address />} />
			</Route>

			<Route path="/cart" element={<Cart />} />
		</Routes>
	);
}

export default Router;
