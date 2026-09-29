import UserSideBarLinks from "@/features/user/components/ui/sideBarLinks";
import UserSideBarProfile from "@/features/user/components/ui/sideBarProfile";

export default function UserSideBar() {
  return (
    <aside className="hidden md:flex w-64 shrink-0 flex-col sticky top-6 self-start bg-card border border-border rounded-lg overflow-hidden shadow-sm min-h-fit">
      <UserSideBarProfile />
      <div className="p-2">
        <UserSideBarLinks />
      </div>
    </aside>
  );
}
