import AmrImg from "@/assets/imgs/team/Amr.jpeg";
import MohamedImg from "@/assets/imgs/team/Mohamed.jpg";
import { FaLinkedinIn, FaGithub } from "react-icons/fa6";
import Contributors from "@/features/about/ui/contributors";
import Title from "@/components/ui/title";

export default function Team() {
	const ContArr = [
		{
			id: "m-mahmoud-alsaid",
			name: "Mohamed Mahmoud",
			img: MohamedImg,
			imgAlt: "Mohamed image",
			role: "Frontend developer",
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
			role: "Backend developer",
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
			<Title title="Developers" />
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
