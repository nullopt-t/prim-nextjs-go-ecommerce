"use client";

import { CustomButton } from "@/components/ui/button";
import { useTranslations } from "next-intl";

export default function DangerZone() {
  const t = useTranslations("settings");

  const dangerActions = [
    {
      id: "danger-logout",
      buttonText: "logout",
    },
    {
      id: "danger-delete-acc",
      buttonText: "deleteAccount",
    },
  ];

  return (
    <div className="flex flex-col sm:flex-row gap-3 border-t border-b border-destructive/50 border-dashed py-5">
      {dangerActions.map((action) => (
        <div
          key={action.id}
          className="bg-destructive text-destructive-foreground rounded-md hover:opacity-90 w-fit"
        >
          <CustomButton text={t(`settings.${action.buttonText}`)} />
        </div>
      ))}
    </div>
  );
}
