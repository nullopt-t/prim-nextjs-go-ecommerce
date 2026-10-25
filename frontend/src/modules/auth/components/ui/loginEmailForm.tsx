"use client";

import { useMemo } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Mail, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import FormTitle from "@/modules/auth/components/ui/formTitle";
import FormSubtitle from "@/modules/auth/components/ui/formSubtitle";
import {
	Form,
	FormControl,
	FormField,
	FormInput,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/shared/ui/form/form";

interface LoginEmailFormProps {
	onSubmit: (email: string) => Promise<void> | void;
	isLoading?: boolean;
}

export function LoginEmailForm({ onSubmit, isLoading = false }: LoginEmailFormProps) {
	const t = useTranslations("auth");

	const schema = useMemo(
		() =>
			z.object({
				email: z
					.string()
					.trim()
					.min(1, t("sign.emailRequired"))
					.pipe(z.email(t("sign.emailInvalid"))),
			}),
		[t],
	);

	const form = useForm<z.infer<typeof schema>>({
		resolver: zodResolver(schema),
		defaultValues: { email: "" },
	});

	const handleSubmit = form.handleSubmit(async (values) => {
		await onSubmit(values.email);
	});

	return (
		<div className="w-full flex flex-col gap-5">
			<div className="flex flex-col gap-1">
				<FormTitle type="login" />
				<FormSubtitle type="login" />
			</div>

			<Form {...form}>
				<form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
					<FormField
						control={form.control}
						name="email"
						render={({ field }) => (
							<FormItem>
								<FormLabel>{t("sign.emailLabel")}</FormLabel>
								<FormControl>
									<div className="relative flex items-center">
										<Mail className="pointer-events-none absolute start-3 size-4 text-muted-foreground" />
										<FormInput
											{...field}
											type="email"
											autoComplete="email"
											inputMode="email"
											dir="ltr"
											placeholder={t("sign.emailPlaceholder")}
											disabled={isLoading}
											className="h-11 ps-9.5"
										/>
									</div>
								</FormControl>
								<FormMessage />
							</FormItem>
						)}
					/>

					<div className="pt-1">
						<button
							type="submit"
							disabled={isLoading}
							className="w-full h-11 rounded-md bg-primary text-primary-foreground font-medium text-sm transition-colors hover:bg-primary/90 flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{isLoading && <Loader2 className="size-4 animate-spin" />}
							<span>{t("sign.signButton")}</span>
						</button>
					</div>
				</form>
			</Form>
		</div>
	);
}
