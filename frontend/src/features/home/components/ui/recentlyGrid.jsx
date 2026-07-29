import RecentlyCard from "@/features/home/components/ui/recentlyCard";
import { useTranslation } from "react-i18next";

export default function RecentlyGrid() {
	const { i18n } = useTranslation("home");

	const recentArr = [
		{
			id: "rec-001",
			title: {
				en: "Smart Watches",
				ar: "الساعات الذكية",
			},
		},
		{
			id: "rec-002",
			title: {
				en: "Wireless Earbuds",
				ar: "سماعات لاسلكية",
			},
		},
		{
			id: "rec-003",
			title: {
				en: "Gaming Keyboards",
				ar: "لوحات مفاتيح الألعاب",
			},
		},
		{
			id: "rec-004",
			title: {
				en: "Bluetooth Speakers",
				ar: "مكبرات صوت بلوتوث",
			},
		},
		{
			id: "rec-005",
			title: {
				en: "Phone Cases",
				ar: "أغطية الهواتف",
			},
		},
	];

	return (
		<div className="flex flex-wrap gap-5 mt-5">
			{recentArr.map((value) => (
				<RecentlyCard
					key={value.id}
					title={
						i18n.resolvedLanguage === "en"
							? value.title.en
							: value.title.ar
					}
				/>
			))}
		</div>
	);
}
