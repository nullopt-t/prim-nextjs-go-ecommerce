"use client";

import {
  FaInstagram,
  FaSquareFacebook,
  FaSquareXTwitter,
  FaTiktok,
} from "react-icons/fa6";
import { Brand } from "@/components/ui/Brand";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

const currentDate = new Date().getFullYear();

export function Footer() {
  const t = useTranslations("common");

  const sections = [
    {
      id: "shop-pp",
      title: "footer.shop",
      content: [
        {
          id: "cat_1",
          slug: "audio",
          name: "Audio",
          path: "/products?category=audio",
        },
        {
          id: "cat_2",
          slug: "wearables",
          name: "Wearables",
          path: "/products?category=wearables",
        },
        {
          id: "cat_3",
          slug: "desk-setup",
          name: "Desk Setup",
          path: "/products?category=desk-setup",
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
          path: "/about",
        },
        {
          id: "supp_2",
          name: "footer.faqs",
          path: "/about",
        },
        {
          id: "supp_3",
          name: "footer.shipping",
          path: "/about",
        },
        {
          id: "supp_4",
          name: "footer.returns",
          path: "/about",
        },
        {
          id: "supp_5",
          name: "footer.trackOrders",
          path: "/user/orders",
        },
        {
          id: "supp_6",
          name: "footer.sizeGuide",
          path: "/about",
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
          path: "/about",
        },
        {
          id: "comp_3",
          name: "footer.press",
          path: "/about",
        },
        {
          id: "comp_4",
          name: "footer.sustainability",
          path: "/about",
        },
        {
          id: "comp_5",
          name: "footer.blog",
          path: "/about",
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
    <footer className="bg-footer p-2.5 md:p-5 lg:p-10 text-white mt-auto">
      <div className="mb-5 flex flex-col gap-10 md:flex-row text-white text-txt-sm md:text-txt-md lg:text-txt-lg">
        <div className="flex-1">
          <Brand />
          <p className="text-muted-foreground mt-2">
            {t("footer.footerDescription")}
          </p>
        </div>
        {sections.map((value) => (
          <div key={value.id} className="flex-1">
            <p className="mb-5 text-txt-lg font-medium capitalize">
              {t(value.title)}
            </p>
            <div className="flex flex-col gap-2.5">
              {value.content.map((content) => (
                <Link
                  href={content.path}
                  key={content.id}
                  className="capitalize text-muted-foreground cursor-pointer hover:text-accent-brand transition-colors"
                >
                  {content.name.startsWith("footer.")
                    ? t(content.name)
                    : content.name}
                </Link>
              ))}
            </div>
          </div>
        ))}
      </div>
      <hr className="border-border text-muted-foreground" />
      <div className="flex flex-col gap-5 md:flex-row md:justify-between text-muted-foreground pt-5">
        <p>{t("footer.copyright", { year: currentDate })}</p>

        <div className="flex gap-2.5">
          {social.map((value) => (
            <value.icon
              key={value.id}
              className="size-6 hover:text-accent-brand cursor-pointer transition-colors"
            />
          ))}
        </div>
      </div>
    </footer>
  );
}
