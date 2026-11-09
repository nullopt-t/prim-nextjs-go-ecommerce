"use client";

import { Brand } from "@/shared/ui/Brand";
import { StoreContactInfo } from "@/shared/ui/storeContactInfo";
import { CurrencySwitcher } from "@/shared/ui/currencySwitcher";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { useTranslations } from "next-intl";
import { Link } from "@/shared/i18n/navigation";

const currentDate = new Date().getFullYear();

export function Footer() {
  const t = useTranslations("common");
  const { settings } = usePublicSettings();
  const { isAuthenticated } = useAuthContext();
  const storeName = settings?.storeName?.trim() || t("footer.defaultStore");

  const sections = [
    {
      id: "shop-section",
      title: "footer.shop",
      content: [
        {
          id: "cat_all",
          nameKey: "footer.allProducts",
          path: "/products",
        },
        {
          id: "cat_cats",
          nameKey: "footer.categories",
          path: "/categories",
        },
        ...(isAuthenticated
          ? [
              {
                id: "cat_wishlist",
                nameKey: "footer.wishlist",
                path: "/wishlist",
              },
            ]
          : []),
      ],
    },
    {
      id: "support-section",
      title: "footer.support",
      content: [
        {
          id: "supp_contact",
          nameKey: "footer.contactUs",
          path: "/contact",
        },
        {
          id: "supp_faqs",
          nameKey: "footer.faqs",
          path: "/faqs",
        },
        {
          id: "supp_shipping",
          nameKey: "footer.shipping",
          path: "/shipping",
        },
        {
          id: "supp_returns",
          nameKey: "footer.returns",
          path: "/returns",
        },
      ],
    },
    {
      id: "platform-section",
      title: "footer.company",
      content: [
        {
          id: "comp_about",
          nameKey: "footer.aboutUs",
          path: "/about",
        },
        {
          id: "comp_orders",
          nameKey: "footer.trackOrders",
          path: isAuthenticated ? "/user/orders" : "/order-lookup",
        },
        {
          id: "comp_sell",
          nameKey: "footer.sellOnPrim",
          path: "/sell",
        },
        {
          id: "comp_seller",
          nameKey: "footer.sellerPortal",
          path: "/seller",
        },
        {
          id: "comp_careers",
          nameKey: "footer.careers",
          path: "/careers",
        },
        {
          id: "comp_press",
          nameKey: "footer.press",
          path: "/press",
        },
      ],
    },
  ];

  return (
    <footer className="bg-secondary/40 text-foreground mt-auto border-t border-border/60 transition-colors">
      <div className="w-full px-3 sm:px-6 lg:px-8 py-10 md:py-14">
        <div className="mb-10 flex flex-col gap-10 md:flex-row text-txt-sm md:text-txt-md lg:text-txt-lg">
          <div className="flex-1 max-w-sm">
            <Brand />
            <p className="text-muted-foreground mt-3 text-sm leading-relaxed">
              {t("footer.footerDescription", { store: storeName })}
            </p>
            <StoreContactInfo />
            <div className="mt-4">
              <span className="mb-2 block text-xs font-semibold text-muted-foreground">
                {t("currency.label")}
              </span>
              <CurrencySwitcher variant="footer" />
            </div>
          </div>
          {sections.map((value) => (
            <div key={value.id} className="flex-1">
              <h2 className="mb-4 text-base font-bold text-foreground">
                {t(value.title)}
              </h2>
              <div className="flex flex-col gap-2.5">
                {value.content.map((content) => (
                  <Link
                    href={content.path}
                    key={content.id}
                    className="text-muted-foreground hover:text-accent-brand text-sm transition-colors cursor-pointer py-1.5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-brand"
                  >
                    {t(content.nameKey)}
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
        <hr className="border-border/60 text-muted-foreground mb-6" />
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between text-muted-foreground text-xs sm:text-sm">
          <p>{t("footer.copyright", { year: currentDate, store: storeName })}</p>
        </div>
      </div>
    </footer>
  );
}
