"use client";

import { useEffect, useRef, type TableHTMLAttributes } from "react";

/**
 * A table that collapses into plain cards on small screens.
 *
 * On screens below the `sm` breakpoint the header row is hidden and every cell
 * is rendered as a "label / value" row using the header text captured from the
 * matching `<th>` via a `data-label` attribute. This keeps dense data tables
 * readable on phones without duplicating card markup for every list.
 */
export function ResponsiveTable({
  children,
  className = "",
  ...props
}: TableHTMLAttributes<HTMLTableElement>) {
  const ref = useRef<HTMLTableElement>(null);

  useEffect(() => {
    const table = ref.current;
    if (!table) return;

    const labels = Array.from(
      table.querySelectorAll<HTMLTableCellElement>("thead th")
    ).map((th) => th.textContent?.trim() ?? "");

    table
      .querySelectorAll<HTMLTableRowElement>("tbody tr")
      .forEach((row) => {
        Array.from(row.children).forEach((cell, index) => {
          if (cell.tagName !== "TD" || cell.hasAttribute("colspan")) {
            cell.removeAttribute("data-label");
            return;
          }
          const label = labels[index];
          if (label) {
            cell.setAttribute("data-label", label);
          } else {
            cell.removeAttribute("data-label");
          }
        });
      });
  });

  return (
    <table ref={ref} className={`rtable ${className}`.trim()} {...props}>
      {children}
    </table>
  );
}
