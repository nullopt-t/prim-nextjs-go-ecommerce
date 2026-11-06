import type { Metadata } from "next";
import {
  ArrowRight,
  BadgePercent,
  BarChart3,
  Check,
  ShieldCheck,
  Store,
  Truck,
} from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("marketing.meta");
  return { title: t("title"), description: t("description") };
}

export default async function SellPage() {
  const t = await getTranslations("marketing");

  const benefits = [
    { icon: BarChart3, key: "reach" },
    { icon: BadgePercent, key: "commission" },
    { icon: ShieldCheck, key: "payouts" },
    { icon: Truck, key: "operations" },
  ] as const;

  const steps = ["apply", "review", "list", "getPaid"] as const;
  const faqs = ["approval", "fees", "payouts", "suspended"] as const;

  return (
    <main>
      <section className="border-b border-border bg-gradient-to-b from-secondary/50 to-background">
        <div className="mx-auto grid max-w-6xl gap-10 px-4 py-16 sm:px-6 lg:grid-cols-[1.15fr_0.85fr] lg:items-center lg:py-24">
          <div>
            <div className="inline-flex items-center gap-2 rounded-full border border-border bg-background px-3 py-1 text-xs font-medium text-muted-foreground">
              <Store className="size-3.5 text-accent-brand" />
              {t("hero.eyebrow")}
            </div>
            <h1 className="mt-5 max-w-2xl text-4xl font-semibold leading-[1.08] tracking-tight sm:text-5xl">
              {t("hero.title")}
            </h1>
            <p className="mt-5 max-w-xl text-base leading-7 text-muted-foreground">
              {t("hero.description")}
            </p>
            <div className="mt-8 flex flex-wrap items-center gap-3">
              <Link
                href="/seller"
                className="inline-flex h-11 items-center gap-2 rounded-lg bg-accent-brand px-5 text-sm font-semibold text-white hover:bg-accent-brand/90"
              >
                {t("hero.cta")}
                <ArrowRight className="size-4 rtl:rotate-180" />
              </Link>
              <Link
                href="/auth?redirect=%2Fseller"
                className="inline-flex h-11 items-center rounded-lg border border-border bg-background px-5 text-sm font-medium hover:bg-secondary"
              >
                {t("hero.secondaryCta")}
              </Link>
            </div>
            <ul className="mt-7 flex flex-wrap gap-x-5 gap-y-2 text-xs text-muted-foreground">
              {(["free", "approval", "dashboard"] as const).map((item) => (
                <li key={item} className="flex items-center gap-1.5">
                  <Check className="size-3.5 text-emerald-600" />
                  {t(`hero.points.${item}`)}
                </li>
              ))}
            </ul>
          </div>

          <div className="rounded-2xl border border-border bg-background p-5 shadow-sm">
            <div className="flex items-center justify-between border-b border-border pb-4">
              <div>
                <div className="text-sm font-semibold">{t("preview.title")}</div>
                <div className="text-xs text-muted-foreground">{t("preview.subtitle")}</div>
              </div>
              <span className="rounded-full bg-emerald-500/10 px-2.5 py-1 text-[11px] font-medium text-emerald-700">
                {t("preview.status")}
              </span>
            </div>
            <dl className="mt-4 space-y-3 text-sm">
              {(["sales", "commission", "net", "nextPayout"] as const).map((row, index) => (
                <div key={row} className="flex items-center justify-between gap-4">
                  <dt className={index === 2 ? "font-medium" : "text-muted-foreground"}>
                    {t(`preview.rows.${row}.label`)}
                  </dt>
                  <dd
                    className={`tabular-nums ${index === 2 ? "text-lg font-semibold" : "font-medium"}`}
                  >
                    {t(`preview.rows.${row}.value`)}
                  </dd>
                </div>
              ))}
            </dl>
            <div className="mt-5 rounded-xl border border-border bg-muted/30 p-3 text-xs text-muted-foreground">
              {t("preview.note")}
            </div>
          </div>
        </div>
      </section>

      <section className="border-b border-border">
        <div className="mx-auto grid max-w-6xl gap-px bg-border px-4 py-10 sm:grid-cols-2 sm:px-6 lg:grid-cols-4">
          {(["commission", "listing", "settlement", "enterprise"] as const).map((term) => (
            <div key={term} className="bg-background p-4 first:rounded-s-xl last:rounded-e-xl">
              <div className="text-lg font-semibold tracking-tight">
                {t(`terms.items.${term}.value`)}
              </div>
              <div className="mt-1 text-xs text-muted-foreground">
                {t(`terms.items.${term}.label`)}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-4 py-16 sm:px-6">
        <div className="max-w-2xl">
          <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            {t("benefits.title")}
          </h2>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {t("benefits.description")}
          </p>
        </div>
        <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {benefits.map(({ icon: Icon, key }) => (
            <article key={key} className="rounded-xl border border-border bg-background p-5">
              <div className="flex size-9 items-center justify-center rounded-lg bg-secondary">
                <Icon className="size-4" />
              </div>
              <h3 className="mt-4 text-sm font-semibold">{t(`benefits.items.${key}.title`)}</h3>
              <p className="mt-1.5 text-xs leading-5 text-muted-foreground">
                {t(`benefits.items.${key}.description`)}
              </p>
            </article>
          ))}
        </div>
      </section>

      <section className="border-y border-border bg-muted/20">
        <div className="mx-auto max-w-6xl px-4 py-16 sm:px-6">
          <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            {t("steps.title")}
          </h2>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
            {t("steps.description")}
          </p>
          <ol className="mt-8 grid gap-4 md:grid-cols-4">
            {steps.map((step, index) => (
              <li key={step} className="rounded-xl border border-border bg-background p-5">
                <div className="flex size-7 items-center justify-center rounded-full bg-foreground text-xs font-semibold text-background">
                  {index + 1}
                </div>
                <h3 className="mt-3 text-sm font-semibold">{t(`steps.items.${step}.title`)}</h3>
                <p className="mt-1.5 text-xs leading-5 text-muted-foreground">
                  {t(`steps.items.${step}.description`)}
                </p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      <section className="mx-auto grid max-w-6xl gap-10 px-4 py-16 sm:px-6 lg:grid-cols-2">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            {t("faq.title")}
          </h2>
          <div className="mt-6 divide-y divide-border rounded-xl border border-border">
            {faqs.map((faq) => (
              <details key={faq} className="group p-4 open:bg-muted/20">
                <summary className="cursor-pointer list-none text-sm font-medium marker:hidden">
                  <span className="flex items-center justify-between gap-3">
                    {t(`faq.items.${faq}.question`)}
                    <span className="text-muted-foreground transition-transform group-open:rotate-45">
                      +
                    </span>
                  </span>
                </summary>
                <p className="mt-2 text-xs leading-5 text-muted-foreground">
                  {t(`faq.items.${faq}.answer`)}
                </p>
              </details>
            ))}
          </div>
        </div>

        <div className="flex flex-col justify-center rounded-2xl border border-border bg-foreground p-8 text-background">
          <h2 className="text-2xl font-semibold tracking-tight">{t("closing.title")}</h2>
          <p className="mt-3 text-sm leading-6 text-background/70">{t("closing.description")}</p>
          <div className="mt-6">
            <Link
              href="/seller"
              className="inline-flex h-11 items-center gap-2 rounded-lg bg-background px-5 text-sm font-semibold text-foreground hover:bg-background/90"
            >
              {t("closing.cta")}
              <ArrowRight className="size-4 rtl:rotate-180" />
            </Link>
          </div>
          <p className="mt-4 text-xs text-background/60">{t("closing.note")}</p>
        </div>
      </section>
    </main>
  );
}
