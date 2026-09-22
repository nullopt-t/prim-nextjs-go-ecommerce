import AuthLayout from "@/features/auth/components/layouts/authLayout";

export default function Layout({ children }: { children: React.ReactNode }) {
  return <AuthLayout>{children}</AuthLayout>;
}
