import { Header, Footer } from "@/components/ui";

function MainLayout({ children }) {
	return (
		<>
			<Header />
			<div className="p-2.5 md:p-5">{children}</div>
			<Footer />
		</>
	);
}

export default MainLayout;
