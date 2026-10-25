import { RouteMessages } from "@/shared/i18n/messages";

export default function VendorGroupLayout({ children }: { children: React.ReactNode }) {
	return <RouteMessages group="vendor">{children}</RouteMessages>;
}
