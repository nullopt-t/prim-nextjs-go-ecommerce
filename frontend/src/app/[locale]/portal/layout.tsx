import { RouteMessages } from "@/shared/i18n/messages";

export default function PortalLayout({ children }: { children: React.ReactNode }) {
	return <RouteMessages group="portal">{children}</RouteMessages>;
}
