import { Header, Footer } from "@/components/ui";

function MainLayout({ children }) {
	return (
		<div className="flex flex-col min-h-screen">
			<Header />
			<div className="p-2.5 pb-0 md:p-5 md:pb-0 flex-1">{children}</div>
			<Footer />
		</div>
	);
}

export default MainLayout;
