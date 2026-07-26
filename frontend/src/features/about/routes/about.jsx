import MainLayout from "@/components/layouts/mainLayout";
import Team from "@/features/about/ui/team";
import Beliefs from "@/features/about/ui/beliefs";

export function About() {
	return (
		<MainLayout>
			<div className="mb-5">
				<div className="mb-10">
					<Beliefs />
				</div>
				<Team />
			</div>
		</MainLayout>
	);
}
