import { RouteMessages } from "@/shared/i18n/messages";

export default function AdminGroupLayout({ children }: { children: React.ReactNode }) {
	return <RouteMessages group="admin">{children}</RouteMessages>;
}
