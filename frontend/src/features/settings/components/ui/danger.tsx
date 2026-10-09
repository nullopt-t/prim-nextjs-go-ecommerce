"use client";

import { useAuthContext } from "@/context/AuthContext";
import { useRouter } from "@/i18n/navigation";
interface DangerRowProps {
  title: string;
  description: string;
  action: React.ReactNode;
}

function DangerRow({ title, description, action }: DangerRowProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 py-4 border-b border-border last:border-0">
      <div>
        <p className="text-sm font-medium text-foreground">{title}</p>
        <p className="text-xs text-muted-foreground mt-0.5 max-w-sm">{description}</p>
      </div>
      <div className="sm:shrink-0">{action}</div>
    </div>
  );
}

export default function DangerZone() {
  const { logout } = useAuthContext();
  const router = useRouter();

  const handleLogoutAll = async () => {
    await logout();
    router.push("/auth");
  };

  return (
    <div className="flex flex-col">
      <DangerRow
        title="Sign Out Everywhere"
        description="Remove all active sessions from all devices."
        action={
          <button
            type="button"
            onClick={handleLogoutAll}
            className="h-8 px-3 rounded-md text-xs font-medium border border-destructive text-destructive hover:bg-destructive/10 transition"
          >
            Sign Out All
          </button>
        }
      />
      <DangerRow
        title="Delete Account"
        description="Permanently delete your account and all data. This cannot be undone."
        action={
          <button
            type="button"
            disabled
            title="Account deletion is not yet available. Contact support."
            className="h-8 px-3 rounded-md text-xs font-medium bg-destructive text-destructive-foreground opacity-50 cursor-not-allowed transition"
          >
            Delete Account
          </button>
        }
      />
    </div>
  );
}
