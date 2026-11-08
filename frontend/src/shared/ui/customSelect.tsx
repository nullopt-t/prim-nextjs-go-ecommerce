"use client";

import { forwardRef } from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "@/shared/lib/cn";

export interface CustomSelectOption {
	value: string;
	label: string;
	disabled?: boolean;
}

interface CustomSelectProps {
	value: string;
	onChange: (value: string) => void;
	options: CustomSelectOption[];
	/** Shown when no option matches the current value (e.g. value === ""). */
	placeholder?: string;
	disabled?: boolean;
	/** Wrapper sizing (e.g. "w-full", "sm:w-auto"). */
	className?: string;
	size?: "sm" | "md";
	id?: string;
	name?: string;
	ariaLabel?: string;
}

const sizeClasses = {
	sm: "px-2.5 py-1.5 text-xs",
	md: "px-3 py-2 text-sm",
} as const;

/**
 * The one shared select for the app: a native `<select>` (fully
 * keyboard/screen-reader accessible, empty-string values work) dressed with
 * the design-system tokens — bg-background, border-border, accent focus ring,
 * and a chevron. `onChange` receives the string value directly; `ref`
 * forwards to the underlying element for react-hook-form registration.
 */
export const CustomSelect = forwardRef<HTMLSelectElement, CustomSelectProps>(function CustomSelect(
	{ value, onChange, options, placeholder, disabled, className, size = "md", id, name, ariaLabel },
	ref,
) {
	const hasValue = options.some((option) => option.value === value);
	return (
		<span className={cn("relative inline-flex max-w-full", className)}>
			<select
				ref={ref}
				id={id}
				name={name}
				aria-label={ariaLabel}
				className={cn(
					"w-full min-w-0 max-w-full cursor-pointer appearance-none rounded-md border border-border bg-background pe-8 text-foreground",
					"focus:outline-none focus:ring-2 focus:ring-accent-brand/30",
					"disabled:cursor-not-allowed disabled:opacity-50",
					sizeClasses[size],
				)}
				value={value}
				disabled={disabled}
				onChange={(event) => onChange(event.target.value)}
			>
				{placeholder && !hasValue ? (
					<option value="" disabled hidden>
						{placeholder}
					</option>
				) : null}
				{options.map((option) => (
					<option key={option.value} value={option.value} disabled={option.disabled}>
						{option.label}
					</option>
				))}
			</select>
			<ChevronDown
				className="pointer-events-none absolute end-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
				aria-hidden="true"
			/>
		</span>
	);
});
