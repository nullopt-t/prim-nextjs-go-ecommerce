import React from "react";

interface FilterCheckboxProps {
  labelTxt: string;
  checked?: boolean;
  onChange?: (checked: boolean) => void;
}

export default function FilterCheckbox({
  labelTxt,
  checked,
  onChange,
}: FilterCheckboxProps) {
  return (
    <label className="flex gap-2.5 items-center group cursor-pointer select-none">
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange?.(e.target.checked)}
        className="size-4 accent-accent-brand rounded"
      />
      <span className="text-muted-foreground group-hover:text-accent-brand text-txt-sm md:text-txt-md transition-colors">
        {labelTxt}
      </span>
    </label>
  );
}
