import { SettingsLayout } from "@/features/settings";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Account Settings | PRIM",
  description: "Manage your personal information and preferences",
};

export default function SettingsPage() {
  return <SettingsLayout />;
}
