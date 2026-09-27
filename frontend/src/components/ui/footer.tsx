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
      name: "Instagram",
      href: "https://instagram.com",
      icon: FaInstagram,
    },
    {
      id: "x-1",
      name: "X (Twitter)",
      href: "https://x.com",
      icon: FaSquareXTwitter,
    },
    {
      id: "tik-1",
      name: "TikTok",
      href: "https://tiktok.com",
      icon: FaTiktok,
    },
    {
      id: "facebook-1",
      name: "Facebook",
      href: "https://facebook.com",
      icon: FaSquareFacebook,
    },
  ];

  return (
    <footer className="bg-footer text-white mt-auto border-t border-border/20">
      <div className="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-10 py-10 md:py-14">
        <div className="mb-10 flex flex-col gap-10 md:flex-row text-white text-txt-sm md:text-txt-md lg:text-txt-lg">
          <div className="flex-1 max-w-sm">
            <Brand />
            <p className="text-muted-foreground mt-3 text-sm leading-relaxed">
              {t("footer.footerDescription")}
            </p>
          </div>
          {sections.map((value) => (
            <div key={value.id} className="flex-1">
              <p className="mb-4 text-base font-bold capitalize text-white">
                {t(value.title)}
              </p>
              <div className="flex flex-col gap-2.5">
                {value.content.map((content) => (
                  <Link
                    href={content.path}
                    key={content.id}
                    className="capitalize text-muted-foreground hover:text-accent-brand text-sm transition-colors cursor-pointer"
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
        <hr className="border-border/40 text-muted-foreground mb-6" />
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between text-muted-foreground text-xs sm:text-sm">
          <p>{t("footer.copyright", { year: currentDate })}</p>

          <div className="flex gap-4">
            {social.map((value) => (
              <a
                key={value.id}
                href={value.href}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={value.name}
                className="text-muted-foreground hover:text-accent-brand transition-colors"
              >
                <value.icon className="size-5" />
              </a>
            ))}
          </div>
        </div>
      </div>
    </footer>
  );
}
