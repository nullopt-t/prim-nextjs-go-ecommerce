import Team from "@/features/about/components/ui/team";
import Hero from "@/features/about/components/ui/hero";
import Beliefs from "@/features/about/components/ui/beliefs";

export default function AboutLayout() {
	return (
		<div className="mb-5 flex flex-col gap-10">
			<Hero />
			<Beliefs />
			<Team />
		</div>
	);
}
