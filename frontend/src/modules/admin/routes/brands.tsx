import { useState, useEffect, useCallback } from "react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { Brand } from "@/modules/catalog/types";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { revalidateCatalog } from "@/app/actions/revalidateCatalog";

const brandPlaceholder = (name: string) => {
	const initials =
		name
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((word) => word[0]?.toUpperCase() ?? "")
			.join("") || "?";
	const svg =
		`<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96">` +
		`<rect width="96" height="96" rx="20" fill="#eef2f7"/>` +
		`<text x="48" y="59" font-family="system-ui,-apple-system,sans-serif" font-size="34" font-weight="700" fill="#334155" text-anchor="middle">${initials}</text></svg>`;
	return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
};

export function AdminBrands() {
	const t = useTranslations("admin");
	const [brands, setBrands] = useState<any[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [loading, setLoading] = useState(true);
	const [search, setSearch] = useState("");
	const [isModalOpen, setIsModalOpen] = useState(false);
	const [editingBrand, setEditingBrand] = useState<any>(null);
	const [loadError, setLoadError] = useState(false);
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

	const loadBrands = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await catalogService.getBrands({ cursor: cursor || undefined, pageSize: 20, keyset: true, search: debouncedSearch || undefined });
			const list = res.data ?? [];
			setBrands(list.map((b: Brand) => ({
				id: b.id,
				name: b.name,
				website: b.link ?? "",
				productsCount: b.productCount ?? 0,
				logo: b.logoUrl ? normalizeMediaUrl(b.logoUrl) : brandPlaceholder(b.name)
			})));
			if ("meta" in res && res.meta) setMeta(res.meta as CursorMeta);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, [cursor, debouncedSearch]);

	useEffect(() => {
		loadBrands();
	}, [loadBrands]);

	useEffect(() => {
		const timer = setTimeout(() => {
			setDebouncedSearch(search.trim());
			reset();
		}, 300);
		return () => clearTimeout(timer);
	}, [search, reset]);

	const [formData, setFormData] = useState({ name: "", website: "", logo: "" });

	const handleOpenModal = (brand: any = null) => {
		if (brand) {
			setEditingBrand(brand);
			setFormData({ name: brand.name, website: brand.website, logo: brand.logo });
		} else {
			setEditingBrand(null);
			setFormData({ name: "", website: "", logo: "" });
		}
		setIsModalOpen(true);
	};

	const handleSave = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!formData.name) return;

		try {
			if (editingBrand) {
				await catalogService.updateBrand(editingBrand.id, {
					name: formData.name,
					link: formData.website || undefined
				});
				notify.success(t("brands.toast.updated"));
			} else {
				await catalogService.createBrand({
					name: formData.name,
					link: formData.website || undefined
				});
				notify.success(t("brands.toast.created"));
			}
			setIsModalOpen(false);
			void revalidateCatalog().catch(() => false);
			loadBrands();
		} catch (err: any) {
			notify.error(apiErrorMessage(err, t("brands.toast.saveFailed")));
		}
	};

	const handleDelete = async (id: string) => {
		if (confirm(t("brands.confirmDelete"))) {
			try {
				await catalogService.deleteBrand(id);
				notify.success(t("brands.toast.deleted"));
				void revalidateCatalog().catch(() => false);
				loadBrands();
			} catch (err: any) {
				notify.error(apiErrorMessage(err, t("brands.toast.deleteFailed")));
			}
		}
	};

	const filtered = brands;

	if (loadError) {
		return <ErrorState onRetry={loadBrands} />;
	}

	return (
		<div className="space-y-5">
			<PageHeader
				title={t("brands.title")}
				actions={
					<button
						onClick={() => handleOpenModal()}
						className="rounded-xl border border-border/70 bg-card px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all cursor-pointer"
					>
						{t("brands.add")}
					</button>
				}
			/>

			<input
				type="text"
				placeholder={t("brands.searchPlaceholder")}
				value={search}
				onChange={(e) => setSearch(e.target.value)}
				onKeyDown={(e) => {
					if (e.key === "Enter") {
						setDebouncedSearch(search.trim());
						reset();
					}
				}}
				className="w-full sm:max-w-xs rounded-xl border border-border/70 bg-card px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all focus:outline-none"
			/>

			<div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{filtered.map((b) => (
					<div key={b.id} className="flex flex-col justify-between gap-4 rounded-xl border border-border/70 bg-card p-4 hover:bg-muted/30 active:scale-[0.99] select-none transition-all">
						<div className="flex items-center gap-3">
							<img src={normalizeMediaUrl(b.logo)} alt="" className="size-10 rounded-md border border-border object-cover" />
							<div className="min-w-0">
								<h3 className="truncate text-sm font-medium text-foreground">{b.name}</h3>
								{b.website ? (
									<a
										href={b.website}
										target="_blank"
										rel="noreferrer"
										className="block truncate text-xs text-muted-foreground hover:text-foreground"
									>
										{b.website}
									</a>
								) : null}
							</div>
						</div>

						<div className="flex items-center justify-between border-t border-border/70 pt-3 text-xs">
							<span className="text-muted-foreground">{t("brands.associatedProducts", { count: b.productsCount })}</span>
							<div className="flex items-center gap-1">
								<button
									onClick={() => handleOpenModal(b)}
									className="rounded-lg px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/30 hover:text-foreground active:scale-[0.99] select-none transition-all"
								>
									{t("actions.edit")}
								</button>
								<button
									onClick={() => handleDelete(b.id)}
									className="rounded-lg px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/30 hover:text-destructive active:scale-[0.99] select-none transition-all"
								>
									{t("actions.delete")}
								</button>
							</div>
						</div>
					</div>
				))}
			</div>

			<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
				<CursorPager
					hasPrev={hasPrev}
					hasMore={meta.hasMore}
					page={page}
					onPrev={prev}
					onNext={() => next(meta.nextCursor)}
					loading={loading}
					shown={filtered.length}
				/>
			</div>

			{isModalOpen && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
					<div className="w-full max-w-md space-y-4 rounded-xl border border-border/70 bg-card p-5 shadow-xl">
						<div className="flex items-center justify-between border-b border-border/70 pb-3">
							<h3 className="text-sm font-semibold text-foreground">{editingBrand ? t("brands.editTitle") : t("brands.createTitle")}</h3>
							<button
								onClick={() => setIsModalOpen(false)}
								className="rounded-lg px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/30 hover:text-foreground active:scale-[0.99] select-none transition-all"
								aria-label={t("actions.cancel")}
							>
								{t("actions.close")}
							</button>
						</div>
						<form onSubmit={handleSave} className="space-y-4">
							<div className="space-y-1">
								<label className="text-sm font-medium text-foreground">{t("brands.form.name")}</label>
								<input
									type="text"
									required
									value={formData.name}
									onChange={(e) => setFormData({ ...formData, name: e.target.value })}
									className="w-full rounded-xl border border-border/70 bg-card px-3 py-2 text-sm text-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all"
								/>
							</div>
							<div className="space-y-1">
								<label className="text-sm font-medium text-foreground">{t("brands.form.website")}</label>
								<input
									type="text"
									value={formData.website}
									onChange={(e) => setFormData({ ...formData, website: e.target.value })}
									className="w-full rounded-xl border border-border/70 bg-card px-3 py-2 text-sm text-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all"
								/>
							</div>
							<div className="flex justify-end gap-2 border-t border-border/70 pt-3">
								<button type="button" onClick={() => setIsModalOpen(false)} className="rounded-xl border border-border/70 bg-card px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all">{t("actions.cancel")}</button>
								<button type="submit" className="rounded-xl bg-foreground px-3 py-2 text-sm font-medium text-background active:scale-[0.99] select-none transition-all">{t("actions.save")}</button>
							</div>
						</form>
					</div>
				</div>
			)}
		</div>
	);
}
