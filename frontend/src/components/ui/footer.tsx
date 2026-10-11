import { Brand } from "@/components/ui/Brand";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

const currentDate = new Date().getFullYear();

export function Footer() {
  const t = useTranslations("common");

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
        {
          id: "cat_wishlist",
          nameKey: "footer.wishlist",
          path: "/wishlist",
        },
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
          path: "/user/orders",
        },
        {
          id: "comp_vendor",
          nameKey: "footer.vendorPortal",
          path: "/vendor",
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
              {t("footer.footerDescription")}
            </p>
          </div>
          {sections.map((value) => (
            <div key={value.id} className="flex-1">
              <p className="mb-4 text-base font-bold text-foreground">
                {t(value.title)}
              </p>
              <div className="flex flex-col gap-2.5">
                {value.content.map((content) => (
                  <Link
                    href={content.path}
                    key={content.id}
                    className="text-muted-foreground hover:text-accent-brand text-sm transition-colors cursor-pointer"
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
          <p>{t("footer.copyright", { year: currentDate })}</p>
        </div>
      </div>
    </footer>
  );
}
