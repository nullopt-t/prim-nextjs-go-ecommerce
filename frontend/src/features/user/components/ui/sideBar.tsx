import UserSideBarLinks from "@/features/user/components/ui/sideBarLinks";
import UserSideBarProfile from "@/features/user/components/ui/sideBarProfile";

export default function UserSideBar() {
  return (
    <aside className="w-14 md:w-60 shrink-0 bg-sidebar border-r border-sidebar-border p-2 md:p-4 min-h-[70vh]">
      <UserSideBarProfile />
      <UserSideBarLinks />
    </aside>
  );
}
