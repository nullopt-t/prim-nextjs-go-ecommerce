import UserSideBar from "@/features/user/components/ui/sideBar";
import MobileNav from "@/features/user/components/ui/mobileNav";

export default function UserLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8 pb-24 md:pb-8">
        {/* Desktop: sidebar + content */}
        <div className="flex flex-row gap-6">
          <UserSideBar />
          <main className="flex-1 min-w-0">{children}</main>
        </div>
      </div>
      {/* Fixed bottom nav — mobile only */}
      <MobileNav />
    </>
  );
}

