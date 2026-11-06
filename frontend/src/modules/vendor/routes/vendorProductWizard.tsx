"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { ArrowLeft, ArrowRight, Check, Loader2, PackagePlus } from "lucide-react";
import { Link, useRouter } from "@/shared/i18n/navigation";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorProduct } from "@/modules/vendor/types";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { Brand, Category, CategoryAttribute } from "@/modules/catalog/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

interface ProductForm {
	title: string;
	slug: string;
	brandId: string;
	categoryId: string;
	description: string;
	highlights: string;
	productType: "simple" | "variable";
	variationAxes: string;
	sku: string;
	price: string;
	crossesOutPrice: string;
	currency: string;
	initialStock: string;
}

const EMPTY_FORM: ProductForm = {
	title: "",
	slug: "",
	brandId: "",
	categoryId: "",
	description: "",
	highlights: "",
	productType: "simple",
	variationAxes: "",
	sku: "",
	price: "",
	crossesOutPrice: "",
	currency: "EGP",
	initialStock: "0",
};

type WizardStep = 0 | 1 | 2 | 3;

function slugify(value: string) {
	return value
		.trim()
		.toLowerCase()
		.normalize("NFKD")
		.replace(/[^\w\s-]/g, "")
		.replace(/[\s_-]+/g, "-")
		.replace(/^-+|-+$/g, "");
}

function parseMajorPrice(value: string): number | undefined {
	const parsed = Number(value);
	if (value.trim() === "" || !Number.isFinite(parsed) || parsed <= 0) return undefined;
	return Math.round(parsed * 100);
}

function draftKeyFor(productId?: string) {
	return productId ? `prim:vendor-product-draft:${productId}` : "prim:vendor-product-draft:new";
}

export function VendorProductWizard() {
	const t = useTranslations("vendor");
	const router = useRouter();
	const params = useParams();
	const productId = typeof params?.id === "string" ? params.id : undefined;
	const mode: "create" | "edit" = productId ? "edit" : "create";

	const [categories, setCategories] = useState<Category[]>([]);
	const [brands, setBrands] = useState<Brand[]>([]);
	const [categoryAttributes, setCategoryAttributes] = useState<CategoryAttribute[]>([]);
	const [attributesLoading, setAttributesLoading] = useState(false);
	const [attributesError, setAttributesError] = useState<string | null>(null);
	const [loadingCatalog, setLoadingCatalog] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const [formData, setFormData] = useState<ProductForm>(EMPTY_FORM);
	const [step, setStep] = useState<WizardStep>(0);
	const [errors, setErrors] = useState<Partial<Record<keyof ProductForm, string>>>({});
	const [saving, setSaving] = useState(false);
	const [publishing, setPublishing] = useState(false);
	const [hydrated, setHydrated] = useState(false);
	const [existingProduct, setExistingProduct] = useState<VendorProduct | null>(null);
	const [lastSavedAt, setLastSavedAt] = useState<number | null>(null);

	const steps = useMemo(
		() => [
			t("wizard.steps.identity"),
			t("wizard.steps.details"),
			t("wizard.steps.offer"),
			t("wizard.steps.review"),
		],
		[t],
	);

	const readiness = useMemo(() => {
		const checks = [
			{
				label: steps[0],
				done: Boolean(
					formData.title.trim() &&
						(mode === "edit" || formData.categoryId) &&
						(!formData.slug.trim() || /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(formData.slug.trim().toLowerCase())),
				),
			},
			{
				label: steps[1],
				done:
					formData.productType === "simple" ||
					formData.variationAxes.split(",").map((a) => a.trim()).filter(Boolean).length > 0,
			},
			{
				label: steps[2],
				done:
					mode === "edit" ||
					((!formData.price || parseMajorPrice(formData.price) !== undefined) &&
						(!formData.crossesOutPrice || parseMajorPrice(formData.crossesOutPrice) !== undefined) &&
						(!formData.price ||
							!formData.crossesOutPrice ||
							(parseMajorPrice(formData.crossesOutPrice) ?? 0) >= (parseMajorPrice(formData.price) ?? 0)) &&
						/^[A-Z]{3}$/.test(formData.currency.trim().toUpperCase()) &&
						Number.isInteger(Number.parseInt(formData.initialStock, 10)) &&
						Number.parseInt(formData.initialStock, 10) >= 0),
			},
			{ label: steps[3], done: false },
		];
		const doneCount = checks.filter((c) => c.done).length;
		return { checks, percent: Math.round((doneCount / checks.length) * 100) };
	}, [formData, steps, mode]);

	const draftKey = draftKeyFor(productId);

	// Hydrate from saved draft (create) or server product summary (edit).
	useEffect(() => {
		let cancelled = false;
		const hydrate = async () => {
			const [categoryList, brandList] = await Promise.all([
				catalogService.getCategories(100),
				catalogService.getBrands(100),
			]);
			if (cancelled) return;
			setCategories(Array.isArray(categoryList) ? categoryList : categoryList.data);
			setBrands(Array.isArray(brandList) ? brandList : brandList.data);

			const saved = sessionStorage.getItem(draftKey);
			if (saved) {
				try {
					setFormData({ ...EMPTY_FORM, ...(JSON.parse(saved) as ProductForm) });
					setHydrated(true);
					return;
				} catch {
					// Corrupt draft: fall through to server data.
				}
			}

			if (mode === "edit" && productId) {
				// The vendor list API is the summary source; fetch enough pages to find the product.
				let found: VendorProduct | undefined;
				for (let page = 1; page <= 10 && !found; page += 1) {
					const response = await vendorService.getProducts({ page, pageSize: 100 });
					found = response.data.find((item) => item.id === productId);
					if (response.data.length < 100) break;
				}
				if (!found) throw new Error(t("wizard.loadFailed"));
				setExistingProduct(found);
				setFormData({
					...EMPTY_FORM,
					title: found.title,
					slug: found.slug,
					productType: (found.productType as "simple" | "variable") || "simple",
					sku: found.sku ?? "",
					price: found.price ? String(found.price / 100) : "",
					currency: found.currency ?? "EGP",
				});
			}
		};

		hydrate()
			.catch((error) => {
				if (!cancelled) setLoadError(error instanceof Error ? error : new Error(String(error)));
			})
			.finally(() => {
				if (!cancelled) {
					setLoadingCatalog(false);
					setHydrated(true);
				}
			});
		return () => {
			cancelled = true;
		};
	}, [draftKey, mode, productId, t]);

	// Load variation-axis definitions for the selected category so the seller can
	// pick real axes instead of guessing keys. Categories with no attributes get a
	// clear empty state; request failures are surfaced, not silently ignored.
	useEffect(() => {
		let cancelled = false;
		if (!formData.categoryId) {
			setCategoryAttributes([]);
			setAttributesLoading(false);
			setAttributesError(null);
			return () => {
				cancelled = true;
			};
		}
		setAttributesLoading(true);
		setAttributesError(null);
		void catalogService
			.getCategoryAttributes(formData.categoryId)
			.then((response) => {
				if (cancelled) return;
				setCategoryAttributes(response.data ?? []);
			})
			.catch((error: unknown) => {
				if (cancelled) return;
				setCategoryAttributes([]);
				setAttributesError(apiErrorMessage(error, t("wizard.axes.loadFailed")));
			})
			.finally(() => {
				if (!cancelled) setAttributesLoading(false);
			});
		return () => {
			cancelled = true;
		};
	}, [formData.categoryId, t]);

	// Persist the staged draft so back/forward navigation and refresh keep state.
	useEffect(() => {
		if (!hydrated) return;
		sessionStorage.setItem(draftKey, JSON.stringify(formData));
		setLastSavedAt(Date.now());
	}, [draftKey, formData, hydrated]);

	const setField = <K extends keyof ProductForm>(key: K, value: ProductForm[K]) => {
		setFormData((current) => ({ ...current, [key]: value }));
		setErrors((current) => {
			if (!current[key]) return current;
			const next = { ...current };
			delete next[key];
			return next;
		});
	};

	const validateStep = (currentStep: WizardStep): boolean => {
		const nextErrors: Partial<Record<keyof ProductForm, string>> = {};
		if (currentStep === 0) {
			if (!formData.title.trim()) nextErrors.title = t("products.validation.titleRequired");
			else if (formData.title.trim().length > 200) nextErrors.title = t("products.validation.titleTooLong");
			if (formData.slug.trim() && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(formData.slug.trim().toLowerCase())) {
				nextErrors.slug = t("products.validation.slugInvalid");
			}
			if (mode === "create" && !formData.categoryId) {
				nextErrors.categoryId = t("products.validation.categoryRequired");
			}
		}
		if (currentStep === 1 && formData.productType === "variable") {
			if (formData.variationAxes.split(",").map((a) => a.trim()).filter(Boolean).length === 0) {
				nextErrors.variationAxes = t("products.validation.axesRequired");
			}
		}
		if (currentStep === 2) {
			if (formData.price && parseMajorPrice(formData.price) === undefined) {
				nextErrors.price = t("products.validation.priceInvalid");
			}
			if (formData.crossesOutPrice && parseMajorPrice(formData.crossesOutPrice) === undefined) {
				nextErrors.crossesOutPrice = t("products.validation.priceInvalid");
			}
			if (
				formData.price &&
				formData.crossesOutPrice &&
				(parseMajorPrice(formData.crossesOutPrice) ?? 0) < (parseMajorPrice(formData.price) ?? 0)
			) {
				nextErrors.crossesOutPrice = t("products.validation.compareAtTooLow");
			}
			if (!/^[A-Z]{3}$/.test(formData.currency.trim().toUpperCase())) {
				nextErrors.currency = t("products.validation.currencyInvalid");
			}
			const stock = Number.parseInt(formData.initialStock, 10);
			if (!Number.isInteger(stock) || stock < 0) nextErrors.initialStock = t("products.validation.stockInvalid");
		}
		setErrors(nextErrors);
		return Object.keys(nextErrors).length === 0;
	};

	const goNext = () => {
		if (validateStep(step) && step < 3) setStep((step + 1) as WizardStep);
	};

	const goBack = () => {
		if (step > 0) setStep((step - 1) as WizardStep);
	};

	const finish = async (shouldPublish: boolean) => {
		if (!validateStep(0)) {
			setStep(0);
			return;
		}
		if (!validateStep(1)) {
			setStep(1);
			return;
		}
		if (mode === "create" && !validateStep(2)) {
			setStep(2);
			return;
		}

		const priceMinor = formData.price ? parseMajorPrice(formData.price) : undefined;
		const crossedOutMinor = formData.crossesOutPrice ? parseMajorPrice(formData.crossesOutPrice) : undefined;
		setSaving(true);
		setPublishing(shouldPublish);
		try {
			if (mode === "edit" && productId) {
				await vendorService.updateProduct(productId, {
					title: formData.title.trim(),
					description: formData.description.trim() || undefined,
					highlights: formData.highlights.split("\n").map((h) => h.trim()).filter(Boolean),
					productType: formData.productType,
					variationAxes:
						formData.productType === "variable"
							? formData.variationAxes.split(",").map((axis) => axis.trim()).filter(Boolean)
							: [],
				});
				if (shouldPublish) await vendorService.publishProduct(productId);
				notify.success(shouldPublish ? t("toasts.productPublished") : t("toasts.productUpdated"));
			} else {
				const slug = formData.slug.trim() || slugify(formData.title);
				const highlightList = formData.highlights.split("\n").map((h) => h.trim()).filter(Boolean);
				const created = await vendorService.createProduct({
					brandId: formData.brandId || undefined,
					categoryId: formData.categoryId,
					slug,
					title: formData.title.trim(),
					description: formData.description.trim() || undefined,
					highlights: highlightList.length > 0 ? highlightList : undefined,
					productType: formData.productType,
					variationAxes:
						formData.productType === "variable"
							? formData.variationAxes.split(",").map((axis) => axis.trim()).filter(Boolean)
							: undefined,
				});
				const variant = await vendorService.createVariant(created.data.id, {
					title: formData.title.trim(),
					sku: formData.sku.trim() || undefined,
					price: priceMinor,
					crossedOutPrice: crossedOutMinor,
					currency: formData.currency.trim() || "EGP",
					isDefault: true,
				});
				const initialStock = Number.parseInt(formData.initialStock, 10);
				if (Number.isFinite(initialStock) && initialStock !== 0) {
					await vendorService.adjustStock(variant.data.id, { quantity: initialStock, reason: "restock" });
				}
				if (shouldPublish) await vendorService.publishProduct(created.data.id);
				notify.success(shouldPublish ? t("toasts.productPublished") : t("toasts.productAdded"));
			}
			sessionStorage.removeItem(draftKey);
			router.replace("/vendor/products");
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setSaving(false);
			setPublishing(false);
		}
	};

	const fieldError = (key: keyof ProductForm) =>
		errors[key] ? (
			<p className="text-xs text-destructive" role="alert">
				{errors[key]}
			</p>
		) : null;

	if (loadingCatalog || !hydrated) {
		return (
			<div className="flex min-h-[50vh] items-center justify-center">
				<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
			</div>
		);
	}

	if (loadError) {
		return (
			<div className="space-y-5">
				<PageHeader title={mode === "edit" ? t("products.editTitle") : t("products.addTitle")} />
				<ErrorState
					title={t("products.loadTitle")}
					description={loadError.message || t("wizard.loadFailed")}
					onRetry={() => router.refresh()}
				/>
			</div>
		);
	}

	const reviewRows = [
		{ label: t("products.form.title"), value: formData.title.trim() || "—" },
		{ label: t("products.form.highlights"), value: formData.highlights.split("\n").map((h) => h.trim()).filter(Boolean).join(" · ") || "—" },
		{ label: t("products.form.slug"), value: formData.slug.trim() || slugify(formData.title) || "—" },
		{ label: t("products.form.category"), value: categories.find((c) => c.id === formData.categoryId)?.name ?? (mode === "edit" ? t("wizard.review.categoryUnchanged") : "—") },
		{ label: t("products.form.brand"), value: brands.find((b) => b.id === formData.brandId)?.name ?? t("products.form.noBrand") },
		{ label: t("products.form.productType"), value: formData.productType === "variable" ? t("products.form.variable") : t("products.form.simple") },
		{ label: t("products.form.variationAxes"), value: formData.variationAxes.trim() || "—" },
		{ label: t("products.form.skuCode"), value: formData.sku.trim() || "—" },
		{ label: t("products.form.sellingPrice"), value: formData.price ? `${formData.currency.toUpperCase()} ${(Number(formData.price)).toFixed(2)}` : "—" },
		{ label: t("products.form.compareAtPrice"), value: formData.crossesOutPrice ? `${formData.currency.toUpperCase()} ${Number(formData.crossesOutPrice).toFixed(2)}` : "—" },
		{ label: t("products.form.initialStock"), value: formData.initialStock || "—" },
	];

	return (
		<div className="mx-auto w-full max-w-3xl space-y-5">
			<PageHeader
				title={mode === "edit" ? t("products.editTitle") : t("products.addTitle")}
				description={t("wizard.subtitle")}
				actions={
					<Link
						href="/vendor/products"
						className="inline-flex items-center gap-1.5 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
					>
						<ArrowLeft className="size-4 rtl:rotate-180" />
						{t("actions.cancel")}
					</Link>
				}
			/>

			<ol className="grid grid-cols-2 gap-2 sm:grid-cols-4" aria-label={t("wizard.progressLabel")}>
				{steps.map((label, index) => {
					const isCurrent = index === step;
					const isDone = index < step;
					return (
						<li
							key={label}
							className={`rounded-lg border px-3 py-2.5 text-xs ${
								isCurrent
									? "border-primary bg-primary/5 font-semibold text-foreground"
									: "border-border bg-background text-muted-foreground"
							}`}
							aria-current={isCurrent ? "step" : undefined}
						>
							<span className="flex items-center gap-1.5">
								{isDone ? <Check className="size-3.5 text-emerald-600" /> : <span className="font-bold">{index + 1}</span>}
								{label}
							</span>
						</li>
					);
				})}
			</ol>


			<section className="rounded-lg border border-border bg-background p-4">
				<div className="flex items-center justify-between">
					<h2 className="text-sm font-semibold text-foreground">{t("wizard.readinessTitle")}</h2>
					<span className="text-xs text-muted-foreground">
						{t("wizard.readinessPercent", { percent: readiness.percent })}
						{lastSavedAt ? ` · ${t("wizard.autosavedAt", { time: new Date(lastSavedAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) })}` : ""}
					</span>
				</div>
				<div className="mt-3 h-1.5 overflow-hidden rounded-full bg-secondary" role="progressbar" aria-valuenow={readiness.percent} aria-valuemax={100}>
					<div className="h-full bg-primary transition-all" style={{ width: `${readiness.percent}%` }} />
				</div>
				<ul className="mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4">
					{readiness.checks.map((check) => (
						<li key={check.label} className={`rounded-md border px-2.5 py-2 ${check.done ? "border-emerald-600/30 bg-emerald-600/5 text-foreground" : "border-border text-muted-foreground"}`}>
							{check.done ? "✓ " : ""}{check.label}
						</li>
					))}
				</ul>
			</section>


			<section className="rounded-lg border border-border bg-background p-5">
				<h2 className="text-base font-semibold text-foreground">{steps[step]}</h2>
				<p className="mt-1 text-xs text-muted-foreground">{t("wizard.stepHint", { step: step + 1, total: steps.length })}</p>

				{step === 0 && (
					<div className="mt-5 space-y-4">
						<div className="space-y-1">
							<label htmlFor="wizard-title" className="text-sm font-medium text-foreground">{t("products.form.title")}</label>
							<input
								id="wizard-title"
								required
								value={formData.title}
								onChange={(e) => setField("title", e.target.value)}
								className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
							/>
							{fieldError("title")}
						</div>
						{mode === "create" && (
							<div className="space-y-1">
								<label htmlFor="wizard-slug" className="text-sm font-medium text-foreground">{t("products.form.slug")}</label>
								<input
									id="wizard-slug"
									dir="ltr"
									value={formData.slug}
									onChange={(e) => setField("slug", e.target.value)}
									placeholder={formData.title ? slugify(formData.title) : t("products.form.slugPlaceholder")}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
								{fieldError("slug")}
							</div>
						)}
						{mode === "create" && (
							<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
								<div className="space-y-1">
									<label htmlFor="wizard-category" className="text-sm font-medium text-foreground">{t("products.form.category")}</label>
									<select
										id="wizard-category"
										required
										value={formData.categoryId}
										onChange={(e) => setField("categoryId", e.target.value)}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									>
										<option value="">{t("products.form.chooseCategory")}</option>
										{categories.map((c) => (
											<option key={c.id} value={c.id}>{c.name}</option>
										))}
									</select>
									{fieldError("categoryId")}
								</div>
								<div className="space-y-1">
									<label htmlFor="wizard-brand" className="text-sm font-medium text-foreground">{t("products.form.brand")}</label>
									<select
										id="wizard-brand"
										value={formData.brandId}
										onChange={(e) => setField("brandId", e.target.value)}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									>
										<option value="">{t("products.form.noBrand")}</option>
										{brands.map((b) => (
											<option key={b.id} value={b.id}>{b.name}</option>
										))}
									</select>
								</div>
							</div>
						)}
					</div>
				)}

				{step === 1 && (
					<div className="mt-5 space-y-4">
						<div className="space-y-1">
							<label htmlFor="wizard-description" className="text-sm font-medium text-foreground">{t("products.form.description")}</label>
							<textarea
								id="wizard-description"
								value={formData.description}
								onChange={(e) => setField("description", e.target.value)}
								rows={4}
								className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
							/>
						</div>
						<div className="space-y-1">
							<label htmlFor="wizard-highlights" className="text-sm font-medium text-foreground">{t("products.form.highlights")}</label>
							<textarea
								id="wizard-highlights"
								value={formData.highlights}
								onChange={(e) => setField("highlights", e.target.value)}
								rows={3}
								placeholder={t("products.form.highlightsPlaceholder")}
								className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
							/>
							<p className="text-xs text-muted-foreground">{t("products.form.highlightsNote")}</p>
						</div>
						<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
							<div className="space-y-1">
								<label htmlFor="wizard-type" className="text-sm font-medium text-foreground">{t("products.form.productType")}</label>
								<select
									id="wizard-type"
									value={formData.productType}
									onChange={(e) => setField("productType", e.target.value as "simple" | "variable")}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								>
									<option value="simple">{t("products.form.simple")}</option>
									<option value="variable">{t("products.form.variable")}</option>
								</select>
							</div>
							{formData.productType === "variable" && !categoryAttributes.length && (
								<div className="space-y-1 rounded-md border border-amber-600/25 bg-amber-600/5 p-3 text-xs leading-5 text-muted-foreground">
									<p className="font-medium text-foreground">{t("wizard.axes.noneTitle")}</p>
									<p>
										{attributesLoading
											? t("wizard.axes.loading")
											: attributesError
												? t("wizard.axes.loadFailed")
												: t("wizard.axes.noneDescription")}
									</p>
								</div>
							)}
							{formData.productType === "variable" && categoryAttributes.length > 0 && (
								<div className="space-y-2 rounded-md border border-border p-3">
									<fieldset>
										<legend className="text-xs font-medium text-muted-foreground">
											{t("wizard.axes.selectLabel")}
										</legend>
										<div className="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-2">
											{categoryAttributes.map((attribute) => {
												const selected = formData.variationAxes
													.split(",")
													.map((axis) => axis.trim())
													.filter(Boolean)
													.includes(attribute.key);
												return (
													<label
														key={attribute.key}
														className="flex cursor-pointer items-start gap-2 rounded-md border border-border px-3 py-2 text-sm text-foreground"
													>
														<input
															type="checkbox"
															checked={selected}
															onChange={(e) => {
																const axes = formData.variationAxes
																	.split(",")
																	.map((v) => v.trim())
																	.filter(Boolean);
																const next = e.target.checked
																	? [...axes, attribute.key]
																	: axes.filter((v) => v !== attribute.key);
																setField("variationAxes", next.join(", "));
															}}
														/>
														<span>
															<span className="font-medium">{attribute.label}</span>
															{attribute.allowedValues?.length ? (
																<span className="block text-xs text-muted-foreground">
																	{attribute.allowedValues.slice(0, 4).join(" · ")}
																</span>
															) : null}
														</span>
													</label>
												);
											})}
										</div>
									</fieldset>
								</div>
							)}
							{formData.productType === "variable" && (
								<div className="space-y-1">
									<label htmlFor="wizard-axes" className="text-sm font-medium text-foreground">{t("products.form.variationAxes")}</label>
									<input
										id="wizard-axes"
										value={formData.variationAxes}
										onChange={(e) => setField("variationAxes", e.target.value)}
										placeholder={t("products.form.variationAxesPlaceholder")}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									/>
									{fieldError("variationAxes")}
								</div>
							)}
						</div>
					</div>
				)}

				{step === 2 && (
					<div className="mt-5 space-y-4">
						<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
							<div className="space-y-1">
								<label htmlFor="wizard-sku" className="text-sm font-medium text-foreground">{t("products.form.skuCode")}</label>
								<input
									id="wizard-sku"
									value={formData.sku}
									onChange={(e) => setField("sku", e.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
							</div>
							<div className="space-y-1">
								<label htmlFor="wizard-price" className="text-sm font-medium text-foreground">{t("products.form.sellingPrice")}</label>
								<input
									id="wizard-price"
									type="number"
									min="0.01"
									step="0.01"
									value={formData.price}
									onChange={(e) => setField("price", e.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
								{fieldError("price")}
							</div>
						</div>
						<div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
							<div className="space-y-1">
								<label htmlFor="wizard-compare" className="text-sm font-medium text-foreground">{t("products.form.compareAtPrice")}</label>
								<input
									id="wizard-compare"
									type="number"
									min="0"
									step="0.01"
									value={formData.crossesOutPrice}
									onChange={(e) => setField("crossesOutPrice", e.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
								{fieldError("crossesOutPrice")}
							</div>
							<div className="space-y-1">
								<label htmlFor="wizard-currency" className="text-sm font-medium text-foreground">{t("products.form.currency")}</label>
								<input
									id="wizard-currency"
									value={formData.currency}
									onChange={(e) => setField("currency", e.target.value.toUpperCase())}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
								{fieldError("currency")}
							</div>
							<div className="space-y-1">
								<label htmlFor="wizard-stock" className="text-sm font-medium text-foreground">{t("products.form.initialStock")}</label>
								<input
									id="wizard-stock"
									type="number"
									min="0"
									value={formData.initialStock}
									onChange={(e) => setField("initialStock", e.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
								{fieldError("initialStock")}
							</div>
						</div>
						<p className="text-xs text-muted-foreground">{t("products.form.initialPurchaseNote")}</p>
					</div>
				)}

				{step === 3 && (
					<div className="mt-5 space-y-4">
						<dl className="divide-y divide-border rounded-md border border-border text-sm">
							{reviewRows.map((row) => (
								<div key={row.label} className="flex items-center justify-between gap-4 px-4 py-2.5">
									<dt className="text-muted-foreground">{row.label}</dt>
									<dd className="truncate text-end font-medium text-foreground">{row.value}</dd>
								</div>
							))}
						</dl>
						<p className="text-xs text-muted-foreground">{t("wizard.review.note")}</p>
						{existingProduct?.status === "published" && mode === "edit" ? (
							<p className="rounded-md border border-border bg-secondary/40 px-3 py-2 text-xs text-muted-foreground">
								{t("wizard.review.publishedNote")}
							</p>
						) : null}
					</div>
				)}

				<div className="mt-6 flex flex-col-reverse gap-2 border-t border-border pt-4 sm:flex-row sm:items-center sm:justify-between">
					<button
						type="button"
						onClick={goBack}
						disabled={step === 0 || saving}
						className="inline-flex items-center justify-center gap-1.5 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground disabled:opacity-50"
					>
						<ArrowLeft className="size-4 rtl:rotate-180" />
						{t("wizard.back")}
					</button>
					<div className="flex flex-wrap items-center justify-end gap-2">
						{step < 3 ? (
							<button
								type="button"
								onClick={goNext}
								className="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
							>
								{t("wizard.continue")}
								<ArrowRight className="size-4 rtl:rotate-180" />
							</button>
						) : (
							<>
								<button
									type="button"
									disabled={saving}
									onClick={() => void finish(false)}
									className="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground disabled:opacity-60"
								>
									{saving && !publishing ? <Loader2 className="size-4 animate-spin" /> : null}
									{t("wizard.saveDraft")}
								</button>
								<button
									type="button"
									disabled={saving}
									onClick={() => void finish(true)}
									className="inline-flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
								>
									{saving && publishing ? <Loader2 className="size-4 animate-spin" /> : <PackagePlus className="size-4" />}
									{t("wizard.savePublish")}
								</button>
							</>
						)}
					</div>
				</div>
			</section>
		</div>
	);
}
