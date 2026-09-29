"use client";

import { useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { Lock } from "lucide-react";

export default function PersonalInfoForm() {
  const { user } = useAuthContext();

  const [displayName, setDisplayName] = useState(user?.name || "");

  return (
    <form
      className="flex flex-col gap-5 max-w-xl"
      onSubmit={(e) => e.preventDefault()}
    >
      {/* Display Name */}
      <div className="flex flex-col gap-1.5">
        <label htmlFor="display-name" className="text-sm font-medium text-foreground">
          Display Name
        </label>
        <input
          id="display-name"
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder="Your display name"
          className="h-10 w-full rounded-md border border-border bg-input-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-accent-brand/30 focus:border-accent-brand/60 transition"
        />
      </div>

      {/* Email (disabled) */}
      <div className="flex flex-col gap-1.5">
        <label htmlFor="email" className="text-sm font-medium text-foreground">
          Email Address
        </label>
        <div className="relative">
          <input
            id="email"
            type="email"
            value={user?.email || ""}
            readOnly
            disabled
            className="h-10 w-full rounded-md border border-border bg-input-background px-3 pr-10 text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-60 disabled:cursor-not-allowed focus:outline-none transition"
          />
          <Lock className="absolute right-3 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/60" />
        </div>
        <p className="text-xs text-muted-foreground">
          Email cannot be changed. Contact support to update.
        </p>
      </div>

      {/* Phone (disabled) */}
      <div className="flex flex-col gap-1.5">
        <label htmlFor="phone" className="text-sm font-medium text-foreground">
          Phone Number
        </label>
        <input
          id="phone"
          type="tel"
          value={user?.phone || ""}
          readOnly
          disabled
          placeholder="Not set"
          className="h-10 w-full rounded-md border border-border bg-input-background px-3 text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-60 disabled:cursor-not-allowed focus:outline-none transition"
        />
        <p className="text-xs text-muted-foreground">
          Contact support to update your phone number.
        </p>
      </div>

      {/* Submit */}
      <div className="flex justify-end">
        <button
          type="submit"
          className="h-9 px-4 rounded-md bg-accent-brand text-white text-sm font-medium hover:bg-accent-brand/90 transition shadow-sm"
        >
          Save Changes
        </button>
      </div>
    </form>
  );
}
