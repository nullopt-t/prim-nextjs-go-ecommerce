"use client";

import { useCallback, useEffect, useState } from "react";
import { CheckCircle2, FileText, Loader2, Plus, Trash2, Upload, XCircle } from "lucide-react";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerDocument, SellerDocumentType } from "@/modules/seller/types";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { ApiError } from "@/shared/api/client";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { formatDateTime } from "@/shared/lib/datetime";
import { normalizeSlug, validateSlug } from "@/modules/seller/lib/slug";
import { KYC_FILE_EXTENSIONS, validateKycFile } from "@/modules/seller/lib/kycFileValidation";
import { useTranslations, useLocale } from "next-intl";
import { useSellerContext } from "./sellerContext";

type SlugStatus = "idle" | "checking" | "available" | "unavailable";

interface BankEntry {
	key: string;
	value: string;
}

function toEntries(info?: Record<string, unknown>): BankEntry[] {
	if (!info) return [];
	return Object.entries(info).map(([key, value]) => ({
		key,
		value:
			typeof value === "string" || typeof value === "number" || typeof value === "boolean"
				? String(value)
				: JSON.stringify(value),
	}));
}

export function SellerSettings() {
	const t = useTranslations("seller");
	const locale = useLocale();
	const { seller, refreshSeller } = useSellerContext();

	const [storeName, setStoreName] = useState("");
	const [slug, setSlug] = useState("");
	const [description, setDescription] = useState("");
	const [legalName, setLegalName] = useState("");
	const [taxNumber, setTaxNumber] = useState("");
	const [commercialRegister, setCommercialRegister] = useState("");
	const [documents, setDocuments] = useState<SellerDocument[]>([]);
	const [uploadingType, setUploadingType] = useState<SellerDocumentType | null>(null);
	const [bankEntries, setBankEntries] = useState<BankEntry[]>([]);
	const [slugStatus, setSlugStatus] = useState<SlugStatus>("idle");
	const [slugReason, setSlugReason] = useState<string | null>(null);
	const [saving, setSaving] = useState(false);
	const [initializedFor, setInitializedFor] = useState<string | null>(null);

	const slugEditable = seller?.status !== "active";

	useEffect(() => {
		if (!seller || initializedFor === seller.id) return;
		setStoreName(seller.storeName ?? "");
		setSlug(seller.slug ?? "");
		setDescription(seller.description ?? "");
		setLegalName(seller.legalName ?? "");
		setTaxNumber(seller.taxNumber ?? "");
		setCommercialRegister(seller.commercialRegister ?? "");
		setBankEntries(toEntries(seller.bankAccountInfo));
		setInitializedFor(seller.id);
	}, [seller, initializedFor]);

	const loadDocuments = useCallback(async () => {
		try {
			const response = await sellerService.getDocuments();
			setDocuments(Array.isArray(response?.data) ? response.data : []);
		} catch {
			setDocuments([]);
		}
	}, []);

	useEffect(() => {
		void loadDocuments();
	}, [loadDocuments]);

	const handleDocumentUpload = async (type: SellerDocumentType, file?: File) => {
		if (!file) return;
		const fileError = validateKycFile(file);
		if (fileError) {
			notify.error(t(`settings.kyc.${fileError}`));
			return;
		}
		const formData = new FormData();
		formData.append("documentType", type);
		formData.append("file", file);
		setUploadingType(type);
		try {
			await sellerService.uploadDocument(formData);
			notify.success(t("settings.kyc.uploaded"));
			await loadDocuments();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("settings.kyc.uploadFailed")));
		} finally {
			setUploadingType(null);
		}
	};

	const handleDocumentDelete = async (id: string) => {
		try {
			await sellerService.deleteDocument(id);
			notify.success(t("settings.kyc.deleted"));
			await loadDocuments();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("settings.kyc.deleteFailed")));
		}
	};

	useEffect(() => {
		if (!slugEditable) return;
		const normalized = normalizeSlug(slug);
		if (!normalized || normalized === normalizeSlug(seller?.slug ?? "")) {
			setSlugStatus("idle");
			setSlugReason(null);
			return;
		}
		const validation = validateSlug(normalized);
		if (validation.reason) {
			setSlugStatus("unavailable");
			setSlugReason(validation.reason);
			return;
		}
		let cancelled = false;
		setSlugStatus("checking");
		setSlugReason(null);
		const timer = window.setTimeout(() => {
			void sellerService
				.checkSlugAvailability(normalized)
				.then((response) => {
					if (cancelled) return;
					setSlugStatus(response.data.available ? "available" : "unavailable");
					setSlugReason(response.data.reason ?? null);
					if (response.data.available) setSlug(normalized);
				})
				.catch((error: unknown) => {
					if (cancelled) return;
					setSlugStatus("unavailable");
					setSlugReason(apiErrorMessage(error, t("register.slugCheckFailed")));
				});
		}, 400);
		return () => {
			cancelled = true;
			window.clearTimeout(timer);
		};
	}, [slug, slugEditable, seller?.slug, t]);

	if (!seller) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("settings.title")} />
				<ErrorState
					title={t("profileLoadTitle")}
					description={t("profileLoadDescription")}
					onRetry={() => void refreshSeller()}
				/>
			</div>
		);
	}

	const handleSubmit = async (event: React.FormEvent) => {
		event.preventDefault();
		if (storeName.trim().length < 2) {
			notify.error(t("settings.storeNameInvalid"));
			return;
		}
		if (description.length > 2000) {
			notify.error(t("settings.descriptionInvalid"));
			return;
		}
		if (slugEditable) {
			const validation = validateSlug(slug);
			if (validation.reason) {
				setSlugStatus("unavailable");
				setSlugReason(validation.reason);
				return;
			}
			if (slugStatus === "checking") return;
		}

		const bankInfo: Record<string, unknown> = {};
		for (const entry of bankEntries) {
			const key = entry.key.trim();
			if (!key) {
				notify.error(t("settings.bankKeyRequired"));
				return;
			}
			if (key in bankInfo) {
				notify.error(t("settings.bankKeyDuplicate", { key }));
				return;
			}
			bankInfo[key] = entry.value;
		}

		setSaving(true);
		try {
			const payload: {
				storeName?: string;
				slug?: string;
				description?: string;
				legalName?: string;
				taxNumber?: string;
				commercialRegister?: string;
				bankAccountInfo?: Record<string, unknown>;
			} = {};
			if (storeName.trim() !== seller.storeName) payload.storeName = storeName.trim();
			if (slugEditable && normalizeSlug(slug) !== normalizeSlug(seller.slug)) {
				payload.slug = normalizeSlug(slug);
			}
			if (description.trim() !== (seller.description ?? "")) {
				payload.description = description.trim();
			}
			if (legalName.trim() !== (seller.legalName ?? "")) payload.legalName = legalName.trim();
			if (taxNumber.trim() !== (seller.taxNumber ?? "")) payload.taxNumber = taxNumber.trim();
			if (commercialRegister.trim() !== (seller.commercialRegister ?? "")) {
				payload.commercialRegister = commercialRegister.trim();
			}
			payload.bankAccountInfo = bankInfo;

			await sellerService.updateProfile(payload);
			notify.success(t("settings.saved"));
			await refreshSeller();
		} catch (error) {
			if (error instanceof ApiError && error.code === "SELLER_SLUG_TAKEN") {
				setSlugStatus("unavailable");
				setSlugReason(t("register.slugTaken"));
			}
			notify.error(apiErrorMessage(error, t("settings.saveFailed")));
		} finally {
			setSaving(false);
		}
	};

	return (
		<div className="space-y-5">
			<PageHeader
				title={t("settings.title")}
				description={t("settings.subtitle")}
			/>

			<div className="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1fr)_320px]">
				<form
					onSubmit={(event) => void handleSubmit(event)}
					className="space-y-5 rounded-lg border border-border bg-background p-5"
				>
					<div className="space-y-1.5">
						<label htmlFor="seller-settings-store-name" className="text-sm font-medium text-foreground">
							{t("register.storeName")}
						</label>
						<input
							id="seller-settings-store-name"
							type="text"
							required
							minLength={2}
							maxLength={150}
							value={storeName}
							onChange={(event) => setStoreName(event.target.value)}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>

					<div className="space-y-1.5">
						<label htmlFor="seller-settings-slug" className="text-sm font-medium text-foreground">
							{t("register.slug")}
						</label>
						<div className="relative">
							<span className="pointer-events-none absolute inset-y-0 start-3 flex items-center text-sm text-muted-foreground">
								/store/
							</span>
							<input
								id="seller-settings-slug"
								type="text"
								required
								minLength={3}
								maxLength={63}
								dir="ltr"
								disabled={!slugEditable}
								value={slug}
								onChange={(event) => setSlug(event.target.value)}
								onBlur={() => setSlug(normalizeSlug(slug))}
								placeholder={t("register.slugPlaceholder")}
								aria-invalid={slugEditable && slugStatus === "unavailable"}
								aria-describedby="seller-settings-slug-status"
								className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-16 text-sm text-foreground disabled:bg-secondary/40 disabled:text-muted-foreground"
							/>
						</div>
						{slugEditable ? (
							<div
								id="seller-settings-slug-status"
								className={`flex items-start gap-1.5 text-xs ${slugStatus === "unavailable" ? "text-destructive" : "text-muted-foreground"}`}
							>
								{slugStatus === "checking" ? <Loader2 className="mt-0.5 size-3.5 animate-spin" /> : null}
								{slugStatus === "available" ? <CheckCircle2 className="mt-0.5 size-3.5 text-emerald-600" /> : null}
								{slugStatus === "unavailable" ? <XCircle className="mt-0.5 size-3.5" /> : null}
								<span>
									{slugStatus === "checking"
										? t("register.slugChecking")
										: slugStatus === "available"
											? t("register.slugAvailable")
											: slugReason || t("register.slugHint")}
								</span>
							</div>
						) : (
							<p className="text-xs text-muted-foreground">{t("settings.slugImmutable")}</p>
						)}
					</div>

					<div className="space-y-1.5">
						<label htmlFor="seller-settings-description" className="text-sm font-medium text-foreground">
							{t("register.description")}
						</label>
						<textarea
							id="seller-settings-description"
							value={description}
							onChange={(event) => setDescription(event.target.value)}
							placeholder={t("register.descriptionPlaceholder")}
							rows={4}
							maxLength={2000}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
						<p className="text-xs text-muted-foreground">{t("settings.descriptionHint", { count: description.length })}</p>
					</div>

					<fieldset className="space-y-3 rounded-md border border-border p-4">
						<legend className="px-1 text-sm font-medium text-foreground">{t("settings.kyc.title")}</legend>
						<p className="text-xs text-muted-foreground">{t("settings.kyc.subtitle")}</p>
						<div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
							<div className="space-y-1.5">
								<label htmlFor="seller-kyc-legal-name" className="text-sm font-medium text-foreground">
									{t("settings.kyc.legalName")}
								</label>
								<input
									id="seller-kyc-legal-name"
									type="text"
									maxLength={200}
									value={legalName}
									onChange={(event) => setLegalName(event.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
							</div>
							<div className="space-y-1.5">
								<label htmlFor="seller-kyc-tax-number" className="text-sm font-medium text-foreground">
									{t("settings.kyc.taxNumber")}
								</label>
								<input
									id="seller-kyc-tax-number"
									type="text"
									dir="ltr"
									maxLength={100}
									value={taxNumber}
									onChange={(event) => setTaxNumber(event.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
							</div>
							<div className="space-y-1.5">
								<label htmlFor="seller-kyc-commercial-register" className="text-sm font-medium text-foreground">
									{t("settings.kyc.commercialRegister")}
								</label>
								<input
									id="seller-kyc-commercial-register"
									type="text"
									dir="ltr"
									maxLength={100}
									value={commercialRegister}
									onChange={(event) => setCommercialRegister(event.target.value)}
									className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
								/>
							</div>
						</div>
						<div className="space-y-2">
							{([
								{ type: "commercial_register", label: t("settings.kyc.docCommercialRegister") },
								{ type: "tax_card", label: t("settings.kyc.docTaxCard") },
								{ type: "national_id", label: t("settings.kyc.docNationalId") },
							] as const).map(({ type, label }) => {
								const doc = documents.find((item) => item.documentType === type);
								return (
									<div key={type} className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border p-3">
										<div className="min-w-0">
											<p className="text-sm font-medium text-foreground">{label}</p>
											{doc ? (
												<p className="truncate text-xs text-muted-foreground">
													{doc.fileName} · {Math.round(doc.fileSize / 1024)} KB
												</p>
											) : (
												<p className="text-xs text-muted-foreground">{t("settings.kyc.notUploaded")}</p>
											)}
											{doc?.status === "rejected" ? (
												<div className="mt-2 rounded-md border border-destructive/30 bg-destructive/5 p-2">
													<p className="flex items-center gap-1 text-xs font-medium text-destructive">
														<XCircle className="size-3.5" />
														{t("settings.kyc.rejectionReason")}
													</p>
													<p className="mt-1 text-sm text-destructive">
														{doc.rejectionReason || t("settings.kyc.noReason")}
													</p>
												</div>
											) : null}
										</div>
										<div className="flex items-center gap-2">
											{doc?.status === "approved" ? (
												<span className="inline-flex items-center gap-1 rounded-md border border-emerald-600/20 bg-emerald-600/10 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:text-emerald-400">
													<CheckCircle2 className="size-3.5" />
													{t("settings.kyc.verified")}
												</span>
											) : doc ? (
												<StatusPill tone={statusTone(doc.status)}>
													{t(`settings.kyc.status.${doc.status}`)}
												</StatusPill>
											) : null}
											{doc?.documentUrl ? (
												<a
													href={normalizeMediaUrl(doc.documentUrl)}
													target="_blank"
													rel="noopener noreferrer"
													aria-label={`${t("settings.kyc.view")} ${label}`}
													className="rounded-md border border-border p-2 text-muted-foreground hover:text-foreground"
												>
													<FileText className="size-4" />
												</a>
											) : null}
											{doc?.status !== "approved" ? (
												<label className="inline-flex cursor-pointer items-center rounded-md border border-border p-2 text-muted-foreground hover:text-foreground">
													{uploadingType === type ? (
														<Loader2 className="size-4 animate-spin" />
													) : (
														<Upload className="size-4" />
													)}
													<span className="sr-only">{t("settings.kyc.upload")}</span>
													<input
														type="file"
														accept={KYC_FILE_EXTENSIONS.join(",")}
														className="sr-only"
														disabled={uploadingType === type}
														onChange={(event) => {
															void handleDocumentUpload(type, event.target.files?.[0]);
															event.target.value = "";
														}}
													/>
												</label>
											) : null}
											{doc && doc.status !== "approved" ? (
												<button
													type="button"
													onClick={() => void handleDocumentDelete(doc.id)}
													aria-label={t("settings.kyc.remove")}
													className="rounded-md border border-border p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
												>
													<Trash2 className="size-4" />
												</button>
											) : null}
										</div>
									</div>
								);
							})}
						</div>
					</fieldset>

					<fieldset className="space-y-3 rounded-md border border-border p-4">
						<legend className="px-1 text-sm font-medium text-foreground">{t("settings.bankTitle")}</legend>
						<p className="text-xs text-muted-foreground">{t("settings.bankNote")}</p>
						{bankEntries.length === 0 ? (
							<p className="rounded-md border border-dashed border-border p-3 text-xs text-muted-foreground">
								{t("settings.bankEmpty")}
							</p>
						) : null}
						{bankEntries.map((entry, index) => (
							<div key={index} className="grid grid-cols-1 items-end gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
								<div className="space-y-1">
									<label className="text-xs font-medium text-muted-foreground">
										{t("settings.bankKey")}
									</label>
									<input
										type="text"
										value={entry.key}
										onChange={(event) =>
											setBankEntries((current) =>
												current.map((item, i) => (i === index ? { ...item, key: event.target.value } : item)),
											)
										}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
										aria-label={`${t("settings.bankKey")} ${index + 1}`}
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs font-medium text-muted-foreground">
										{t("settings.bankValue")}
									</label>
									<input
										type="text"
										value={entry.value}
										onChange={(event) =>
											setBankEntries((current) =>
												current.map((item, i) => (i === index ? { ...item, value: event.target.value } : item)),
											)
										}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
										aria-label={`${t("settings.bankValue")} ${index + 1}`}
									/>
								</div>
								<button
									type="button"
									onClick={() => setBankEntries((current) => current.filter((_, i) => i !== index))}
									className="rounded-md border border-border p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
									aria-label={t("actions.delete")}
								>
									<Trash2 className="size-4" />
								</button>
							</div>
						))}
						<button
							type="button"
							onClick={() => setBankEntries((current) => [...current, { key: "", value: "" }])}
							className="inline-flex items-center gap-1.5 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
						>
							<Plus className="size-4" />
							{t("settings.bankAdd")}
						</button>
					</fieldset>

					{seller.status === "rejected" ? (
						<p className="rounded-md border border-border bg-secondary/40 p-3 text-xs text-muted-foreground">
							{t("settings.resubmitNote")}
						</p>
					) : null}

					<div className="flex justify-end border-t border-border pt-4">
						<button
							type="submit"
							disabled={saving || slugStatus === "checking"}
							className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
						>
							{saving ? <Loader2 className="size-4 animate-spin" /> : null}
							{t("actions.saveChanges")}
						</button>
					</div>
				</form>

				<aside className="space-y-4">
					<section className="rounded-lg border border-border bg-background p-4">
						<h2 className="text-sm font-semibold text-foreground">{t("settings.accountTitle")}</h2>
						<dl className="mt-3 space-y-3 text-sm">
							<div className="flex items-center justify-between gap-3">
								<dt className="text-muted-foreground">{t("settings.accountStatus")}</dt>
								<dd><StatusPill tone={statusTone(seller.status)}>{t(`status.${seller.status}`) ?? seller.status}</StatusPill></dd>
							</div>
							<div className="flex items-center justify-between gap-3">
								<dt className="text-muted-foreground">{t("layout.commissionRate")}</dt>
								<dd className="font-medium tabular-nums text-foreground">{(seller.commissionRateBps / 100).toFixed(2)}%</dd>
							</div>
							<div className="flex items-center justify-between gap-3">
								<dt className="text-muted-foreground">{t("settings.memberSince")}</dt>
								<dd className="text-foreground">{formatDateTime(seller.createdAt, locale)}</dd>
							</div>
							<div className="flex items-center justify-between gap-3">
								<dt className="text-muted-foreground">{t("settings.payoutDetails")}</dt>
								<dd className="text-foreground">
									{bankEntries.length > 0 ? t("settings.payoutDetailsSaved", { count: bankEntries.length }) : t("settings.payoutDetailsMissing")}
								</dd>
							</div>
						</dl>
						{seller.statusReason ? (
							<p className="mt-3 rounded-md border border-border bg-secondary/40 p-3 text-xs text-muted-foreground">
								{seller.statusReason}
							</p>
						) : null}
					</section>
					<p className="rounded-lg border border-border bg-secondary/30 p-3 text-xs text-muted-foreground">
						{t("settings.commissionNote")}
					</p>
				</aside>
			</div>
		</div>
	);
}
