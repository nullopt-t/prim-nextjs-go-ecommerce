import { LoginView } from "@/features/auth";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.auth.signIn");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function LoginPage() {
  return <LoginView />;
}
