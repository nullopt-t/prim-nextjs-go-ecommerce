import Team from "@/features/about/ui/team";
import Hero from "@/features/about/ui/hero";
import Beliefs from "@/features/about/ui/beliefs";

export default function AboutLayout() {
	return (
		<div className="mb-5 flex flex-col gap-10">
			<Hero />
			<Beliefs />
			<Team />
		</div>
	);
}
