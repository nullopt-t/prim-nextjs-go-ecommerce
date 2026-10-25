"use client";

import * as React from "react";
import {
	Controller,
	FormProvider,
	useFormContext,
	type ControllerProps,
	type FieldPath,
	type FieldValues,
} from "react-hook-form";
import * as LabelPrimitive from "@radix-ui/react-label";
import { Slot } from "@radix-ui/react-slot";
import { cn } from "@/shared/lib/cn";

export const Form = FormProvider;

type FormFieldContextValue = { name: string };
const FormFieldContext = React.createContext<FormFieldContextValue>(
	{} as FormFieldContextValue,
);

type FormItemContextValue = { id: string };
const FormItemContext = React.createContext<FormItemContextValue>(
	{} as FormItemContextValue,
);

export function FormField<
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>(props: ControllerProps<TFieldValues, TName>) {
	return (
		<FormFieldContext.Provider value={{ name: props.name }}>
			<Controller {...props} />
		</FormFieldContext.Provider>
	);
}

export function useFormField() {
	const fieldContext = React.useContext(FormFieldContext);
	const itemContext = React.useContext(FormItemContext);
	const { getFieldState, formState } = useFormContext();

	if (!fieldContext.name) {
		throw new Error("useFormField must be used within <FormField>");
	}
	if (!itemContext.id) {
		throw new Error("useFormField must be used within <FormItem>");
	}

	const fieldState = getFieldState(fieldContext.name, formState);
	const { id } = itemContext;

	return {
		id,
		name: fieldContext.name,
		formItemId: `${id}-form-item`,
		formDescriptionId: `${id}-form-item-description`,
		formMessageId: `${id}-form-item-message`,
		...fieldState,
	};
}

export function FormItem({
	className,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	const id = React.useId();

	return (
		<FormItemContext.Provider value={{ id }}>
			<div className={cn("flex flex-col gap-1.5", className)} {...props} />
		</FormItemContext.Provider>
	);
}

export function FormLabel({
	className,
	...props
}: React.ComponentProps<typeof LabelPrimitive.Root>) {
	const { error, formItemId } = useFormField();

	return (
		<LabelPrimitive.Root
			htmlFor={formItemId}
			className={cn(
				"text-xs font-medium text-foreground",
				error && "text-destructive",
				className,
			)}
			{...props}
		/>
	);
}

export function FormControl({
	...props
}: React.ComponentProps<typeof Slot>) {
	const {
		error,
		formItemId,
		formDescriptionId,
		formMessageId,
	} = useFormField();

	return (
		<Slot
			id={formItemId}
			aria-describedby={
				error ? `${formDescriptionId} ${formMessageId}` : formDescriptionId
			}
			aria-invalid={!!error}
			{...props}
		/>
	);
}

export function FormDescription({
	className,
	...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
	const { formDescriptionId } = useFormField();

	return (
		<p
			id={formDescriptionId}
			className={cn("text-[11px] text-muted-foreground", className)}
			{...props}
		/>
	);
}

export function FormMessage({
	className,
	children,
	...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
	const { error, formMessageId } = useFormField();
	const body = error ? String(error.message ?? "") : children;

	if (!body) return null;

	return (
		<p
			id={formMessageId}
			className={cn("text-[11px] font-medium text-destructive", className)}
			{...props}
		>
			{body}
		</p>
	);
}

export const FormInput = React.forwardRef<
	HTMLInputElement,
	React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => (
	<input
		ref={ref}
		className={cn(
			"h-10 w-full rounded-xl border border-border bg-background px-3.5 text-xs sm:text-sm text-foreground",
			"placeholder:text-muted-foreground focus:outline-none focus:border-accent-brand focus:ring-1 focus:ring-accent-brand",
			"disabled:opacity-60 disabled:cursor-not-allowed transition",
			"aria-invalid:border-destructive aria-invalid:focus:border-destructive aria-invalid:focus:ring-destructive",
			className,
		)}
		{...props}
	/>
));
FormInput.displayName = "FormInput";

export const FormTextarea = React.forwardRef<
	HTMLTextAreaElement,
	React.TextareaHTMLAttributes<HTMLTextAreaElement>
>(({ className, ...props }, ref) => (
	<textarea
		ref={ref}
		className={cn(
			"min-h-24 w-full rounded-xl border border-border bg-background px-3.5 py-2.5 text-xs sm:text-sm text-foreground",
			"placeholder:text-muted-foreground focus:outline-none focus:border-accent-brand focus:ring-1 focus:ring-accent-brand",
			"disabled:opacity-60 disabled:cursor-not-allowed transition",
			"aria-invalid:border-destructive aria-invalid:focus:border-destructive aria-invalid:focus:ring-destructive",
			className,
		)}
		{...props}
	/>
));
FormTextarea.displayName = "FormTextarea";
