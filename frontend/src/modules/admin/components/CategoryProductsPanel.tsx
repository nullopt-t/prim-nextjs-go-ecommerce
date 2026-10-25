"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { FolderTree, Loader2, Package, Pencil, RefreshCw, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { notify } from "@/shared/lib/notify";
import { normalizeMediaUrl } from "@/shared/lib/media";

export interface PanelCategory {
  id: string;
  name: string;
  parent?: string;
  parentId?: string | null;
  productsCount?: number;
}

interface CategoryProductsPanelProps {
  category: PanelCategory | null;
  categories: PanelCategory[];
  onEdit: () => void;
  onClear: () => void;
}

const PRODUCTS_PAGE_SIZE = 40;

/**
 * Lists the products attached to the category selected in the tree map.
 *
 * Selecting a parent includes products from its whole subtree, matching the
 * rolled-up counts shown on the map nodes. The panel is deliberately a plain
 * list: expanding products inside the map would turn the hierarchy into noise.
 */
export default function CategoryProductsPanel({
  category,
  categories,
  onEdit,
  onClear,
}: CategoryProductsPanelProps) {
  const t = useTranslations("admin");
  const [products, setProducts] = useState<any[]>([]);
  const [meta, setMeta] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [reloadToken, setReloadToken] = useState(0);
  const activeRequestKey = useRef("");

  // Include descendants so a parent's list is consistent with its rolled-up
  // product count on the map.
  const categoryIds = useMemo(() => {
    if (!category) return [];

    const childrenByParent = new Map<string, string[]>();
    for (const item of categories) {
      if (!item.parentId) continue;
      const children = childrenByParent.get(item.parentId) ?? [];
      children.push(item.id);
      childrenByParent.set(item.parentId, children);
    }

    const ids: string[] = [];
    const seen = new Set<string>();
    const stack = [category.id];
    while (stack.length > 0) {
      const id = stack.pop()!;
      if (seen.has(id)) continue;
      seen.add(id);
      ids.push(id);
      stack.push(...(childrenByParent.get(id) ?? []));
    }
    return ids;
  }, [category, categories]);

  const categoryKey = categoryIds.join(",");

  const totalProducts = useMemo(() => {
    const byId = new Map(categories.map((item) => [item.id, item]));
    return categoryIds.reduce((sum, id) => sum + (byId.get(id)?.productsCount ?? 0), 0);
  }, [categoryIds, categories]);

  useEffect(() => {
    if (!category) {
      activeRequestKey.current = "";
      setProducts([]);
      setMeta(null);
      setLoadError(false);
      setLoading(false);
      return;
    }

    activeRequestKey.current = categoryKey;
    let cancelled = false;
    setLoading(true);
    setLoadError(false);
    setProducts([]);
    setMeta(null);

    catalogService
      .getProducts({ category: categoryKey, pageSize: PRODUCTS_PAGE_SIZE, page: 1 })
      .then((response: any) => {
        if (cancelled || activeRequestKey.current !== categoryKey) return;
        const list = Array.isArray(response) ? response : Array.isArray(response?.data) ? response.data : [];
        setProducts(list);
        setMeta(response?.meta ?? null);
      })
      .catch(() => {
        if (!cancelled && activeRequestKey.current === categoryKey) setLoadError(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [category, categoryKey, reloadToken]);

  const loadMore = async () => {
    if (!category || loadingMore || !meta?.hasNext) return;

    const requestKey = categoryKey;
    const nextPage = (meta.page ?? 1) + 1;
    setLoadingMore(true);
    try {
      const response: any = await catalogService.getProducts({
        category: requestKey,
        pageSize: PRODUCTS_PAGE_SIZE,
        page: nextPage,
      });
      if (activeRequestKey.current !== requestKey) return;

      const list = Array.isArray(response) ? response : Array.isArray(response?.data) ? response.data : [];
      setProducts((previous) => [...previous, ...list]);
      setMeta(response?.meta ?? null);
    } catch {
      notify.error(t("categories.productsPanel.loadFailed"));
    } finally {
      setLoadingMore(false);
    }
  };

  if (!category) {
    return (
      <div className="flex h-[70vh] min-h-[520px] max-h-[820px] flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-border bg-secondary/10 p-8 text-center">
        <span className="flex size-12 items-center justify-center rounded-2xl border border-border bg-background text-muted-foreground">
          <FolderTree className="size-6" />
        </span>
        <p className="max-w-[220px] text-xs text-muted-foreground">{t("categories.productsPanel.hint")}</p>
      </div>
    );
  }

  return (
    <section className="flex h-[70vh] min-h-[520px] max-h-[820px] flex-col overflow-hidden rounded-2xl border border-border bg-background">
      <header className="flex items-start justify-between gap-2 border-b border-border p-4">
        <div className="min-w-0">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            {category.parent && category.parent !== "None"
              ? category.parent
              : t("categories.productsPanel.rootCategory")}
          </p>
          <h3 className="truncate text-sm font-bold text-foreground" title={category.name}>
            {category.name}
          </h3>
          <p className="mt-0.5 text-[11px] text-muted-foreground">
            {t("categories.items", { count: totalProducts })}
            {categoryIds.length > 1 ? ` · ${t("categories.productsPanel.includingSubcategories")}` : ""}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <button
            type="button"
            onClick={onEdit}
            className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-accent-brand"
            title={t("categories.productsPanel.edit")}
            aria-label={t("categories.productsPanel.edit")}
          >
            <Pencil className="size-4" />
          </button>
          <button
            type="button"
            onClick={onClear}
            className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
            title={t("categories.productsPanel.clear")}
            aria-label={t("categories.productsPanel.clear")}
          >
            <X className="size-4" />
          </button>
        </div>
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {loading ? (
          <div className="divide-y divide-border">
            {[0, 1, 2, 3].map((row) => (
              <div key={row} className="flex animate-pulse items-center gap-3 p-3">
                <div className="size-11 shrink-0 rounded-lg bg-secondary" />
                <div className="flex-1 space-y-2">
                  <div className="h-2.5 w-3/4 rounded bg-secondary" />
                  <div className="h-2 w-1/2 rounded bg-secondary" />
                </div>
              </div>
            ))}
          </div>
        ) : loadError ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
            <p className="text-xs text-muted-foreground">{t("categories.productsPanel.loadFailed")}</p>
            <button
              type="button"
              onClick={() => setReloadToken((token) => token + 1)}
              className="flex items-center gap-1.5 rounded-lg border border-border bg-secondary px-3 py-1.5 text-xs font-semibold text-foreground hover:bg-secondary/80"
            >
              <RefreshCw className="size-3.5" />
              {t("categories.productsPanel.retry")}
            </button>
          </div>
        ) : products.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
            <Package className="size-7 text-muted-foreground" />
            <p className="text-xs text-muted-foreground">{t("categories.productsPanel.empty")}</p>
          </div>
        ) : (
          <ul className="divide-y divide-border">
            {products.map((product) => (
              <li key={product.variantId || product.id} className="flex items-center gap-3 px-3 py-2.5">
                {product.thumbnail ? (
                  <img
                    src={normalizeMediaUrl(product.thumbnail)}
                    alt=""
                    loading="lazy"
                    className="size-11 shrink-0 rounded-lg border border-border object-cover"
                  />
                ) : (
                  <span className="flex size-11 shrink-0 items-center justify-center rounded-lg border border-border bg-secondary text-muted-foreground">
                    <Package className="size-4" />
                  </span>
                )}
                <div className="min-w-0 flex-1">
                  <p className="truncate text-xs font-semibold text-foreground" title={product.title}>
                    {product.title}
                  </p>
                  <p className="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
                    <span className="font-medium text-foreground">{product.price}</span>
                    <span aria-hidden>·</span>
                    <span className={product.inStock ? "" : "text-destructive"}>
                      {product.inStock
                        ? t("categories.productsPanel.inStock", { count: product.stockQuantity ?? 0 })
                        : t("categories.productsPanel.outOfStock")}
                    </span>
                  </p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {meta?.hasNext && !loading && !loadError && (
        <div className="border-t border-border p-2">
          <button
            type="button"
            onClick={loadMore}
            disabled={loadingMore}
            className="flex w-full items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-xs font-semibold text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground disabled:opacity-60"
          >
            {loadingMore ? <Loader2 className="size-3.5 animate-spin" /> : <RefreshCw className="size-3.5" />}
            {t("categories.productsPanel.loadMore")}
          </button>
        </div>
      )}
    </section>
  );
}
