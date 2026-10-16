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
      id: "nullopt-t",
      name: "Mohamed Mahmoud",
      img: MohamedImg,
      imgAlt: t("team.mohamedAlt"),
      role: "developers.frontend",
      links: [
        {
          id: "nullopt-t-linkedin",
          link: "https://www.linkedin.com/in/nullopt-t/",
          icon: FaLinkedinIn,
        },
        {
          id: "nullopt-t-github",
          link: "https://github.com/nullopt-t",
          icon: FaGithub,
        },
      ],
    },
    {
      id: "nullopt-t",
      name: "Amr",
      img: AmrImg,
      imgAlt: t("team.amrAlt"),
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
