import MainLayout from "@/components/layouts/mainLayout";
import HomeLayout from "@/features/home/components/layout/homeLayout";
import Recently from "@/features/home/components/ui/recently";

export function Home() {
	return (
		<MainLayout recently={<Recently />}>
			<HomeLayout />
		</MainLayout>
	);
}
