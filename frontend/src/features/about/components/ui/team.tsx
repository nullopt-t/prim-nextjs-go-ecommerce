"use client";

import AmrImg from "@/assets/imgs/team/Amr.jpeg";
import MohamedImg from "@/assets/imgs/team/Mohamed.jpg";
import { FaLinkedinIn, FaGithub } from "react-icons/fa6";
import Contributors, { Contributor } from "@/features/about/components/ui/contributors";
import { Title } from "@/components/ui/title";
import { useTranslations } from "next-intl";

export default function Team() {
  const t = useTranslations("about");

  const ContArr: Contributor[] = [
    {
      id: "m-mahmoud-alsaid",
      name: "Mohamed Mahmoud",
      img: MohamedImg,
      imgAlt: "Mohamed photo",
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
      imgAlt: "Amr photo",
      role: "developers.backend",
      links: [
        {
          id: "nullopt-t-linkedin",
          link: "https://linkedin.com",
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
    <div className="mt-8">
      <Title title={t("developers.title")} />
      <div className="flex flex-wrap gap-12 mt-4">
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
