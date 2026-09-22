import UserSideBar from "@/features/user/components/ui/sideBar";

export default function UserLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col md:flex-row gap-5 border border-border rounded-xl overflow-hidden bg-background">
      <UserSideBar />
      <div className="flex-1 p-4 md:p-8 overflow-y-auto">{children}</div>
    </div>
  );
}
