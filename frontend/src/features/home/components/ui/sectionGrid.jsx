import SectionCard from "@/features/home/components/ui/sectionCard";
import { useTranslation } from "react-i18next";

export default function SectionGrid() {
	const { i18n } = useTranslation("home");

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
		{
			id: "cat_6",
			slug: "gaming",
			name: {
				en: "Gaming",
				ar: "الألعاب",
			},
		},
		{
			id: "cat_7",
			slug: "smart-home",
			name: {
				en: "Smart Home",
				ar: "المنزل الذكي",
			},
		},
		{
			id: "cat_",
			slug: "Others",
			name: {
				en: "Others",
				ar: "اخر",
			},
		},
	];

	return (
		<div className="grid grid-cols-[repeat(auto-fill,minmax(100px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(150px,1fr))] gap-5">
			{categories.map((value) => (
				<SectionCard
					key={value.id}
					category={
						i18n.resolvedLanguage === "en"
							? value.name.en
							: value.name.ar
					}
				/>
			))}
		</div>
	);
}
