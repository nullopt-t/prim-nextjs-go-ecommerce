import { Routes, Route, Navigate } from "react-router-dom";
import { Cart } from "@/features/cart";
import { Products } from "@/features/products";
import { Home } from "@/features/home";
import { About } from "@/features/about";
import { Wishlist } from "@/features/wishlist";
import { Settings } from "@/features/settings";
import { Reviews } from "@/features/reviews";
import { Auth, Login, Verify } from "@/features/auth";
import { NotFound } from "@/app/pages/notFound";
import { User, Payment, Overview, Orders, Address } from "@/features/user";

function Router() {
	return (
		<Routes>
			<Route path="/" element={<Navigate to="/home" />} replace />

			<Route path="/home" element={<Home />} />

			<Route path="/auth" element={<Auth />}>
				<Route index element={<Login />} />
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
			<Route path="/about" element={<About />} />
			<Route path="/products" element={<Products />} />

			<Route path="*" element={<NotFound />} />
		</Routes>
	);
}

export default Router;
