import { VerifyView } from "@/features/auth";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Verify Account | PRIM",
  description: "Verify your email to continue",
};

export default function VerifyPage() {
  return <VerifyView />;
}
