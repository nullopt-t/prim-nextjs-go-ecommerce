import {
	FaInstagram,
	FaSquareFacebook,
	FaSquareXTwitter,
	FaTiktok,
} from "react-icons/fa6";
import { useTranslation } from "react-i18next";
const currentDate = new Date().getFullYear();

export function Footer() {
	const { t } = useTranslation("common");

	const sections = [
		{
			id: "shop-pp",
			title: "footer.shop",
			content: [
				{
					id: "cat_1",
					slug: "audio",
					name: "Audio",
				},
				{
					id: "cat_2",
					slug: "wearables",
					name: "Wearables",
				},
				{
					id: "cat_3",
					slug: "desk-setup",
					name: "Desk Setup",
				},
			],
		},
		{
			id: "support-pp",
			title: "footer.support",
			content: [
				{
					id: "supp-1",
					name: "footer.contactUs",
				},
				{
					id: "supp_2",
					name: "footer.faqs",
				},
				{
					id: "supp_3",
					name: "footer.shipping",
				},
				{
					id: "supp_4",
					name: "footer.returns",
				},
				{
					id: "supp_5",
					name: "footer.trackOrders",
				},
				{
					id: "supp_6",
					name: "footer.sizeGuide",
				},
			],
		},
		{
			id: "company-pp",
			title: "footer.company",
			content: [
				{
					id: "comp-1",
					name: "footer.aboutUs",
				},
				{
					id: "comp-2",
					name: "footer.careers",
				},
				{
					id: "comp-3",
					name: "footer.press",
				},
				{
					id: "comp-4",
					name: "footer.sustainability",
				},
				{
					id: "comp-5",
					name: "footer.blog",
				},
			],
		},
	];

	const social = [
		{
			id: "insta-1",
			icon: FaInstagram,
		},
		{
			id: "x-1",
			icon: FaSquareXTwitter,
		},
		{
			id: "tik-1",
			icon: FaTiktok,
		},
		{
			id: "facebook-1",
			icon: FaSquareFacebook,
		},
	];

	return (
		<div className="bg-footer p-2.5 md:p-5 lg:p-10">
			<div className="mb-5 flex flex-col gap-10 md:flex-row text-white text-txt-sm md:text-txt-md lg:text-txt-lg">
				<div className="flex-1">
					<p className="mb-2.5 font-black text-title-sm md:text-title-md">
						<span className="text-white">PRI</span>
						<span className="text-accent-brand">M</span>
					</p>
					<p className="text-muted-foreground">
						{t("footer.footerDescription")}
					</p>
				</div>
				{sections.map((value) => (
					<div key={value.id} className="flex-1">
						<p className="mb-5 text-txt-lg font-medium capitalize">
							{t(value.title)}
						</p>
						<p className="flex flex-col gap-2.5">
							{value.content.map((content) => (
								<span
									key={content.id}
									className="capitalize text-muted-foreground cursor-pointer hover:text-accent-brand"
								>
									{t(content.name)}
								</span>
							))}
						</p>
					</div>
				))}
			</div>
			<hr className="text-muted-foreground" />
			<div className="flex flex-col gap-5 md:flex-row md:justify-between text-muted-foreground pt-5">
				<p>{t("footer.copyright", { year: currentDate })}</p>

				<div className="flex gap-2.5">
					{social.map((value) => (
						<value.icon
							key={value.id}
							className="size-6 hover:text-accent-brand"
						/>
					))}
				</div>
			</div>
		</div>
	);
}
