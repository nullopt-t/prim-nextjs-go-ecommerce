"use client";

import * as React from "react";
import * as SelectPrimitive from "@radix-ui/react-select";
import { Check, ChevronDown, ChevronUp } from "lucide-react";
import { cn } from "@/shared/lib/cn";

export const Select = SelectPrimitive.Root;
export const SelectGroup = SelectPrimitive.Group;
export const SelectValue = SelectPrimitive.Value;

export const SelectTrigger = React.forwardRef<
	React.ComponentRef<typeof SelectPrimitive.Trigger>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
	<SelectPrimitive.Trigger
		ref={ref}
		className={cn(
			"flex h-10 w-full items-center justify-between gap-2 rounded-xl border border-border bg-background px-3.5 text-xs sm:text-sm text-foreground",
			"placeholder:text-muted-foreground focus:outline-none focus:border-accent-brand focus:ring-1 focus:ring-accent-brand",
			"disabled:opacity-60 disabled:cursor-not-allowed transition",
			"aria-invalid:border-destructive",
			className,
		)}
		{...props}
	>
		{children}
		<SelectPrimitive.Icon asChild>
			<ChevronDown className="size-4 text-muted-foreground" />
		</SelectPrimitive.Icon>
	</SelectPrimitive.Trigger>
));
SelectTrigger.displayName = SelectPrimitive.Trigger.displayName;

export const SelectContent = React.forwardRef<
	React.ComponentRef<typeof SelectPrimitive.Content>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = "popper", ...props }, ref) => (
	<SelectPrimitive.Portal>
		<SelectPrimitive.Content
			ref={ref}
			position={position}
			className={cn(
				"relative z-50 max-h-72 min-w-[8rem] overflow-hidden rounded-xl border border-border bg-card text-foreground shadow-xl",
				"data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95",
				position === "popper" && "translate-y-1",
				className,
			)}
			{...props}
		>
			<SelectPrimitive.ScrollUpButton className="flex h-6 items-center justify-center">
				<ChevronUp className="size-4" />
			</SelectPrimitive.ScrollUpButton>
			<SelectPrimitive.Viewport className="p-1">
				{children}
			</SelectPrimitive.Viewport>
			<SelectPrimitive.ScrollDownButton className="flex h-6 items-center justify-center">
				<ChevronDown className="size-4" />
			</SelectPrimitive.ScrollDownButton>
		</SelectPrimitive.Content>
	</SelectPrimitive.Portal>
));
SelectContent.displayName = SelectPrimitive.Content.displayName;

export const SelectItem = React.forwardRef<
	React.ComponentRef<typeof SelectPrimitive.Item>,
	React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
	<SelectPrimitive.Item
		ref={ref}
		className={cn(
			"relative flex w-full cursor-pointer select-none items-center rounded-lg py-2 ps-8 pe-2.5 text-xs sm:text-sm outline-none",
			"focus:bg-secondary data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute start-2.5 flex size-3.5 items-center justify-center">
			<SelectPrimitive.ItemIndicator>
				<Check className="size-4 text-accent-brand" />
			</SelectPrimitive.ItemIndicator>
		</span>
		<SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
	</SelectPrimitive.Item>
));
SelectItem.displayName = SelectPrimitive.Item.displayName;

export function SelectLabel({
	className,
	...props
}: React.ComponentPropsWithoutRef<typeof SelectPrimitive.Label>) {
	return (
		<SelectPrimitive.Label
			className={cn("px-2 py-1.5 text-[11px] font-semibold text-muted-foreground", className)}
			{...props}
		/>
	);
}

export function SelectSeparator({
	className,
	...props
}: React.ComponentPropsWithoutRef<typeof SelectPrimitive.Separator>) {
	return (
		<SelectPrimitive.Separator
			className={cn("-mx-1 my-1 h-px bg-border", className)}
			{...props}
		/>
	);
}
