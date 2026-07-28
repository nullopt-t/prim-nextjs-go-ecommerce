import { Header, Footer } from "@/components/ui";

function MainLayout({ children, recently }) {
	return (
		<div className="flex flex-col min-h-screen">
			<Header />
			<div className="p-2.5 pt-0 md:p-5 md:pt-0 lg:pt-0 lg:p-10 flex-1">
				{children}
			</div>
			{recently}
			<Footer />
		</div>
	);
}

export default MainLayout;
