import { OverviewContent } from "@/features/user";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Account Overview | PRIM",
  description: "User dashboard and account overview",
};

export default function OverviewPage() {
  return <OverviewContent />;
}
