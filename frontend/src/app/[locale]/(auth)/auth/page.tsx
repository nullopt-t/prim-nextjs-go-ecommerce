import { LoginView } from "@/features/auth";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Sign In | PRIM",
  description: "Sign in to your PRIM account",
};

export default function LoginPage() {
  return <LoginView />;
}
