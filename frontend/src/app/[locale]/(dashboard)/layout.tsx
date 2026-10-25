import { UserLayout } from "@/modules/account";
import { RouteMessages } from "@/shared/i18n/messages";

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <RouteMessages group="dashboard">
      <UserLayout>{children}</UserLayout>
    </RouteMessages>
  );
}
