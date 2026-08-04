import {
	FaInstagram,
	FaSquareFacebook,
	FaSquareXTwitter,
	FaTiktok,
} from "react-icons/fa6";
import { Brand } from "@/components/ui";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";

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
					path: "/home/products/audio",
				},
				{
					id: "cat_2",
					slug: "wearables",
					name: "Wearables",
					path: "/home/products/wearables",
				},
				{
					id: "cat_3",
					slug: "desk-setup",
					name: "Desk Setup",
					path: "/home/products/desk-setup",
				},
			],
		},
		{
			id: "support-pp",
			title: "footer.support",
			content: [
				{
					id: "supp_1",
					name: "footer.contactUs",
					path: "/home/contact",
				},
				{
					id: "supp_2",
					name: "footer.faqs",
					path: "/home/faqs",
				},
				{
					id: "supp_3",
					name: "footer.shipping",
					path: "/home/shipping",
				},
				{
					id: "supp_4",
					name: "footer.returns",
					path: "/home/returns",
				},
				{
					id: "supp_5",
					name: "footer.trackOrders",
					path: "/home/user/orders",
				},
				{
					id: "supp_6",
					name: "footer.sizeGuide",
					path: "/home/size-guide",
				},
			],
		},
		{
			id: "company-pp",
			title: "footer.company",
			content: [
				{
					id: "comp_1",
					name: "footer.aboutUs",
					path: "/about",
				},
				{
					id: "comp_2",
					name: "footer.careers",
					path: "/home/careers",
				},
				{
					id: "comp_3",
					name: "footer.press",
					path: "/home/press",
				},
				{
					id: "comp_4",
					name: "footer.sustainability",
					path: "/home/sustainability",
				},
				{
					id: "comp_5",
					name: "footer.blog",
					path: "/home/blog",
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
					<Brand />
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
								<Link
									to={content.path}
									key={content.id}
									className="capitalize text-muted-foreground cursor-pointer hover:text-accent-brand"
								>
									{t(content.name)}
								</Link>
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
