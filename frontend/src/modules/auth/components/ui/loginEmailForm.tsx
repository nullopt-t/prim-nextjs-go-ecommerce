"use client";

import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2 } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import {
	parsePhoneNumberFromString,
	type CountryCode,
} from "libphonenumber-js";
import FormTitle from "@/modules/auth/components/ui/formTitle";
import FormSubtitle from "@/modules/auth/components/ui/formSubtitle";
import { getPhoneCountries } from "@/shared/lib/countries";
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
	onSubmit: (phone: string) => Promise<void> | void;
	isLoading?: boolean;
}

export function LoginEmailForm({ onSubmit, isLoading = false }: LoginEmailFormProps) {
	const t = useTranslations("auth");
	const locale = useLocale();
	const [country, setCountry] = useState<CountryCode>("EG");

	const countries = useMemo(() => getPhoneCountries(locale), [locale]);

	const schema = useMemo(
		() =>
			z.object({
				phone: z
					.string()
					.trim()
					.min(1, t("sign.emailRequired")),
			}),
		[t],
	);

	const form = useForm<z.infer<typeof schema>>({
		resolver: zodResolver(schema),
		defaultValues: { phone: "" },
	});

	const handleSubmit = form.handleSubmit(async (values) => {
		const parsed = parsePhoneNumberFromString(values.phone, country);
		if (!parsed?.isValid()) {
			form.setError("phone", { message: t("sign.emailInvalid") });
			return;
		}
		await onSubmit(parsed.number);
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
						name="phone"
						render={({ field }) => (
							<FormItem>
								<FormLabel>{t("sign.emailLabel")}</FormLabel>
								<FormControl>
									<div className="flex items-center rounded-xl border border-border bg-background transition focus-within:border-accent-brand focus-within:ring-1 focus-within:ring-accent-brand aria-invalid:border-destructive">
										<select
											aria-label={t("sign.countryCodeLabel")}
											value={country}
											onChange={(e) => setCountry(e.target.value as CountryCode)}
											disabled={isLoading}
											dir="ltr"
											className="h-10 shrink-0 cursor-pointer bg-transparent ps-3 pe-1 text-xs sm:text-sm font-medium text-foreground focus:outline-none disabled:cursor-not-allowed disabled:opacity-60"
										>
											{countries.map((c) => (
												<option key={c.code} value={c.code}>
													{`${c.dialCode} ${c.code}`}
												</option>
											))}
										</select>
										<span className="h-5 w-px shrink-0 bg-border" aria-hidden="true" />
										<FormInput
											{...field}
											type="tel"
											autoComplete="tel-national"
											inputMode="tel"
											dir="ltr"
											placeholder={t("sign.emailPlaceholder")}
											disabled={isLoading}
											className="h-10 rounded-none border-0 bg-transparent ps-2.5 focus:border-0 focus:ring-0"
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
