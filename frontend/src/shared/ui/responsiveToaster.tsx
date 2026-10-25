"use client";

import { Toaster } from "sonner";
import { useEffect, useState } from "react";

/**
 * Renders Sonner toasts at a position that avoids colliding with the
 * mobile bottom navigation bar.
 *
 * - Phones (< 768px):  top-right  (keeps bottom nav reachable)
 * - Tablet+ (≥ 768px): bottom-right (desktop convention)
 */
export function ResponsiveToaster({
  richColors = true,
  className = "",
}: {
  richColors?: boolean;
  className?: string;
}) {
  const [position, setPosition] = useState<"top-right" | "bottom-right">(
    "bottom-right"
  );

  useEffect(() => {
    const update = () => {
      setPosition(window.innerWidth < 768 ? "top-right" : "bottom-right");
    };
    update();
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, []);

  return (
    <Toaster
      position={position}
      richColors={richColors}
      className={className}
      closeButton={true}
    />
  );
}
