"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { ArrowLeft, ArrowRight, Check, Loader2, PackagePlus, Plus, Trash2 } from "lucide-react";
import { Link, useRouter } from "@/shared/i18n/navigation";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerProduct, SellerCustomAttribute } from "@/modules/seller/types";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { Brand, Category, CategoryAttribute } from "@/modules/catalog/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ProductGallery } from "@/modules/catalog/components/ui/ProductGallery";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/form/select";
import { colorHexFor, isColorAttribute } from "@/modules/seller/lib/colorSwatch";
import type {
	ProductForm,
	VariationRow,
	VariantOverride,
	WizardStep,
} from "./wizard/wizardTypes";
import {
	EMPTY_FORM,
	bundlePayloadItems,
	slugify,
	parseMajorPrice,
	draftKeyFor,
	uniqueKeys,
	formatCombination,
	combinationKey,
} from "./wizard/wizardTypes";
import { usePersistDraft } from "./wizard/useWizardDraft";
import { CustomAttributeBuilder } from "./wizard/CustomAttributeBuilder";
import { VariantPricing } from "./wizard/VariantPricing";
import { BundleItems } from "./wizard/BundleItems";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

// Downloadable products deliver a hosted file, so the seller link must be a
// real http(s) URL rather than free text.
function isValidDownloadUrl(value: string) {
	return /^https?:\/\/\S+$/i.test(value.trim());
}

export function SellerProductWizard() {
	const t = useTranslations("seller");
	const router = useRouter();
	const params = useParams();
	const productId = typeof params?.id === "string" ? params.id : undefined;
	const mode: "create" | "edit" = productId ? "edit" : "create";

	const [categories, setCategories] = useState<Category[]>([]);
	const [brands, setBrands] = useState<Brand[]>([]);
	const [categoryAttributes, setCategoryAttributes] = useState<CategoryAttribute[]>([]);
	const [attributesLoading, setAttributesLoading] = useState(false);
	const [attributesError, setAttributesError] = useState<string | null>(null);
	const [variationRows, setVariationRows] = useState<VariationRow[]>([]);
	const [savedCombinations, setSavedCombinations] = useState<Record<string, string>[]>([]);
	const [swatchImages, setSwatchImages] = useState<Record<string, string>>({});
	const [variantOverrides, setVariantOverrides] = useState<Record<string, VariantOverride>>({});
	const [uploadingThumbnailKey, setUploadingThumbnailKey] = useState<string | null>(null);
	const [previewValues, setPreviewValues] = useState<Record<string, string>>({});
	const [previewImageIndex, setPreviewImageIndex] = useState(0);
	const [loadingCatalog, setLoadingCatalog] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const [formData, setFormData] = useState<ProductForm>(EMPTY_FORM);
	const [step, setStep] = useState<WizardStep>(0);
	const [errors, setErrors] = useState<Partial<Record<keyof ProductForm, string>>>({});
	const [saving, setSaving] = useState(false);
	const [publishing, setPublishing] = useState(false);
	const [hydrated, setHydrated] = useState(false);
	const [existingProduct, setExistingProduct] = useState<SellerProduct | null>(null);
	const [lastSavedAt, setLastSavedAt] = useState<number | null>(null);
	const [bundleCandidates, setBundleCandidates] = useState<SellerProduct[]>([]);
	const [bundleCandidatesState, setBundleCandidatesState] = useState<"idle" | "loading" | "ready">("idle");

	const steps = useMemo(
		() => [
			t("wizard.steps.identity"),
			t("wizard.steps.details"),
			t("wizard.steps.offer"),
			t("wizard.steps.review"),
		],
		[t],
	);

	const validBundleItems = useMemo(() => bundlePayloadItems(formData.bundleItems), [formData.bundleItems]);

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
					formData.productType === "bundle"
						? validBundleItems.length > 0
						: formData.productType === "downloadable"
							? isValidDownloadUrl(formData.downloadUrl)
							: formData.productType === "simple" ||
								formData.variationAxes.split(",").map((a) => a.trim()).filter(Boolean).length > 0,
			},
			{
				label: steps[2],
				done:
					mode === "edit" ||
					(formData.productType === "bundle" || formData.productType === "downloadable"
						? parseMajorPrice(formData.price) !== undefined &&
							/^[A-Z]{3}$/.test(formData.currency.trim().toUpperCase())
						: formData.productType === "variable"
							? savedCombinations.length > 0 &&
								savedCombinations.every((combination) => {
									const price = variantOverrides[combinationKey(combination)]?.price;
									return price !== undefined && parseMajorPrice(price) !== undefined;
								}) &&
								/^[A-Z]{3}$/.test(formData.currency.trim().toUpperCase())
							: (!formData.price || parseMajorPrice(formData.price) !== undefined) &&
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
	}, [formData, steps, mode, savedCombinations, variantOverrides, validBundleItems]);

	const combinations = useMemo(
		() => (formData.productType === "variable" ? savedCombinations : []),
		[formData.productType, savedCombinations],
	);

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
					const parsed = JSON.parse(saved) as {
						formData?: ProductForm;
						variantOverrides?: Record<string, VariantOverride>;
						savedCombinations?: Record<string, string>[];
						swatchImages?: Record<string, string>;
					} & ProductForm;
					// Older drafts stored the form directly; newer drafts wrap it
					// so per-variant overrides, saved combinations, and swatch
					// images survive a refresh.
					const draftForm = parsed.formData ?? parsed;
					setFormData({ ...EMPTY_FORM, ...draftForm });
					setVariationRows([]);
					setSavedCombinations(parsed.formData ? (parsed.savedCombinations ?? []) : []);
					setSwatchImages(parsed.formData ? (parsed.swatchImages ?? {}) : {});
					setVariantOverrides(parsed.formData ? (parsed.variantOverrides ?? {}) : {});
					setHydrated(true);
					return;
				} catch {
					// Corrupt draft: fall through to server data.
				}
			}

			if (mode === "edit" && productId) {
				// The seller list API is the summary source; fetch enough pages to find the product.
				let found: SellerProduct | undefined;
				for (let page = 1; page <= 10 && !found; page += 1) {
					const response = await sellerService.getProducts({ page, pageSize: 100 });
					found = response.data.find((item) => item.id === productId);
					if (response.data.length < 100) break;
				}
				if (!found) throw new Error(t("wizard.loadFailed"));
				setExistingProduct(found);
				const initialAxes = found.productType === "variable" ? "" : "";
				setFormData({
					...EMPTY_FORM,
					title: found.title,
					slug: found.slug,
					productType: (found.productType as ProductForm["productType"]) || "simple",
					variationAxes: initialAxes,
					sku: found.sku ?? "",
					price: found.price ? String(found.price / 100) : "",
					currency: found.currency ?? "EGP",
				});
				setVariationRows([]);
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
		void Promise.all([
			catalogService.getCategoryAttributes(formData.categoryId),
			sellerService.getAttributes(formData.categoryId).catch(() => ({ data: [] as SellerCustomAttribute[] })),
		])
			.then(([response, sellerResponse]) => {
				if (cancelled) return;
				const sellerAttributes = (sellerResponse.data ?? []).map((attribute) => ({
					key: attribute.key,
					label: attribute.label,
					valueType: attribute.valueType || "enum",
					allowedValues: Array.isArray(attribute.allowedValues) ? attribute.allowedValues : undefined,
					isFilterable: attribute.isFilterable,
					sortOrder: attribute.sortOrder ?? 0,
				}));
				setCategoryAttributes([...(response.data ?? []), ...sellerAttributes]);
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

	// Bundle composition picks from the seller's own catalog, loaded on first
	// use so non-bundle products never pay for the extra request.
	useEffect(() => {
		if (formData.productType !== "bundle" || bundleCandidatesState !== "idle") return;
		setBundleCandidatesState("loading");
		sellerService
			.getProducts({ page: 1, pageSize: 100 })
			.then((response) => setBundleCandidates(response.data ?? []))
			.catch((error: unknown) => notify.error(apiErrorMessage(error, t("wizard.bundle.itemsLoadFailed"))))
			.finally(() => setBundleCandidatesState("ready"));
	}, [formData.productType, bundleCandidatesState, t]);

	usePersistDraft(
		draftKey,
		{ formData, variantOverrides, savedCombinations, swatchImages },
		hydrated,
		setLastSavedAt,
	);

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
			if (savedCombinations.length === 0) {
				nextErrors.variationAxes = t("products.validation.axesRequired");
			} else if (variationRows.some((row) => row.key && !row.value)) {
				nextErrors.variationAxes = t("wizard.attributes.rowsIncomplete");
			}
		}
		if (currentStep === 1 && formData.productType === "bundle" && validBundleItems.length === 0) {
			nextErrors.bundleItems = t("wizard.bundle.itemsRequired");
		}
		if (currentStep === 1 && formData.productType === "downloadable" && !isValidDownloadUrl(formData.downloadUrl)) {
			nextErrors.downloadUrl = formData.downloadUrl.trim()
				? t("products.validation.downloadUrlInvalid")
				: t("products.validation.downloadUrlRequired");
		}
		if (currentStep === 2) {
			const isVariableCreate = mode === "create" && formData.productType === "variable";
			const isDigitalCreate = mode === "create" && (formData.productType === "bundle" || formData.productType === "downloadable");
			if (isDigitalCreate && parseMajorPrice(formData.price) === undefined) {
				nextErrors.price = t(
					formData.productType === "bundle"
						? "products.validation.bundlePriceRequired"
						: "products.validation.priceInvalid",
				);
			}
			if (!isVariableCreate && !isDigitalCreate) {
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
				const stock = Number.parseInt(formData.initialStock, 10);
				if (!Number.isInteger(stock) || stock < 0) nextErrors.initialStock = t("products.validation.stockInvalid");
			}
			if (!/^[A-Z]{3}$/.test(formData.currency.trim().toUpperCase())) {
				nextErrors.currency = t("products.validation.currencyInvalid");
			}
			if (isVariableCreate) {
				for (const combination of savedCombinations) {
					const override = variantOverrides[combinationKey(combination)];
					if (!override?.price) {
						nextErrors.price = t("products.validation.priceInvalid");
					} else if (parseMajorPrice(override.price) === undefined) {
						nextErrors.price = t("products.validation.priceInvalid");
					}
					if (override?.crossesOutPrice && parseMajorPrice(override.crossesOutPrice) === undefined) {
						nextErrors.crossesOutPrice = t("products.validation.priceInvalid");
					}
					const overridePrice = override?.price ? parseMajorPrice(override.price) : undefined;
					const overrideCompare = override?.crossesOutPrice ? parseMajorPrice(override.crossesOutPrice) : undefined;
					if (overridePrice !== undefined && overrideCompare !== undefined && overrideCompare < overridePrice) {
						nextErrors.crossesOutPrice = t("products.validation.compareAtTooLow");
					}
					if (override?.stock !== undefined && override.stock !== "") {
						const overrideStock = Number.parseInt(override.stock, 10);
						if (!Number.isInteger(overrideStock) || overrideStock < 0) {
							nextErrors.initialStock = t("products.validation.stockInvalid");
						}
					}
				}
			}
		}
		setErrors(nextErrors);
		return Object.keys(nextErrors).length === 0;
	};

	const goNext = () => {
		if (validateStep(step) && step < 3) setStep((step + 1) as WizardStep);
	};

	const syncSavedAxes = (next: Record<string, string>[]) => {
		setSavedCombinations(next);
		setField("variationAxes", uniqueKeys(next).join(", "));
	};

	const saveCombination = () => {
		const combination: Record<string, string> = {};
		for (const row of variationRows) {
			if (!row.key || !row.value) return;
			combination[row.key] = row.value;
		}
		if (Object.keys(combination).length === 0) return;
		const key = combinationKey(combination);
		if (savedCombinations.some((saved) => combinationKey(saved) === key)) return;
		syncSavedAxes([...savedCombinations, combination]);
		setVariationRows([]);
	};

	const removeSavedCombination = (index: number) => {
		syncSavedAxes(savedCombinations.filter((_, i) => i !== index));
	};

	const goBack = () => {
		if (step > 0) setStep((step - 1) as WizardStep);
	};

	const uploadVariantThumbnail = async (key: string, file: File | undefined) => {
		if (!file) return;
		setUploadingThumbnailKey(key);
		try {
			const uploaded = await sellerService.uploadMedia(file);
			setVariantOverrides((current) => ({
				...current,
				[key]: { ...(current[key] ?? {}), thumbnailUrl: uploaded.data.url },
			}));
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setUploadingThumbnailKey(null);
		}
	};

	const uploadSwatchImage = async (swatchKey: string, file: File | undefined) => {
		if (!file) return;
		setUploadingThumbnailKey(swatchKey);
		try {
			const uploaded = await sellerService.uploadMedia(file);
			setSwatchImages((current) => ({ ...current, [swatchKey]: uploaded.data.url }));
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setUploadingThumbnailKey(null);
		}
	};

	const finish = async (shouldPublish: boolean) => {
		if (formData.productType === "variable") {
			setField("variationAxes", uniqueKeys(savedCombinations).join(", "));
		}
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

		if (mode === "create" && formData.productType === "bundle") {
			const bundlePrice = parseMajorPrice(formData.price);
			if (bundlePrice === undefined) {
				setStep(2);
				return;
			}
			setSaving(true);
			setPublishing(shouldPublish);
			try {
				const created = await sellerService.createBundle({
					title: formData.title.trim(),
					description: formData.description.trim() || undefined,
					bundlePrice,
					currency: formData.currency.trim().toUpperCase() || "EGP",
					items: validBundleItems,
				});
				if (shouldPublish) await sellerService.publishProduct(created.data.productId);
				notify.success(shouldPublish ? t("toasts.productPublished") : t("toasts.productAdded"));
				sessionStorage.removeItem(draftKey);
				router.replace("/seller/products");
			} catch (error) {
				notify.error(apiErrorMessage(error, t("products.saveFailed")));
			} finally {
				setSaving(false);
				setPublishing(false);
			}
			return;
		}

		const priceMinor = formData.price ? parseMajorPrice(formData.price) : undefined;
		const crossedOutMinor = formData.crossesOutPrice ? parseMajorPrice(formData.crossesOutPrice) : undefined;
		setSaving(true);
		setPublishing(shouldPublish);
		try {
			if (mode === "edit" && productId) {
				await sellerService.updateProduct(productId, {
					title: formData.title.trim(),
					description: formData.description.trim() || undefined,
					highlights: formData.highlights.split("\n").map((h) => h.trim()).filter(Boolean),
					productType: formData.productType,
					variationAxes:
						formData.productType === "variable"
							? formData.variationAxes.split(",").map((axis) => axis.trim()).filter(Boolean)
							: [],
				});
				if (shouldPublish) await sellerService.publishProduct(productId);
				notify.success(shouldPublish ? t("toasts.productPublished") : t("toasts.productUpdated"));
			} else {
				const slug = formData.slug.trim() || slugify(formData.title);
				const highlightList = formData.highlights.split("\n").map((h) => h.trim()).filter(Boolean);
				const axes = formData.productType === "variable" ? uniqueKeys(savedCombinations) : undefined;
				const created = await sellerService.createProduct({
					brandId: formData.brandId || undefined,
					categoryId: formData.categoryId,
					slug,
					title: formData.title.trim(),
					description: formData.description.trim() || undefined,
					highlights: highlightList.length > 0 ? highlightList : undefined,
					productType: formData.productType,
					variationAxes: axes,
					downloadUrl: formData.productType === "downloadable" ? formData.downloadUrl.trim() || undefined : undefined,
					deliveryInstructions:
						formData.productType === "downloadable" ? formData.deliveryInstructions.trim() || undefined : undefined,
				});
				const initialStock =
					formData.productType === "downloadable" ? 0 : Number.parseInt(formData.initialStock, 10);
				if (combinations.length > 0) {
					for (let index = 0; index < combinations.length; index += 1) {
						const combination = combinations[index];
						const override = variantOverrides[combinationKey(combination)];
						const label = formatCombination(combination);
						const variantPrice = override?.price ? parseMajorPrice(override.price) : undefined;
						const variantCrossedOut = override?.crossesOutPrice ? parseMajorPrice(override.crossesOutPrice) : undefined;
						const variantStock =
							override?.stock !== undefined && override.stock !== "" ? Number.parseInt(override.stock, 10) : undefined;
						const swatchThumbnail = Object.entries(combination)
							.map(([axisKey, axisValue]) => swatchImages[`swatch:${axisKey}=${axisValue}`])
							.find((url): url is string => Boolean(url));
						const variant = await sellerService.createVariant(created.data.id, {
							title: `${formData.title.trim()} - ${label}`,
							sku: override?.sku?.trim() || undefined,
							price: variantPrice,
							crossedOutPrice: variantCrossedOut,
							currency: formData.currency.trim() || "EGP",
							attributes: combination,
							isDefault: index === 0,
							thumbnailUrl: override?.thumbnailUrl ?? swatchThumbnail,
						});
						if (variantStock !== undefined && Number.isFinite(variantStock) && variantStock !== 0) {
							await sellerService.adjustStock(variant.data.id, { quantity: variantStock, reason: "restock" });
						}
					}
				} else {
					const variant = await sellerService.createVariant(created.data.id, {
						title: formData.title.trim(),
						sku: formData.sku.trim() || undefined,
						price: priceMinor,
						crossedOutPrice: crossedOutMinor,
						currency: formData.currency.trim() || "EGP",
						isDefault: true,
					});
					if (Number.isFinite(initialStock) && initialStock !== 0) {
						await sellerService.adjustStock(variant.data.id, { quantity: initialStock, reason: "restock" });
					}
				}
				if (shouldPublish) await sellerService.publishProduct(created.data.id);
				notify.success(shouldPublish ? t("toasts.productPublished") : t("toasts.productAdded"));
			}
			sessionStorage.removeItem(draftKey);
			router.replace("/seller/products");
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

	const isBundle = formData.productType === "bundle";
	const isDownloadable = formData.productType === "downloadable";
	const hidesInventory = isBundle || isDownloadable;
	const productTypeLabel = t(
		isBundle
			? "products.form.bundle"
			: isDownloadable
				? "products.form.downloadable"
				: formData.productType === "variable"
					? "products.form.variable"
					: "products.form.simple",
	);

	const previewCombinations = mode === "create" && formData.productType === "variable" ? savedCombinations : [];
	const previewAxes = uniqueKeys(previewCombinations);
	const previewCombination =
		previewCombinations.find((combination) =>
			previewAxes.every((axis) => {
				const chosen = previewValues[axis];
				return chosen === undefined || combination[axis] === chosen;
			}),
		) ?? previewCombinations[0];
	const previewOverride = previewCombination ? variantOverrides[combinationKey(previewCombination)] : undefined;
	const previewImages = previewCombinations
		.map(
			(combination) =>
				variantOverrides[combinationKey(combination)]?.thumbnailUrl ??
				Object.entries(combination)
					.map(([axisKey, axisValue]) => swatchImages[`swatch:${axisKey}=${axisValue}`])
					.find((url): url is string => Boolean(url)),
		)
		.filter((url): url is string => Boolean(url));
	const previewPrice = formData.productType === "variable" ? previewOverride?.price : formData.price;
	const previewCompareAt = formData.productType === "variable" ? previewOverride?.crossesOutPrice : formData.crossesOutPrice;
	const previewSku = formData.productType === "variable" ? previewOverride?.sku : formData.sku;
	const previewStock = formData.productType === "variable" ? previewOverride?.stock : formData.initialStock;
	const previewDiscount =
		previewPrice && previewCompareAt && Number(previewCompareAt) > Number(previewPrice)
			? `${Math.round((1 - Number(previewPrice) / Number(previewCompareAt)) * 100)}%`
			: undefined;

	const reviewRows = [
		{ label: t("products.form.title"), value: formData.title.trim() || "—" },
		{ label: t("products.form.highlights"), value: formData.highlights.split("\n").map((h) => h.trim()).filter(Boolean).join(" · ") || "—" },
		{ label: t("products.form.slug"), value: formData.slug.trim() || slugify(formData.title) || "—" },
		{ label: t("products.form.category"), value: categories.find((c) => c.id === formData.categoryId)?.name ?? (mode === "edit" ? t("wizard.review.categoryUnchanged") : "—") },
		{ label: t("products.form.brand"), value: brands.find((b) => b.id === formData.brandId)?.name ?? t("products.form.noBrand") },
		{ label: t("products.form.productType"), value: productTypeLabel },
		{ label: t("products.form.variationAxes"), value: savedCombinations.length > 0 ? uniqueKeys(savedCombinations).join(", ") : (formData.variationAxes.trim() || "—") },
		{ label: t("products.form.skuCode"), value: formData.sku.trim() || "—" },
		{ label: t(isBundle ? "bundles.bundlePrice" : "products.form.sellingPrice"), value: formData.price ? `${formData.currency.toUpperCase()} ${(Number(formData.price)).toFixed(2)}` : "—" },
		{ label: t("products.form.compareAtPrice"), value: formData.crossesOutPrice ? `${formData.currency.toUpperCase()} ${Number(formData.crossesOutPrice).toFixed(2)}` : "—" },
		{ label: t("products.form.initialStock"), value: formData.initialStock || "—" },
		...(isDownloadable && mode === "create"
			? [{ label: t("products.form.downloadUrl"), value: formData.downloadUrl.trim() || "—" }]
			: []),
		...(formData.productType === "bundle" && mode === "create"
			? [
					{
						label: t("bundles.itemsTitle"),
						value:
							validBundleItems
								.map((item) => {
									const product = bundleCandidates.find((candidate) => candidate.id === item.productId);
									return `${item.quantity}× ${product?.title ?? item.productId}`;
								})
								.join(" · ") || "—",
					},
				]
			: []),
	];

	return (
		<div className="mx-auto w-full max-w-3xl space-y-5">
			<PageHeader
				title={mode === "edit" ? t("products.editTitle") : t("products.addTitle")}
				description={t("wizard.subtitle")}
				actions={
					<Link
						href="/seller/products"
						className="inline-flex items-center gap-1.5 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
					>
						<ArrowLeft className="size-4 rtl:rotate-180" />
						{t("actions.cancel")}
					</Link>
				}
			/>

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
				<ol className="mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4" aria-label={t("wizard.progressLabel")}>
					{readiness.checks.map((check, index) => {
						const isCurrent = index === step;
						return (
							<li
								key={check.label}
								aria-current={isCurrent ? "step" : undefined}
								className={`rounded-md border px-2.5 py-2 ${
									check.done
										? "border-emerald-600/30 bg-emerald-600/5 text-foreground"
										: isCurrent
											? "border-primary bg-primary/5 font-semibold text-foreground"
											: "border-border text-muted-foreground"
								}`}
							>
								<span className="flex items-center gap-1.5">
									{check.done ? <Check className="size-3.5 shrink-0 text-emerald-600" /> : <span className="font-bold">{index + 1}</span>}
									{check.label}
								</span>
							</li>
						);
					})}
				</ol>
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
									<Select value={formData.categoryId} onValueChange={(value) => setField("categoryId", value)}>
										<SelectTrigger id="wizard-category">
											<SelectValue placeholder={t("products.form.chooseCategory")} />
										</SelectTrigger>
										<SelectContent>
											{categories.map((c) => (
												<SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
											))}
										</SelectContent>
									</Select>
									{fieldError("categoryId")}
								</div>
								<div className="space-y-1">
									<label htmlFor="wizard-brand" className="text-sm font-medium text-foreground">{t("products.form.brand")}</label>
									<Select value={formData.brandId} onValueChange={(value) => setField("brandId", value === "__none" ? "" : value)}>
										<SelectTrigger id="wizard-brand">
											<SelectValue placeholder={t("products.form.noBrand")} />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="__none">{t("products.form.noBrand")}</SelectItem>
											{brands.map((b) => (
												<SelectItem key={b.id} value={b.id}>{b.name}</SelectItem>
											))}
										</SelectContent>
									</Select>
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
								<Select value={formData.productType} onValueChange={(value) => setField("productType", value as ProductForm["productType"])}>
									<SelectTrigger id="wizard-type">
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="simple">{t("products.form.simple")}</SelectItem>
										<SelectItem value="variable">{t("products.form.variable")}</SelectItem>
										{mode === "create" && <SelectItem value="bundle">{t("products.form.bundle")}</SelectItem>}
										{mode === "create" && <SelectItem value="downloadable">{t("products.form.downloadable")}</SelectItem>}
									</SelectContent>
								</Select>
							</div>
						</div>
						{isDownloadable && (
							<div className="space-y-3 rounded-md border border-border p-4">
								<div className="space-y-1">
									<label htmlFor="wizard-download-url" className="text-sm font-medium text-foreground">{t("products.form.downloadUrl")}</label>
									<input
										id="wizard-download-url"
										dir="ltr"
										value={formData.downloadUrl}
										onChange={(e) => setField("downloadUrl", e.target.value)}
										placeholder={t("products.form.downloadUrlPlaceholder")}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									/>
									<p className="text-xs text-muted-foreground">{t("products.form.downloadUrlHint")}</p>
									{fieldError("downloadUrl")}
								</div>
								<div className="space-y-1">
									<label htmlFor="wizard-delivery-instructions" className="text-sm font-medium text-foreground">
										{t("products.form.deliveryInstructions")}
									</label>
									<textarea
										id="wizard-delivery-instructions"
										rows={3}
										value={formData.deliveryInstructions}
										onChange={(e) => setField("deliveryInstructions", e.target.value)}
										placeholder={t("products.form.deliveryInstructionsPlaceholder")}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									/>
								</div>
							</div>
						)}
						{formData.productType === "bundle" && (
							<BundleItems
								products={bundleCandidates}
								items={formData.bundleItems}
								currency={formData.currency}
								loading={bundleCandidatesState === "loading"}
								onChange={(items) => setField("bundleItems", items)}
								error={fieldError("bundleItems")}
							/>
						)}
						{formData.productType === "variable" && (
							<div className="space-y-3 rounded-md border border-border p-4">
								<div className="flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 className="text-sm font-semibold text-foreground">{t("wizard.attributes.title")}</h3>
										<p className="text-xs text-muted-foreground">{t("wizard.attributes.subtitle")}</p>
									</div>
									<button
										type="button"
										disabled={variationRows.length >= Math.max(categoryAttributes.length, 1) && categoryAttributes.length > 0}
										onClick={() => setVariationRows((current) => [...current, { key: "", value: "" }])}
										className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
									>
										<Plus className="size-3.5" />
										{t("wizard.attributes.add")}
									</button>
								</div>

								{attributesLoading && (
									<p className="text-xs text-muted-foreground">{t("wizard.axes.loading")}</p>
								)}
								{attributesError && (
									<p className="text-xs text-destructive">{t("wizard.axes.loadFailed")}</p>
								)}
								{!attributesLoading && !attributesError && !categoryAttributes.length && (
									<p className="rounded-md bg-secondary/50 p-3 text-xs text-muted-foreground">{t("wizard.axes.noneDescription")}</p>
								)}

								{!attributesLoading && !attributesError && formData.categoryId && (
									<CustomAttributeBuilder
										categoryId={formData.categoryId}
										onCreated={(attribute) => setCategoryAttributes((current) => [...current, attribute])}
									/>
								)}
								
								{variationRows.map((row, rowIndex) => {
									const selectedAttribute = categoryAttributes.find((attribute) => attribute.key === row.key);
									const usedKeys = variationRows.map((item) => item.key).filter(Boolean);
									const available = categoryAttributes.filter(
										(attribute) => attribute.key === row.key || !usedKeys.includes(attribute.key),
									);
									const updateRow = (nextRow: VariationRow) => {
										const next = [...variationRows];
										next[rowIndex] = nextRow;
										setVariationRows(next);
									};
									return (
										<div key={rowIndex} className="rounded-lg border border-border p-3">
											<div className="flex items-center gap-2">
												{categoryAttributes.length > 0 ? (
													<Select value={row.key} onValueChange={(value) => updateRow({ ...row, key: value, value: "" })}>
														<SelectTrigger aria-label={`${t("products.form.variationAxes")} ${rowIndex + 1}`}>
															<SelectValue placeholder={t("wizard.attributes.chooseAttribute")} />
														</SelectTrigger>
														<SelectContent>
															{available.map((attribute) => (
																<SelectItem key={attribute.key} value={attribute.key}>
																	{attribute.label} · {attribute.key}
																</SelectItem>
															))}
														</SelectContent>
													</Select>
												) : (
													<input
														value={row.key}
														onChange={(e) => updateRow({ ...row, key: e.target.value.trim() })}
														placeholder={t("products.form.variationAxes")}
														className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
														aria-label={`${t("products.form.variationAxes")} ${rowIndex + 1}`}
													/>
												)}
												<button
													type="button"
													onClick={() => setVariationRows(variationRows.filter((_, index) => index !== rowIndex))}
													className="rounded-md border border-border p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
													aria-label={t("actions.delete")}
												>
													<Trash2 className="size-4" />
												</button>
											</div>

											{row.key && (
												<div className="mt-3">
													<p className="text-xs font-medium text-muted-foreground">{t("wizard.attributes.values")}</p>
													{selectedAttribute?.allowedValues?.length && isColorAttribute(row.key, selectedAttribute.allowedValues) ? (
														<>
															<div className="mt-2 flex flex-wrap gap-2">
																{selectedAttribute.allowedValues.map((value) => {
																	const selected = row.value === value;
																	const swatchKey = `${row.key}=${value}`;
																	const swatchImage = swatchImages[swatchKey];
																	const hex = colorHexFor(value);
																	return (
																		<button
																			key={value}
																			type="button"
																			title={value}
																			onClick={() => updateRow({ ...row, value })}
																			className={`flex items-center gap-1.5 rounded-full border px-2 py-1 text-xs transition ${selected ? "border-primary bg-primary/10 font-semibold text-foreground ring-1 ring-primary" : "border-border text-muted-foreground hover:text-foreground"}`}
																		>
																			{swatchImage ? (
																				<img src={swatchImage} alt={value} className="size-5 rounded-full object-cover" />
																			) : (
																				<span
																					className="size-5 rounded-full border border-border"
																					style={{ backgroundColor: hex ?? "transparent" }}
																				/>
																			)}
																			{value}
																		</button>
																	);
																})}
															</div>
															{row.value && (
																<div className="mt-2 flex flex-wrap items-center gap-2">
																	{swatchImages[`swatch:${row.key}=${row.value}`] ? (
																		<>
																			<img src={swatchImages[`swatch:${row.key}=${row.value}`]} alt={row.value} className="size-10 rounded-md border border-border object-cover" />
																			<button
																				type="button"
																				onClick={() =>
																					setSwatchImages((current) => {
																						const next = { ...current };
																						delete next[`${row.key}=${row.value}`];
																						return next;
																					})
																				}
																				className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-muted-foreground hover:text-destructive"
																			>
																				{t("wizard.variants.removeThumbnail")}
																			</button>
																		</>
																	) : (
																		<label className="inline-flex cursor-pointer items-center rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary">
																			{uploadingThumbnailKey === `swatch:${row.key}=${row.value}` ? <Loader2 className="size-3.5 animate-spin" /> : null}
																			{t("wizard.variants.chooseThumbnail")}
																			<input
																				type="file"
																				accept="image/*"
																				disabled={uploadingThumbnailKey === `swatch:${row.key}=${row.value}`}
																				className="sr-only"
																				onChange={(e) => {
																					void uploadSwatchImage(`swatch:${row.key}=${row.value}`, e.target.files?.[0]);
																					e.target.value = "";
																				}}
																			/>
																		</label>
																	)}
																</div>
															)}
														</>
													) : selectedAttribute?.allowedValues?.length ? (
														<Select value={row.value} onValueChange={(value) => updateRow({ ...row, value })}>
															<SelectTrigger className="mt-2" aria-label={`${t("wizard.attributes.values")} ${rowIndex + 1}`}>
																<SelectValue placeholder={t("wizard.attributes.chooseValue")} />
															</SelectTrigger>
															<SelectContent>
																{selectedAttribute.allowedValues.map((value) => (
																	<SelectItem key={value} value={value}>{value}</SelectItem>
																))}
															</SelectContent>
														</Select>
													) : (
														<input
															value={row.value}
															onChange={(e) => updateRow({ ...row, value: e.target.value })}
															placeholder={t("wizard.attributes.customValuePlaceholder")}
															className="mt-2 w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
														/>
													)}
												</div>
											)}
										</div>
									);
								})}

								{variationRows.length > 0 && (
									<button
										type="button"
										disabled={!variationRows.every((row) => row.key && row.value)}
										onClick={saveCombination}
										className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-xs font-medium text-primary-foreground disabled:opacity-50"
									>
										<Check className="size-3.5" />
										{t("wizard.attributes.saveCombination")}
									</button>
								)}

								{savedCombinations.length > 0 && (
									<div className="space-y-2">
										<p className="text-xs font-semibold text-foreground">
											{t("wizard.attributes.savedTitle")} · {t("wizard.attributes.combinations", { count: savedCombinations.length })}
										</p>
										<div className="overflow-x-auto rounded-md border border-border">
											<table className="w-full text-start text-xs">
												<thead>
													<tr className="border-b border-border bg-secondary/40">
														{uniqueKeys(savedCombinations).map((key) => (
															<th key={key} className="px-3 py-2 text-start font-medium text-muted-foreground">
																{categoryAttributes.find((attribute) => attribute.key === key)?.label ?? key}
															</th>
														))}
														<th className="px-3 py-2" />
													</tr>
												</thead>
												<tbody className="divide-y divide-border">
													{savedCombinations.map((combination, index) => (
														<tr key={combinationKey(combination)}>
															{uniqueKeys(savedCombinations).map((key) => (
																<td key={key} className="px-3 py-2 text-foreground">
																	{combination[key] ?? "—"}
																</td>
															))}
															<td className="px-3 py-2 text-end">
																<button
																	type="button"
																	onClick={() => removeSavedCombination(index)}
																	className="rounded-md border border-border p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
																	aria-label={t("actions.delete")}
																>
																	<Trash2 className="size-3.5" />
																</button>
															</td>
														</tr>
													))}
												</tbody>
											</table>
										</div>
									</div>
								)}
								{fieldError("variationAxes")}
							</div>
						)}
					</div>
				)}

				{step === 2 && (
					<div className="mt-5 space-y-4">
					{!(mode === "create" && formData.productType === "variable") && (<>
						<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
							{!hidesInventory && (
								<div className="space-y-1">
									<label htmlFor="wizard-sku" className="text-sm font-medium text-foreground">{t("products.form.skuCode")}</label>
									<input
										id="wizard-sku"
										value={formData.sku}
										onChange={(e) => setField("sku", e.target.value)}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									/>
								</div>
							)}
							<div className="space-y-1">
								<label htmlFor="wizard-price" className="text-sm font-medium text-foreground">
									{t(isBundle ? "bundles.bundlePrice" : "products.form.sellingPrice")}
								</label>
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
							{!hidesInventory && (
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
							)}
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
							{!hidesInventory && (
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
							)}
						</div>
						{!hidesInventory && (
							<p className="text-xs text-muted-foreground">{t("products.form.initialPurchaseNote")}</p>
						)}
						</>)}
					{mode === "create" && formData.productType === "variable" && (
						<VariantPricing
							currency={formData.currency}
							onCurrencyChange={(value) => setField("currency", value)}
							currencyError={fieldError("currency")}
							combinations={combinations}
							variantOverrides={variantOverrides}
							setVariantOverrides={setVariantOverrides}
							uploadingThumbnailKey={uploadingThumbnailKey}
							onUploadThumbnail={uploadVariantThumbnail}
						/>
					)}
					</div>
				)}

				{step === 3 && (
					<div className="mt-5 space-y-4">
						{mode === "create" && (
							<div className="rounded-md border border-border">
								<p className="border-b border-border px-4 py-2.5 text-sm font-semibold text-foreground">{t("wizard.attributes.previewTitle")}</p>
								<div className="grid grid-cols-1 gap-4 p-4 md:grid-cols-2">
									<ProductGallery
										images={previewImages}
										productName={formData.title}
										activeImage={previewImageIndex}
										onSelectImage={(index) => {
											setPreviewImageIndex(index);
											const combination = previewCombinations[index];
											if (combination) setPreviewValues(combination);
										}}
										discountPercentage={previewDiscount}
									/>
									<div className="space-y-3">
										<div>
											<p className="text-lg font-bold text-foreground">{formData.title || "—"}</p>
											{previewSku && <p className="text-xs text-muted-foreground" dir="ltr">{previewSku}</p>}
										</div>
										<div className="flex flex-wrap items-baseline gap-2">
											{previewPrice ? (
												<span className="text-xl font-bold text-foreground">
													{formData.currency.toUpperCase()} {Number(previewPrice).toFixed(2)}
												</span>
											) : (
												<span className="text-xl font-bold text-foreground">—</span>
											)}
											{previewCompareAt && previewDiscount && (
												<span className="text-sm text-muted-foreground line-through">
													{formData.currency.toUpperCase()} {Number(previewCompareAt).toFixed(2)}
												</span>
											)}
											{previewDiscount && (
												<span className="rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-semibold text-destructive">
													-{previewDiscount}
												</span>
											)}
										</div>
										{previewAxes.map((axis) => {
											const availableValues = [
												...new Set(
													previewCombinations
														.filter((combination) =>
															previewAxes.every(
																(other) => other === axis || previewValues[other] === undefined || combination[other] === previewValues[other],
															),
														)
														.map((combination) => combination[axis])
														.filter((value): value is string => Boolean(value)),
												),
											];
											return (
												<div key={axis} className="space-y-1">
													<label className="text-xs font-medium text-muted-foreground">
														{categoryAttributes.find((attribute) => attribute.key === axis)?.label ?? axis}
													</label>
													<Select
														value={previewValues[axis] ?? previewCombination?.[axis] ?? ""}
														onValueChange={(value) => {
															setPreviewValues((current) => ({ ...current, [axis]: value }));
															setPreviewImageIndex(Math.max(previewCombinations.findIndex((combination) => combination[axis] === value), 0));
														}}
													>
														<SelectTrigger>
															<SelectValue />
														</SelectTrigger>
														<SelectContent>
															{availableValues.map((value) => (
																<SelectItem key={value} value={value}>{value}</SelectItem>
															))}
														</SelectContent>
													</Select>
												</div>
											);
										})}
										<p className="text-xs text-muted-foreground">
											{t("products.form.initialStock")}: {previewStock || "—"}
										</p>
									</div>
								</div>
							</div>
						)}
						<dl className="divide-y divide-border rounded-md border border-border text-sm">
							{reviewRows.map((row) => (
								<div key={row.label} className="flex items-center justify-between gap-4 px-4 py-2.5">
									<dt className="text-muted-foreground">{row.label}</dt>
									<dd className="truncate text-end font-medium text-foreground">{row.value}</dd>
								</div>
							))}
						</dl>
						{mode === "create" && formData.productType === "variable" && combinations.length > 0 && (
							<div className="rounded-md border border-border">
								<p className="border-b border-border px-4 py-2.5 text-sm font-semibold text-foreground">{t("wizard.variants.title")}</p>
								<ul className="divide-y divide-border text-sm">
									{combinations.map((combination) => {
										const override = variantOverrides[combinationKey(combination)];
										const price = override?.price;
										const stock = override?.stock;
										return (
											<li key={combinationKey(combination)} className="flex items-center justify-between gap-4 px-4 py-2.5">
												<span className="text-muted-foreground">{formatCombination(combination)}</span>
												<span className="font-medium text-foreground">
													{price ? `${formData.currency.toUpperCase()} ${Number(price).toFixed(2)}` : "—"}
													{stock ? ` · ${stock}` : ""}
												</span>
											</li>
										);
									})}
								</ul>
							</div>
						)}
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
