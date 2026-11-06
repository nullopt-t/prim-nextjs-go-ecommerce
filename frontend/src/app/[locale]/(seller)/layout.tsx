import { RouteMessages } from "@/shared/i18n/messages";

export default function SellerGroupLayout({ children }: { children: React.ReactNode }) {
	return <RouteMessages group="seller">{children}</RouteMessages>;
}
