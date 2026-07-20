import SectionCard from "@/features/home/components/ui/sectionCard";

export default function SectionGrid() {
	const categories = [
		{
			id: "cat_1",
			slug: "audio",
			name: {
				en: "Audio",
				ar: "الأجهزة الصوتية",
			},
		},
		{
			id: "cat_2",
			slug: "wearables",
			name: {
				en: "Wearables",
				ar: "الأجهزة القابلة للارتداء",
			},
		},
		{
			id: "cat_3",
			slug: "desk-setup",
			name: {
				en: "Desk Setup",
				ar: "مستلزمات المكتب",
			},
		},
		{
			id: "cat_4",
			slug: "accessories",
			name: {
				en: "Accessories",
				ar: "الإكسسوارات",
			},
		},
		{
			id: "cat_5",
			slug: "lighting",
			name: {
				en: "Lighting",
				ar: "وحدات الإضاءة",
			},
		},
	];

	return (
		<div className="grid grid-cols-[repeat(auto-fill,minmax(150px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-5">
			{categories.map((value) => (
				<SectionCard key={value.id} category={value.name.en} />
			))}
		</div>
	);
}
