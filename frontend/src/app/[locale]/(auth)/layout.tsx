import AuthLayout from "@/modules/auth/components/layouts/authLayout";
import { RouteMessages } from "@/shared/i18n/messages";

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <RouteMessages group="auth">
      <AuthLayout>{children}</AuthLayout>
    </RouteMessages>
  );
}
