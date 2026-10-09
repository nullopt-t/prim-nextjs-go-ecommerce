import UserSideBarLinks from "@/features/user/components/ui/sideBarLinks";
import UserSideBarProfile from "@/features/user/components/ui/sideBarProfile";

export default function UserSideBar() {
  return (
    <aside className="hidden md:flex w-64 shrink-0 flex-col sticky top-6 self-start bg-card border border-border rounded-xl overflow-hidden shadow-xs min-h-fit">
      <UserSideBarProfile />
      <div className="p-3">
        <UserSideBarLinks />
      </div>
    </aside>
  );
}
