import AmrImg from "@/assets/imgs/team/Amr.jpeg";
import MohamedImg from "@/assets/imgs/team/Mohamed.jpg";
import { FaLinkedinIn, FaGithub } from "react-icons/fa6";
import Contributors from "@/features/about/components/ui/contributors";
import { Title } from "@/components/ui";
import { useTranslation } from "react-i18next";

export default function Team() {
	const { t } = useTranslation("about");
	const ContArr = [
		{
			id: "m-mahmoud-alsaid",
			name: "Mohamed Mahmoud",
			img: MohamedImg,
			imgAlt: "Mohamed image",
			role: "developers.frontend",
			links: [
				{
					id: "m-mahmoud-alsaid-linkedin",
					link: "https://www.linkedin.com/in/m-mahmoud-alsaid/",
					icon: FaLinkedinIn,
				},
				{
					id: "m-mahmoud-alsaid-github",
					link: "https://github.com/m-mahmoud-alsaid",
					icon: FaGithub,
				},
			],
		},
		{
			id: "nullopt-t",
			name: "Amr",
			img: AmrImg,
			imgAlt: "Amr Image",
			role: "developers.backend",
			links: [
				{
					id: "nullopt-t-linkedin",
					link: "http:",
					icon: FaLinkedinIn,
				},
				{
					id: "nullopt-t-github",
					link: "https://github.com/nullopt-t",
					icon: FaGithub,
				},
			],
		},
	];

	return (
		<div className="">
			<Title title={t("developers.title")} />
			<div className="flex flex-wrap gap-15">
				{ContArr.map((contributorObj) => (
					<Contributors
						key={contributorObj.id}
						contObj={contributorObj}
					/>
				))}
			</div>
		</div>
	);
}
