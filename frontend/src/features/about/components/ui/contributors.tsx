"use client";

import Image, { StaticImageData } from "next/image";
import { useTranslations } from "next-intl";
import { IconType } from "react-icons";

export interface ContributorLink {
  id: string;
  link: string;
  icon: IconType;
}

export interface Contributor {
  id: string;
  name: string;
  img: StaticImageData | string;
  imgAlt: string;
  role: string;
  links: ContributorLink[];
}

export default function Contributors({ contObj }: { contObj: Contributor }) {
  const t = useTranslations("about");

  return (
    <div className="flex flex-col items-center text-center">
      <div className="relative w-20 h-20 md:w-24 md:h-24 rounded-full overflow-hidden border-2 border-border mb-3 bg-secondary">
        {typeof contObj.img === "string" ? (
          <img
            src={contObj.img}
            alt={contObj.imgAlt}
            className="w-full h-full object-cover object-center"
          />
        ) : (
          <Image
            src={contObj.img}
            alt={contObj.imgAlt}
            fill
            className="object-cover object-center"
          />
        )}
      </div>
      <p className="font-medium text-foreground text-txt-sm md:text-txt-md mb-1">
        {contObj.name}
      </p>
      <p className="text-xs text-muted-foreground mb-2">{t(contObj.role)}</p>
      <ul className="flex gap-3">
        {contObj.links.map((link) => {
          const Icon = link.icon;
          return (
            <li key={link.id}>
              <a
                href={link.link}
                target="_blank"
                rel="noreferrer"
                className="text-muted-foreground hover:text-accent-brand transition-colors"
              >
                <Icon className="size-4" />
              </a>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
