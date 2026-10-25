"use client";
import { revalidateCatalog } from "@/app/actions/revalidateCatalog";

import { useState, useEffect, useCallback, useMemo, useRef } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useParams, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  ArrowLeft,
  Archive,
  Check,
  ChevronRight,
  ChevronLeft,
  Layers,
  Sparkles,
  CheckCircle2,
  Image as ImageIcon,
  Loader2,
  Plus,
  Trash2,
  Edit2,
  Star,
  X,
  Save,
  Package,
  Upload,
  Camera,
  Eye,
  Boxes,
} from "lucide-react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { adminService } from "@/modules/admin/api/adminService";
import { ProductBrandFormatFields } from "@/modules/admin/components/ProductBrandFormatFields";
import { ProductCoreFields } from "@/modules/admin/components/ProductCoreFields";
import { ProductOptionAxesFields } from "@/modules/admin/components/ProductOptionAxesFields";
import { VariantAxisInputsFields } from "@/modules/admin/components/VariantAxisInputsFields";
import { VariantAxisValuesFields } from "@/modules/admin/components/VariantAxisValuesFields";
import { VariantBasicsFields } from "@/modules/admin/components/VariantBasicsFields";
import { SimplePricingFields } from "@/modules/admin/components/SimplePricingFields";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { normalizeMediaUrl } from "@/shared/lib/media";

interface VariantFormState {
  mode: "create" | "edit";
  id: string | null;
  title: string;
  sku: string;
  price: string;
  originalPrice: string;
  currency: string;
  stock: string;
  currentStock: number;
  adjustmentReason: string;
  isDefault: boolean;
  attributes: Record<string, string>;
}

function slugifyProduct(value: string): string {
  const slug = value
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^\w\s-]/g, "")
    .trim()
    .replace(/[\s_-]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return slug || `product-${Date.now()}`;
}

function getVariantThumbnailUrl(v: any): string | null {
  if (!v?.thumbnail) return null;
  const raw = typeof v.thumbnail === "string" ? v.thumbnail : (v.thumbnail.url || v.thumbnail.publicUrl || null);
  return raw ? normalizeMediaUrl(raw) : null;
}

function getProductThumbnailUrl(p: any): string | null {
  if (!p?.thumbnail) return null;
  const raw = typeof p.thumbnail === "string" ? p.thumbnail : (p.thumbnail.url || p.thumbnail.publicUrl || null);
  return raw ? normalizeMediaUrl(raw) : null;
}

interface AdminProductEditorProps {
  mode?: "create" | "edit";
}

export function AdminProductEditor({ mode: initialMode }: AdminProductEditorProps) {
  const t = useTranslations("admin");
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const rawId = params?.id ? String(params.id) : null;
  const draftParam = searchParams?.get("draft") ?? null;
  const initialProductId = rawId && rawId !== "new" ? rawId : draftParam;
  const stepParam = Number(searchParams?.get("step") ?? 1);
  const initialWizardStep = (stepParam >= 1 && stepParam <= 5 ? stepParam : 1) as
    | 1
    | 2
    | 3
    | 4
    | 5;

  const [productId, setProductId] = useState<string | null>(initialProductId);
  const isCreate = !productId;
  // Creation keeps the staged wizard even after the draft row is persisted.
  // Editing an existing product uses a section/tab layout instead.
  const isEditMode = initialMode === "edit" && Boolean(initialProductId);


  const [wizardStep, setWizardStep] = useState<1 | 2 | 3 | 4 | 5>(initialWizardStep);
  const [editTab, setEditTab] = useState<"details" | "variants" | "publishing">("details");
  const [categories, setCategories] = useState<any[]>([]);
  const [brands, setBrands] = useState<any[]>([]);
  const [attributes, setAttributes] = useState<any[]>([]);
  const [catalogProducts, setCatalogProducts] = useState<any[]>([]);
  const [productDetail, setProductDetail] = useState<any>(null);

  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [isSavingStep, setIsSavingStep] = useState(false);
  const [savingVariant, setSavingVariant] = useState(false);
  const [matrixSaving, setMatrixSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [archiving, setArchiving] = useState(false);
  const [creatingAxis, setCreatingAxis] = useState(false);

  // Variant editing & media management state
  const [variantForm, setVariantForm] = useState<VariantFormState | null>(null);
  const [selectedVariantForMedia, setSelectedVariantForMedia] = useState<any>(null);
  const [uploadingProductThumbnail, setUploadingProductThumbnail] = useState(false);
  const [uploadingVariantThumb, setUploadingVariantThumb] = useState(false);
  const [uploadingVariantMedia, setUploadingVariantMedia] = useState(false);

  const productFileInputRef = useRef<HTMLInputElement>(null);
  const variantThumbInputRef = useRef<HTMLInputElement>(null);
  const variantMediaInputRef = useRef<HTMLInputElement>(null);

  const [axisDraftValues, setAxisDraftValues] = useState<Record<string, string[]>>({});
  const [matrixStock, setMatrixStock] = useState("");

  // Quick variant inventory adjustment modal state
  const [quickStockVariant, setQuickStockVariant] = useState<any | null>(null);
  const [quickStockAmount, setQuickStockAmount] = useState<string>("");
  const [quickStockReason, setQuickStockReason] = useState<string>("restock");
  const [savingQuickStock, setSavingQuickStock] = useState(false);

  // Product form data
  const [formData, setFormData] = useState({
    name: "",
    sku: `SKU-${Math.floor(1000 + Math.random() * 9000)}`,
    categoryId: "",
    category: "",
    brandId: "",
    brand: "",
    description: "",
    productType: "simple" as "simple" | "variable",
    variationAxes: [] as string[],
    price: "",
    compareAtPrice: "",
    cost: "",
    stock: "",
    adjustmentReason: "restock",
    status: "draft",
    image: "",
  });

  const applyDetailToForm = useCallback((detail: any) => {
    if (!detail) return;
    const variants = Array.isArray(detail.variants) ? detail.variants : [];
    const activeVariant = variants.find((variant: any) => variant.isDefault) || variants[0];
    const availableStock = variants.reduce(
      (total: number, variant: any) => total + Number(variant.stock?.availableQuantity ?? 0),
      0
    );
    const price = Number(activeVariant?.extractedPrice ?? 0);

    setFormData((prev) => ({
      ...prev,
      name: detail.title ?? prev.name,
      categoryId: detail.categoryId ?? prev.categoryId,
      category: detail.category?.name ?? prev.category,
      brandId: detail.brandId ?? prev.brandId ?? "",
      brand: detail.brand?.name ?? prev.brand,
      description: detail.description ?? "",
      productType: (detail.productType === "variable" || (detail.variationAxes?.length ?? 0) > 0) ? "variable" : "simple",
      variationAxes: Array.isArray(detail.variationAxes) ? detail.variationAxes : [],
      sku: activeVariant?.sku ?? prev.sku,
      price: variants.length > 0 && Number.isFinite(price) ? String(price) : prev.price,
      compareAtPrice: activeVariant?.extractedOriginalPrice != null
        ? String(activeVariant.extractedOriginalPrice)
        : (detail.extractedOriginalPrice != null ? String(detail.extractedOriginalPrice) : ""),
      stock: variants.length > 0 ? String(availableStock) : prev.stock,
      image: detail.thumbnail?.url ?? detail.thumbnail?.publicUrl ?? prev.image,
      status: detail.status || "draft",
    }));
  }, []);

  const loadProductDetail = useCallback(async (idToLoad: string) => {
    try {
      setDetailLoading(true);
      const res: any = await catalogService.getAdminProduct(idToLoad);
      const detail = res?.data ?? res ?? null;
      setProductDetail(detail);
      if (detail) {
        applyDetailToForm(detail);
        setSelectedVariantForMedia((prev: any) => {
          if (!prev) return null;
          return detail.variants?.find((v: any) => v.id === prev.id) || prev;
        });
      }
      return detail;
    } catch {
      setProductDetail(null);
      notify.error(t("products.toast.variantLoadFailed"));
      return null;
    } finally {
      setDetailLoading(false);
    }
  }, [applyDetailToForm, t]);

  // Initial load of reference metadata and product (if editing)
  const initData = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);

      const [catsRes, brandsRes, attrsRes, prodsRes]: [any, any, any, any] = await Promise.all([
        catalogService.getAdminCategories(100),
        catalogService.getBrands(100),
        adminService.getAttributes(),
        catalogService.getProducts("pageSize=100"),
      ]);

      const catList = (Array.isArray(catsRes) ? catsRes : (Array.isArray(catsRes?.data) ? catsRes.data : []))
        .filter((category: any) => !category.deletedAt);
      setCategories(catList);

      const brandList = Array.isArray(brandsRes) ? brandsRes : (Array.isArray(brandsRes?.data) ? brandsRes.data : []);
      setBrands(brandList);

      const attrList = Array.isArray(attrsRes) ? attrsRes : (Array.isArray(attrsRes?.data) ? attrsRes.data : []);
      setAttributes(attrList);

      const prodList = Array.isArray(prodsRes) ? prodsRes : (Array.isArray(prodsRes?.data) ? prodsRes.data : []);
      setCatalogProducts(prodList);

      if (productId) {
        await loadProductDetail(productId);
      } else if (catList.length > 0) {
        setFormData((prev) => ({
          ...prev,
          categoryId: prev.categoryId || catList[0].id,
          category: prev.category || catList[0].name,
        }));
      }
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [productId, loadProductDetail]);

  useEffect(() => {
    initData();
  }, [initData]);

  // Keep the creation draft and current stage in the URL so a refresh (or a
  // server round-trip) resumes the same staged draft instead of starting over.
  const persistCreationLocation = useCallback((id: string, step: number) => {
    if (typeof window === "undefined") return;
    const url = new URL(window.location.href);
    url.searchParams.set("draft", id);
    url.searchParams.set("step", String(step));
    window.history.replaceState(null, "", url.toString());
  }, []);

  useEffect(() => {
    if (isEditMode || !productId) return;
    persistCreationLocation(productId, wizardStep);
  }, [isEditMode, productId, wizardStep, persistCreationLocation]);

  // Simple products have no options stage: if the saved product format is
  // simple, the system moves the user directly to the versions stage.
  useEffect(() => {
    if (isEditMode || wizardStep !== 3 || !productDetail) return;
    if (productDetail.productType !== "variable") setWizardStep(4);
  }, [isEditMode, productDetail, wizardStep]);

  // Category attributes & axes computation
  const hasChildren = (productDetail?.variants?.length ?? 0) > 0;
  const formAxesLocked = Boolean(productId) && hasChildren;

  const currentCategoryName = useMemo(() => {
    const cat = categories.find((c: any) => c.id === formData.categoryId);
    return cat?.name || formData.category || t("products.form.category");
  }, [categories, formData.categoryId, formData.category, t]);

  const categoryBrands = useMemo(() => {
    if (!formData.categoryId) return brands;

    const relatedCategoryIds = new Set<string>([formData.categoryId]);
    for (const c of categories) {
      if (c.parentId === formData.categoryId) relatedCategoryIds.add(c.id);
    }
    const currentCat = categories.find((c: any) => c.id === formData.categoryId);
    if (currentCat?.parentId) relatedCategoryIds.add(currentCat.parentId);

    const currentCatName = currentCat?.name ? String(currentCat.name).toLowerCase() : "";

    const associatedIds = new Set<string>();
    for (const p of catalogProducts) {
      const pCatId = p.categoryId || (typeof p.category === "object" ? p.category?.id : undefined);
      const matchId = Boolean(pCatId && relatedCategoryIds.has(pCatId));

      let pCatName = "";
      if (typeof p.category === "string") {
        pCatName = p.category;
      } else if (p.category && typeof p.category === "object") {
        pCatName = p.category.name || p.category.en || "";
      } else if (Array.isArray(p.categories) && p.categories.length > 0) {
        pCatName = typeof p.categories[0] === "string" ? p.categories[0] : (p.categories[0]?.name || "");
      }
      const matchName = Boolean(currentCatName && pCatName && pCatName.toLowerCase() === currentCatName);

      if (matchId || matchName) {
        const pBrandId = p.brandId || (typeof p.brand === "object" ? p.brand?.id : undefined);
        if (pBrandId) associatedIds.add(pBrandId);

        let pBrandName = "";
        if (typeof p.brand === "string") {
          pBrandName = p.brand;
        } else if (p.brand && typeof p.brand === "object") {
          pBrandName = p.brand.name || p.brand.en || "";
        }

        if (pBrandName) {
          const bObj = brands.find((b: any) => {
            const bName = typeof b.name === "string" ? b.name : (b.name?.en || "");
            return bName && bName.toLowerCase() === pBrandName.toLowerCase();
          });
          if (bObj?.id) associatedIds.add(bObj.id);
        }
      }
    }

    const related = brands.filter((b: any) => associatedIds.has(b.id));

    if (formData.brandId && !associatedIds.has(formData.brandId)) {
      const existing = brands.find((b: any) => b.id === formData.brandId);
      if (existing) related.unshift(existing);
    }

    return related.length > 0 ? related : brands;
  }, [brands, catalogProducts, formData.categoryId, formData.brandId, categories]);

  const formCategoryAttributes = useMemo(
    () => attributes.filter((attribute: any) => !attribute.categoryId || attribute.categoryId === formData.categoryId),
    [attributes, formData.categoryId]
  );

  const activeAxes: string[] = useMemo(
    () => (Array.isArray(productDetail?.variationAxes) ? productDetail.variationAxes : formData.variationAxes),
    [productDetail, formData.variationAxes]
  );

  const axisValuesReady = useMemo(
    () => activeAxes.length > 0 && activeAxes.every((key) => (axisDraftValues[key]?.length ?? 0) > 0),
    [activeAxes, axisDraftValues]
  );

  const matrixCombinationCount = useMemo(
    () =>
      activeAxes.length === 0 ||
      activeAxes.some((key) => (axisDraftValues[key]?.length ?? 0) === 0)
        ? 0
        : activeAxes.reduce(
            (total, key) => total * Math.max(1, axisDraftValues[key]?.length ?? 0),
            1
          ),
    [activeAxes, axisDraftValues]
  );

  const attributeByKey = useMemo(() => {
    const map: Record<string, any> = {};
    for (const attribute of attributes) map[attribute.key] = attribute;
    return map;
  }, [attributes]);

  const variantFingerprint = useMemo(
    () =>
      (productDetail?.variants || [])
        .map((v: any) => `${v.id}:${JSON.stringify(v.attributes || {})}`)
        .join("|"),
    [productDetail]
  );
  const axisDraftProductRef = useRef<string | null>(null);
  const axisDraftReadyRef = useRef(false);

  useEffect(() => {
    axisDraftReadyRef.current = false;
  }, [productDetail?.id]);

  useEffect(() => {
    if (!productDetail) {
      axisDraftProductRef.current = null;
      setAxisDraftValues({});
      return;
    }
    const next: Record<string, string[]> = {};
    for (const key of productDetail.variationAxes || []) {
      const values: string[] = [];
      for (const variant of productDetail.variants || []) {
        const value = variant.attributes?.[key];
        if (value === undefined || value === null || value === "") continue;
        const str = String(value);
        if (!values.includes(str)) values.push(str);
      }
      next[key] = values;
    }
    const sameProduct = axisDraftProductRef.current === productDetail.id;
    axisDraftProductRef.current = productDetail.id;
    setAxisDraftValues((prev) => {
      if (!sameProduct) {
        // Restore choices the admin entered before creating versions. Choices
        // already saved as versions remain the source of truth.
        let stored: Record<string, string[]> = {};
        try {
          if (!isEditMode) {
            const raw = window.localStorage.getItem(`prim:axis-draft:${productDetail.id}`);
            if (raw) {
              const parsed = JSON.parse(raw);
              if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
                stored = parsed;
              }
            }
          }
        } catch {
          stored = {};
        }
        const restored: Record<string, string[]> = {};
        for (const key of Object.keys(next)) {
          restored[key] = [...next[key]];
          for (const value of stored[key] || []) {
            if (!restored[key].includes(value)) restored[key].push(value);
          }
        }
        return restored;
      }
      const merged: Record<string, string[]> = {};
      for (const key of Object.keys(next)) {
        const combined = [...next[key]];
        for (const value of prev[key] || []) {
          if (!combined.includes(value)) combined.push(value);
        }
        merged[key] = combined;
      }
      return merged;
    });
  }, [productDetail?.id, productDetail?.updatedAt, variantFingerprint, isEditMode]);

  // Keep unsaved option choices across refreshes while the admin is still in
  // the versions stage. Once versions exist, the server is authoritative.
  useEffect(() => {
    if (isEditMode || !productId || !productDetail) return;
    if (!axisDraftReadyRef.current) {
      axisDraftReadyRef.current = true;
      return;
    }
    try {
      window.localStorage.setItem(`prim:axis-draft:${productId}`, JSON.stringify(axisDraftValues));
    } catch {
      // Storage may be unavailable; the editor still works.
    }
  }, [axisDraftValues, isEditMode, productDetail, productId]);

  // Stage 1 -> Stage 2: Create draft on server if new, or update existing
  const handleStep1Next = async () => {
    if (!formData.name.trim() || !formData.categoryId) {
      notify.error(t("products.toast.requiredFields"));
      return;
    }

    if (!productId) {
      setIsSavingStep(true);
      try {
        const createdResponse: any = await catalogService.createProduct({
          title: formData.name.trim(),
          slug: slugifyProduct(formData.name),
          categoryId: formData.categoryId,
          ...(formData.brandId ? { brandId: formData.brandId } : {}),
          description: formData.description.trim(),
          productType: formData.productType,
          variationAxes: [],
        });
        const created = createdResponse?.data ?? createdResponse;
        if (!created?.id) throw new Error("Product creation did not return an ID");

        setProductId(created.id);
        await loadProductDetail(created.id);
        void revalidateCatalog().catch(() => false);
        // Keep the creation flow mounted on /new. Navigating to the edit route
        // here would remount the editor in edit mode and reset the wizard.
        notify.success(t("products.saveDraftSuccess"));
        setWizardStep(2);
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
      } finally {
        setIsSavingStep(false);
      }
    } else {
      setIsSavingStep(true);
      try {
        await catalogService.updateProduct(productId, {
          title: formData.name.trim(),
          categoryId: formData.categoryId,
          description: formData.description.trim(),
        });
        await loadProductDetail(productId);
        void revalidateCatalog().catch(() => false);
        setWizardStep(2);
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
      } finally {
        setIsSavingStep(false);
      }
    }
  };

  // Stage 2 -> Stage 3 (variable) or Stage 4 (simple): save brand & format
  const handleStep2Next = async () => {
    if (!productId) {
      setWizardStep(formData.productType === "variable" ? 3 : 4);
      return;
    }

    setIsSavingStep(true);
    try {
      const payload: any = {
        title: formData.name.trim(),
        categoryId: formData.categoryId,
        description: formData.description.trim(),
      };
      if (formData.brandId) payload.brandId = formData.brandId;
      if (!formAxesLocked) {
        payload.productType = formData.productType;
        // A simple product has no option dimensions. Switching back to simple
        // must clear any previously selected dimensions before the next stage.
        if (formData.productType === "simple") payload.variationAxes = [];
      }
      await catalogService.updateProduct(productId, payload);
      await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      // The system chooses the next stage from the saved product format.
      setWizardStep(formData.productType === "variable" ? 3 : 4);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    } finally {
      setIsSavingStep(false);
    }
  };

  // Stage 3 -> Stage 4: save option dimensions only. Their choices/versions
  // belong in the next stage, because they depend on this selection.
  const handleOptionsNext = async () => {
    if (formData.productType !== "variable") {
      setWizardStep(4);
      return;
    }
    if (formData.variationAxes.length === 0) {
      notify.error(t("products.form.axesRequired"));
      return;
    }
    if (!productId) {
      setWizardStep(4);
      return;
    }

    setIsSavingStep(true);
    try {
      const payload: any = {
        title: formData.name.trim(),
        categoryId: formData.categoryId,
        description: formData.description.trim(),
        productType: "variable",
        variationAxes: formData.variationAxes,
      };
      if (formData.brandId) payload.brandId = formData.brandId;
      await catalogService.updateProduct(productId, payload);
      await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      setWizardStep(4);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    } finally {
      setIsSavingStep(false);
    }
  };

  // Persist the single-variant fields used by simple products. This is shared
  // by "Next", "Save", and "Publish" so changing price/original price/stock in
  // any of those actions is not lost when the detail is reloaded.
  const saveSimpleVariantPricing = async (): Promise<boolean> => {
    if (!productId) return false;

    const parsedPrice = parseFloat(formData.price);
    if (Number.isNaN(parsedPrice) || parsedPrice < 0 || formData.stock.trim() === "") {
      notify.error(t("products.toast.requiredFields"));
      return false;
    }

    const compareRaw = formData.compareAtPrice.trim();
    const parsedCompare = compareRaw ? parseFloat(compareRaw) : null;
    if (parsedCompare != null && (Number.isNaN(parsedCompare) || parsedCompare < 0)) {
      notify.error(t("products.toast.invalidOriginalPrice"));
      return false;
    }

    // The API treats <= 0 as "clear the original price", so an empty field
    // must send 0 rather than being omitted.
    const crossedOutPrice = parsedCompare != null ? Math.round(parsedCompare * 100) : 0;
    const simpleVariant = (productDetail?.variants || [])[0];

    if (simpleVariant?.id) {
      await catalogService.updateVariant(simpleVariant.id, {
        title: formData.name.trim() || undefined,
        sku: formData.sku.trim() || undefined,
        price: Math.round(parsedPrice * 100),
        crossedOutPrice,
      });

      const newStock = parseInt(formData.stock, 10);
      if (!Number.isNaN(newStock)) {
        const oldStock = Number(simpleVariant.stock?.availableQuantity ?? 0);
        const stockDiff = newStock - oldStock;
        if (stockDiff !== 0) {
          const reason = formData.adjustmentReason || (stockDiff > 0 ? "restock" : "adjustment");
          await adminService.adjustInventory(simpleVariant.id, stockDiff, reason);
        }
      }
    } else {
      await catalogService.createVariant(productId, {
        title: formData.name.trim(),
        sku: formData.sku.trim() || undefined,
        price: Math.round(parsedPrice * 100),
        crossedOutPrice: crossedOutPrice > 0 ? crossedOutPrice : undefined,
        currency: "EGP",
        initialStock: parseInt(formData.stock, 10) || 0,
        isDefault: true,
      });
    }

    return true;
  };

  // Stage 4 -> Stage 5: Save pricing & single/variable version details
  const handleVersionsNext = async () => {
    if (formData.productType === "simple") {
      if (!productId) {
        setWizardStep(5);
        return;
      }

      setIsSavingStep(true);
      try {
        if (!(await saveSimpleVariantPricing())) return;
        await loadProductDetail(productId);
        void revalidateCatalog().catch(() => false);
        setWizardStep(5);
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
      } finally {
        setIsSavingStep(false);
      }
    } else {
      if ((productDetail?.variants?.length ?? 0) === 0) {
        notify.error(t("products.wizard.noVariantsWarning"));
        return;
      }
      setWizardStep(5);
    }
  };

  // Quick save draft at any point
  const handleSaveDraft = async () => {
    if (!productId) {
      if (!formData.name.trim() || !formData.categoryId) {
        notify.error(t("products.toast.requiredFields"));
        return;
      }
      handleStep1Next();
      return;
    }

    setIsSavingStep(true);
    try {
      const payload: any = {
        title: formData.name.trim(),
        description: formData.description.trim(),
        categoryId: formData.categoryId,
        ...(formData.brandId ? { brandId: formData.brandId } : {}),
      };
      if (!formAxesLocked) {
        payload.productType = formData.productType;
        payload.variationAxes = formData.productType === "variable" ? formData.variationAxes : [];
      }
      await catalogService.updateProduct(productId, payload);
      if (formData.productType === "simple" && (isEditMode || wizardStep >= 4)) {
        if (!(await saveSimpleVariantPricing())) return;
      }
      await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.saveDraftSuccess"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    } finally {
      setIsSavingStep(false);
    }
  };

  // Publish product
  const handlePublishProduct = async () => {
    if (!productId) return;
    if ((productDetail?.variants?.length ?? 0) === 0) {
      notify.error(t("products.wizard.noVariantsWarning"));
      return;
    }

    setPublishing(true);
    try {
      const payload: any = {
        title: formData.name.trim(),
        description: formData.description.trim(),
        categoryId: formData.categoryId,
        ...(formData.brandId ? { brandId: formData.brandId } : {}),
      };
      if (!formAxesLocked) {
        payload.productType = formData.productType;
        payload.variationAxes = formData.productType === "variable" ? formData.variationAxes : [];
      }
      await catalogService.updateProduct(productId, payload);
      if (formData.productType === "simple" && !(await saveSimpleVariantPricing())) return;

      await catalogService.publishProduct(productId);
      notify.success(t("products.toast.published"));
      void revalidateCatalog().catch(() => false);
      if (isEditMode) {
        await loadProductDetail(productId);
      } else {
        router.push("/admin/products");
      }
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.publishFailed")));
    } finally {
      setPublishing(false);
    }
  };

  // Archive an existing product without leaving the editor.
  const handleArchiveProduct = async () => {
    if (!productId) return;
    if (!window.confirm(t("products.editMode.archiveConfirm"))) return;

    setArchiving(true);
    try {
      await catalogService.archiveProduct(productId);
      notify.success(t("products.editMode.archiveSuccess"));
      await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.editMode.archiveFailed")));
    } finally {
      setArchiving(false);
    }
  };

  // Inline Option / Axis creation
  const handleCreateInlineAxis = async (customLabel?: string) => {
    const label = (customLabel || "").trim();
    if (!label || !formData.categoryId || formAxesLocked) return;

    const key =
      label
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "_")
        .replace(/^_+|_+$/g, "") || `axis_${Date.now()}`;

    if (formData.variationAxes.includes(key)) {
      notify.info(t("products.wizard.axisAlreadyExists", { name: label }));
      return;
    }

    const existingAttr = attributes.find(
      (a: any) =>
        (!a.categoryId || a.categoryId === formData.categoryId) &&
        (a.key === key || (typeof a.label === "string" && a.label.toLowerCase() === label.toLowerCase()))
    );

    if (existingAttr) {
      setFormData((prev) => ({
        ...prev,
        variationAxes: [...prev.variationAxes, existingAttr.key],
      }));
      notify.success(t("products.wizard.axisCreated", { name: existingAttr.label }));
      return;
    }

    setCreatingAxis(true);
    try {
      const res: any = await adminService.createAttribute({
        categoryId: formData.categoryId,
        key,
        label,
        valueType: "text",
        allowedValues: [],
        isFilterable: true,
        sortOrder: 0,
      });
      const created = res?.data ?? res;
      const newAttr = {
        id: created?.id || key,
        categoryId: formData.categoryId,
        key,
        label,
        valueType: "text",
        allowedValues: [],
        isFilterable: true,
        sortOrder: 0,
      };
      setAttributes((prev) => [...prev, newAttr]);
      setFormData((prev) => ({
        ...prev,
        variationAxes: [...prev.variationAxes, key],
      }));
      notify.success(t("products.wizard.axisCreated", { name: label }));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.wizard.axisCreateFailed")));
    } finally {
      setCreatingAxis(false);
    }
  };

  const handleCreateMissingCombinations = async () => {
    if (!productDetail || activeAxes.length === 0) return;

    const lists: string[][] = [];
    for (const key of activeAxes) {
      const values = axisDraftValues[key] || [];
      if (values.length === 0) {
        notify.error(t("products.variants.axisValuesRequired"));
        return;
      }
      lists.push(values);
    }

    const cartesian = lists.reduce<string[][]>(
      (acc, list) => acc.flatMap((prefix) => list.map((value) => [...prefix, value])),
      [[]]
    );

    const used = new Set<string>();
    for (const variant of productDetail.variants || []) {
      const values = activeAxes.map((key) => {
        const value = variant.attributes?.[key];
        return value === undefined || value === null ? "" : String(value);
      });
      if (values.every(Boolean)) used.add(values.join("\u0001"));
    }

    const missing = cartesian.filter((values) => !used.has(values.join("\u0001")));
    if (missing.length === 0) {
      notify.info(t("products.variants.matrixNothingMissing"));
      return;
    }

    const fallbackPrice = parseFloat(formData.price);
    if (Number.isNaN(fallbackPrice) || fallbackPrice < 0) {
      notify.error(t("products.toast.requiredFields"));
      return;
    }
    const defaultStock = Math.max(0, parseInt(matrixStock, 10) || 0);

    setMatrixSaving(true);
    try {
      await catalogService.createVariantsBatch(productDetail.id, {
        variants: missing.map((values) => {
          const variantAttributes: Record<string, string> = {};
          activeAxes.forEach((key, index) => {
            variantAttributes[key] = values[index];
          });
          return {
            attributes: variantAttributes,
            title: `${productDetail.title} - ${values.join(", ")}`,
            price: Math.round(fallbackPrice * 100),
            crossedOutPrice: formData.compareAtPrice && !isNaN(parseFloat(formData.compareAtPrice))
              ? Math.round(parseFloat(formData.compareAtPrice) * 100)
              : undefined,
            currency: "EGP",
            initialStock: defaultStock,
            isDefault: false,
          };
        }),
      });
      await loadProductDetail(productDetail.id);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.variants.matrixCreated", { count: missing.length }));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.matrixSaveFailed")));
    } finally {
      setMatrixSaving(false);
    }
  };

  const openVariantCreate = () => {
    const initialAttrs: Record<string, string> = {};
    for (const key of activeAxes) {
      const draftVals = axisDraftValues[key] || [];
      initialAttrs[key] = draftVals[0] || "";
    }

    const parts = activeAxes.map((k) => initialAttrs[k]).filter(Boolean);
    const suggestedTitle =
      parts.length > 0
        ? `${formData.name.trim() || productDetail?.title || "Variant"} - ${parts.join(", ")}`
        : "";

    setVariantForm({
      mode: "create",
      id: null,
      title: suggestedTitle,
      sku: "",
      price: formData.price || "",
      originalPrice: formData.compareAtPrice || "",
      currency: "EGP",
      stock: "",
      currentStock: 0,
      adjustmentReason: "restock",
      isDefault: (productDetail?.variants?.length ?? 0) === 0,
      attributes: initialAttrs,
    });
  };

  const openVariantEdit = (variant: any) => {
    const existingAttrs = variant.attributes || {};
    const initialAttrs: Record<string, string> = {};
    for (const key of activeAxes) {
      initialAttrs[key] =
        existingAttrs[key] !== undefined && existingAttrs[key] !== null ? String(existingAttrs[key]) : "";
    }

    const currentStock = Number(variant.stock?.availableQuantity ?? 0);
    setVariantForm({
      mode: "edit",
      id: variant.id,
      title: variant.title || "",
      sku: variant.sku || "",
      price: variant.extractedPrice != null ? String(variant.extractedPrice) : "",
      originalPrice:
        variant.extractedOriginalPrice != null
          ? String(variant.extractedOriginalPrice)
          : (variant.originalPrice ? String(variant.originalPrice).replace(/[^0-9.-]+/g, "") : ""),
      currency: variant.currency || "EGP",
      stock: String(currentStock),
      currentStock: currentStock,
      adjustmentReason: "restock",
      isDefault: Boolean(variant.isDefault),
      attributes: initialAttrs,
    });
  };

  const handleSaveVariant = async () => {
    if (!variantForm || !productDetail) return;

    const price = parseFloat(variantForm.price);
    if (
      !variantForm.title.trim() ||
      Number.isNaN(price) ||
      price < 0
    ) {
      notify.error(t("products.variants.requiredFields"));
      return;
    }

    const stock = Number(variantForm.stock);
    if (!Number.isInteger(stock) || stock < 0) {
      notify.error(t("products.variants.validation.stockInvalid"));
      return;
    }

    if (activeAxes.length > 0) {
      const missingAxis = activeAxes.find(
        (key) => !variantForm.attributes?.[key] || String(variantForm.attributes[key]).trim() === ""
      );
      if (missingAxis) {
        notify.error(t("products.variants.axisValuesRequired"));
        return;
      }
    }

    setSavingVariant(true);
    try {
      if (variantForm.mode === "create") {
        await catalogService.createVariant(productDetail.id, {
          title: variantForm.title.trim(),
          price: Math.round(price * 100),
          crossedOutPrice: variantForm.originalPrice && !isNaN(parseFloat(variantForm.originalPrice))
            ? Math.round(parseFloat(variantForm.originalPrice) * 100)
            : undefined,
          currency: variantForm.currency || "EGP",
          attributes: variantForm.attributes,
          initialStock: variantForm.stock ? parseInt(variantForm.stock, 10) : 0,
          isDefault: Boolean(variantForm.isDefault),
        });
        notify.success(t("products.variants.created"));
      } else if (variantForm.id) {
        await catalogService.updateVariant(variantForm.id, {
          title: variantForm.title.trim(),
          sku: variantForm.sku || undefined,
          price: Math.round(price * 100),
          crossedOutPrice: variantForm.originalPrice && !isNaN(parseFloat(variantForm.originalPrice))
            ? Math.round(parseFloat(variantForm.originalPrice) * 100)
            : 0,
          currency: variantForm.currency || "EGP",
          attributes: variantForm.attributes,
        });

        // Update inventory per variant if stock changed
        const newStock = parseInt(variantForm.stock, 10);
        if (!isNaN(newStock)) {
          const oldStock = Number(variantForm.currentStock ?? 0);
          const stockDiff = newStock - oldStock;
          if (stockDiff !== 0) {
            const reason = variantForm.adjustmentReason || (stockDiff > 0 ? "restock" : "adjustment");
            await adminService.adjustInventory(variantForm.id, stockDiff, reason);
          }
        }

        notify.success(t("products.variants.updated"));
      }

      setVariantForm(null);
      await loadProductDetail(productDetail.id);
      void revalidateCatalog().catch(() => false);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.saveFailed")));
    } finally {
      setSavingVariant(false);
    }
  };

  const handleApplyQuickStock = async () => {
    if (!quickStockVariant) return;
    const newStock = parseInt(quickStockAmount, 10);
    if (isNaN(newStock) || newStock < 0) {
      notify.error(t("products.variants.requiredFields"));
      return;
    }
    const current = Number(quickStockVariant.stock?.availableQuantity ?? 0);
    const diff = newStock - current;
    if (diff === 0) {
      setQuickStockVariant(null);
      return;
    }

    setSavingQuickStock(true);
    try {
      await adminService.adjustInventory(quickStockVariant.id, diff, quickStockReason);
      notify.success(t("products.variants.updated"));
      setQuickStockVariant(null);
      if (productId) await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    } finally {
      setSavingQuickStock(false);
    }
  };

  const handleDeleteVariant = async (variant: any) => {
    if (!productDetail || !confirm(t("products.variants.confirmDelete", { title: variant.title }))) return;
    try {
      await catalogService.deleteVariant(variant.id);
      notify.success(t("products.variants.deleted"));
      await loadProductDetail(productDetail.id);
      void revalidateCatalog().catch(() => false);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.deleteFailed")));
    }
  };

  const handleSetDefaultVariant = async (variant: any) => {
    if (!productDetail) return;
    try {
      await catalogService.setDefaultVariant(productDetail.id, variant.id);
      notify.success(t("products.variants.defaultSet"));
      await loadProductDetail(productDetail.id);
      void revalidateCatalog().catch(() => false);
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.defaultFailed")));
    }
  };

  // Product Thumbnail Handlers
  const handleUploadProductThumbnail = async (file: File) => {
    if (!productId) {
      notify.error(t("products.toast.requiredFields"));
      return;
    }
    setUploadingProductThumbnail(true);
    try {
      const res: any = await catalogService.uploadProductThumbnail(productId, file);
      const thumbUrl = res?.data?.url || res?.data?.publicUrl || "";
      if (thumbUrl) {
        setFormData((prev) => ({ ...prev, image: thumbUrl }));
      }
      await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.form.thumbnailSuccess"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.uploadFailed")));
    } finally {
      setUploadingProductThumbnail(false);
      if (productFileInputRef.current) productFileInputRef.current.value = "";
    }
  };

  const handleDeleteProductThumbnail = async () => {
    if (!productId) return;
    setUploadingProductThumbnail(true);
    try {
      await catalogService.deleteProductThumbnail(productId);
      setFormData((prev) => ({ ...prev, image: "" }));
      await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.form.thumbnailRemoved"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    } finally {
      setUploadingProductThumbnail(false);
    }
  };

  // Variant Thumbnail & Previews Handlers
  const handleUploadVariantThumbnail = async (variantId: string, file: File) => {
    setUploadingVariantThumb(true);
    try {
      await catalogService.uploadVariantThumbnail(variantId, file);
      if (productId) await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.variants.mediaUploadSuccess"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.uploadFailed")));
    } finally {
      setUploadingVariantThumb(false);
      if (variantThumbInputRef.current) variantThumbInputRef.current.value = "";
    }
  };

  const handleDeleteVariantThumbnail = async (variantId: string) => {
    setUploadingVariantThumb(true);
    try {
      await catalogService.deleteVariantThumbnail(variantId);
      if (productId) await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.variants.mediaDeleteSuccess"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    } finally {
      setUploadingVariantThumb(false);
    }
  };

  const handleUploadVariantMedia = async (variantId: string, files: FileList | File[]) => {
    if (!files || files.length === 0) return;
    setUploadingVariantMedia(true);
    try {
      for (const file of Array.from(files)) {
        await catalogService.uploadVariantMedia(variantId, file);
      }
      if (productId) await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.variants.mediaUploadSuccess"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.uploadFailed")));
    } finally {
      setUploadingVariantMedia(false);
      if (variantMediaInputRef.current) variantMediaInputRef.current.value = "";
    }
  };

  const handleDeleteVariantMedia = async (variantId: string, mediaId: string) => {
    try {
      await catalogService.detachVariantMedia(variantId, mediaId);
      if (productId) await loadProductDetail(productId);
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.variants.mediaDeleteSuccess"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
    }
  };

  if (loadError) {
    return <ErrorState onRetry={initData} />;
  }

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[50vh] gap-3">
        <Loader2 className="size-8 animate-spin text-accent-brand" />
        <p className="text-xs text-muted-foreground">{t("products.audit.loading")}</p>
      </div>
    );
  }

  const productStatus = productDetail?.status || "draft";
  const isPublished = productStatus === "published";
  const statusLabel = t(`status.${productStatus}`);

  // Creation-stage completeness is used both for progress affordances and for
  // guarding the stepper against skipping required configuration.
  const basicsComplete = Boolean(formData.name.trim() && formData.categoryId);
  const effectiveProductType = productDetail
    ? productDetail.productType === "variable"
      ? "variable"
      : "simple"
    : formData.productType;
  const variantsComplete = (productDetail?.variants?.length ?? 0) > 0;

  // Simple products do not have an options stage. The order is fixed by the
  // dependency chain: category -> brand/format -> options -> versions -> publish.
  const applicableStages: number[] =
    effectiveProductType === "simple" ? [1, 2, 4, 5] : [1, 2, 3, 4, 5];
  const currentStageIndex = Math.max(0, applicableStages.indexOf(wizardStep));
  const totalStages = applicableStages.length;

  const stepComplete = (step: number): boolean => {
    if (step === 1) return basicsComplete;
    if (step === 2) return basicsComplete && Boolean(productId);
    if (step === 3) return effectiveProductType === "variable" && formData.variationAxes.length > 0;
    if (step === 4) return variantsComplete;
    return isPublished;
  };

  // The editor drives the flow: the user can review previous stages, but cannot
  // jump ahead of the stage the system has brought them to.
  const canOpenStep = (step: number): boolean => {
    if (!applicableStages.includes(step)) return false;
    return step <= wizardStep;
  };

  const showStep = (step: 1 | 2 | 3 | 4 | 5): boolean => {
    if (!isEditMode) return wizardStep === step;
    if (editTab === "details") return step === 1 || step === 2 || step === 3;
    if (editTab === "variants") return step === 4;
    return step === 5;
  };

  const goBackStage = () => {
    const index = applicableStages.indexOf(wizardStep);
    if (index > 0) {
      setWizardStep(applicableStages[index - 1] as 1 | 2 | 3 | 4 | 5);
    }
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* Sticky Top Header */}
      <div className="sticky top-0 z-20 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-background/95 backdrop-blur-md p-4 sm:p-5 rounded-2xl border border-border shadow-xs">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => router.push("/admin/products")}
            className="p-2 rounded-xl border border-border hover:bg-secondary text-muted-foreground hover:text-foreground transition-all shrink-0"
            title={t("products.backToProducts")}
          >
            <ArrowLeft className="size-4 rtl:rotate-180" />
          </button>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-lg sm:text-xl font-bold text-foreground truncate max-w-md">
                {formData.name.trim() || (isCreate ? t("products.pageTitleNew") : t("products.pageTitleEdit"))}
              </h1>
              <span
                className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider shrink-0 ${
                  isPublished
                    ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/30"
                    : productStatus === "archived"
                    ? "bg-slate-500/10 text-slate-600 border border-slate-500/30"
                    : "bg-amber-500/10 text-amber-600 border border-amber-500/30"
                }`}
              >
                {statusLabel}
              </span>
            </div>
            <p className="text-xs text-muted-foreground mt-0.5">
              {productId ? `ID: ${productId.slice(0, 8)}...` : t("products.wizard.draftNotice")}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <button
            type="button"
            onClick={handleSaveDraft}
            disabled={isSavingStep}
            className="flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold rounded-xl border border-border bg-background hover:bg-secondary text-foreground transition-all disabled:opacity-50"
          >
            {isSavingStep ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
            <span>{isEditMode ? t("products.editMode.saveChanges") : t("products.wizard.saveDraft")}</span>
          </button>

          {(isEditMode || wizardStep === 5) && (
            <button
              type="button"
              onClick={handlePublishProduct}
              disabled={publishing || isSavingStep || !productId || (productDetail?.variants?.length ?? 0) === 0}
              className="flex items-center gap-1.5 px-4 py-2 text-xs font-semibold rounded-xl bg-emerald-600 text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50 transition-all"
              title={(productDetail?.variants?.length ?? 0) === 0 ? t("products.wizard.noVariantsWarning") : ""}
            >
              {publishing ? <Loader2 className="size-3.5 animate-spin" /> : <Sparkles className="size-3.5" />}
              <span>{isPublished ? t("products.editMode.republish") : t("products.publish")}</span>
            </button>
          )}
        </div>
      </div>

      {isEditMode ? (
        <div className="bg-background border border-border rounded-2xl p-2 shadow-xs">
          <div role="tablist" aria-label={t("products.editMode.tabsLabel")} className="grid grid-cols-3 gap-1">
            {[
              {
                id: "details" as const,
                label: t("products.editMode.tabDetails"),
                desc: t("products.editMode.tabDetailsDesc"),
                icon: Layers,
              },
              {
                id: "variants" as const,
                label: t("products.editMode.tabVariants"),
                desc: t("products.editMode.tabVariantsDesc"),
                icon: Boxes,
              },
              {
                id: "publishing" as const,
                label: t("products.editMode.tabPublishing"),
                desc: t("products.editMode.tabPublishingDesc"),
                icon: Sparkles,
              },
            ].map((tab) => {
              const Icon = tab.icon;
              const isActive = editTab === tab.id;
              return (
                <button
                  key={tab.id}
                  type="button"
                  role="tab"
                  aria-selected={isActive}
                  onClick={() => setEditTab(tab.id)}
                  className={`flex items-center justify-center sm:justify-start gap-2.5 rounded-xl px-3 py-2.5 text-start transition-all ${
                    isActive
                      ? "bg-accent-brand/10 text-accent-brand shadow-2xs"
                      : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
                  }`}
                >
                  <Icon className="size-4 shrink-0" />
                  <span className="min-w-0">
                    <span className="block text-xs font-bold truncate">{tab.label}</span>
                    <span className="hidden sm:block text-[11px] text-muted-foreground truncate">
                      {tab.desc}
                    </span>
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="bg-background border border-border rounded-2xl p-3 shadow-xs">
          <div className="flex items-center justify-between px-1 pb-2">
            <span className="text-[11px] font-semibold text-muted-foreground">
              {t("products.wizard.stepOf", { current: currentStageIndex + 1, total: totalStages })}
            </span>
            <span className="text-[11px] font-bold text-accent-brand">
              {currentStageIndex + 1}/{totalStages}
            </span>
          </div>
          <div className="mx-1 mb-3 h-1 overflow-hidden rounded-full bg-secondary">
            <div
              className="h-full rounded-full bg-accent-brand transition-all duration-300"
              style={{ width: `${((currentStageIndex + 1) / totalStages) * 100}%` }}
            />
          </div>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-2">
            {[
              { step: 1, title: t("products.wizard.step1"), desc: t("products.wizard.step1Desc") },
              { step: 2, title: t("products.wizard.step2"), desc: t("products.wizard.step2Desc") },
              { step: 3, title: t("products.wizard.step3"), desc: t("products.wizard.step3Desc") },
              { step: 4, title: t("products.wizard.step4"), desc: t("products.wizard.step4Desc") },
              { step: 5, title: t("products.wizard.step5"), desc: t("products.wizard.step5Desc") },
            ]
              .filter((s) => applicableStages.includes(s.step))
              .map((s) => {
              const isCurrent = wizardStep === s.step;
              const complete = stepComplete(s.step);
              const canClick = canOpenStep(s.step);
              const stageNumber = applicableStages.indexOf(s.step) + 1;
              return (
                <button
                  key={s.step}
                  type="button"
                  disabled={!canClick}
                  aria-current={isCurrent ? "step" : undefined}
                  onClick={() => {
                    if (canClick) setWizardStep(s.step as 1 | 2 | 3 | 4 | 5);
                  }}
                  className={`flex items-center gap-3 p-3 rounded-xl text-start transition-all ${
                    isCurrent
                      ? "bg-accent-brand/10 border border-accent-brand/40 text-accent-brand shadow-2xs"
                      : complete
                      ? "bg-secondary/30 text-foreground border border-border hover:bg-secondary/60"
                      : "text-muted-foreground border border-transparent"
                  } ${!canClick ? "opacity-45 cursor-not-allowed" : "cursor-pointer"}`}
                >
                  <span
                    className={`size-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0 ${
                      isCurrent
                        ? "bg-accent-brand text-white"
                        : complete
                        ? "bg-emerald-600 text-white"
                        : "bg-secondary text-muted-foreground"
                    }`}
                  >
                    {complete && !isCurrent ? <Check className="size-3.5 stroke-[3]" /> : stageNumber}
                  </span>
                  <div className="min-w-0">
                    <div className="text-xs font-bold truncate">{s.title}</div>
                    <div className="text-[11px] text-muted-foreground truncate">{s.desc}</div>
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {isEditMode && productStatus === "published" && (
        <div className="flex items-start gap-3 rounded-xl border border-emerald-500/25 bg-emerald-500/5 p-4">
          <Eye className="mt-0.5 size-4 shrink-0 text-emerald-600" />
          <p className="text-[11px] leading-relaxed text-muted-foreground">
            {t("products.editMode.liveNotice")}
          </p>
        </div>
      )}

      {/* Step Content Card */}
      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-6">
        {/* Step 1: Core Details (Title, Description, Category) */}
        {showStep(1) && (
          <div className="space-y-6 max-w-2xl">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {isEditMode ? t("products.editMode.detailsHeading") : t("products.wizard.step1")}
              </h2>
              <p className="text-xs text-muted-foreground mt-1">
                {isEditMode
                  ? t("products.editMode.detailsHint")
                  : t("products.wizard.categoryRequiredFirst")}
              </p>
            </div>

            <ProductCoreFields
              value={{
                name: formData.name,
                categoryId: formData.categoryId,
                description: formData.description,
              }}
              categories={categories}
              currentCategoryName={currentCategoryName}
              onChange={(patch) =>
                setFormData((prev) => ({
                  ...prev,
                  ...patch,
                  ...(patch.categoryId
                    ? {
                        category:
                          categories.find((c: any) => c.id === patch.categoryId)
                            ?.name || prev.category,
                      }
                    : {}),
                }))
              }
            />
          </div>
        )}

        {/* Step 2: Brand & format, then Step 3: Options (dependent on the chosen format) */}
        {(showStep(2) || showStep(3)) && (
          <div className="space-y-6">
            {!isEditMode && (
              <div>
                <h2 className="text-base font-bold text-foreground">
                  {wizardStep === 3 ? t("products.wizard.step3") : t("products.wizard.step2")}
                </h2>
                <p className="text-xs text-muted-foreground mt-1">
                  {wizardStep === 3
                    ? t("products.wizard.step3OptionsHint")
                    : t("products.wizard.step2DependencyHint")}
                </p>
              </div>
            )}

            {showStep(2) && (
              <>
            {/* Stage 1 Category Context Banner */}
            <div className="rounded-xl border border-accent-brand/20 bg-accent-brand/5 p-4 flex items-start gap-3">
              <div className="p-2 rounded-lg bg-accent-brand/10 text-accent-brand shrink-0">
                <Layers className="size-4" />
              </div>
              <div className="text-xs space-y-0.5">
                <div className="font-bold text-foreground">
                  {t("products.wizard.selectedCategoryNotice", {
                    category: currentCategoryName,
                  })}
                </div>
                <div className="text-muted-foreground text-[11px]">
                  {isEditMode
                    ? t("products.editMode.optionsDependencyHint")
                    : t("products.wizard.step2DependencyHint")}
                </div>
              </div>
            </div>

            <ProductBrandFormatFields
              value={{
                brandId: formData.brandId,
                productType: formData.productType,
              }}
              brands={categoryBrands}
              currentCategoryName={currentCategoryName}
              axesLocked={formAxesLocked}
              onChange={(patch) =>
                setFormData((prev) => ({ ...prev, ...patch }))
              }
            />
              </>
            )}

            {/* Step 3: Option dimensions (dependent on the product format chosen in Step 2) */}
            {showStep(3) && formData.productType === "variable" && (
              <ProductOptionAxesFields
                value={{ variationAxes: formData.variationAxes }}
                categoryAttributes={formCategoryAttributes}
                locked={formAxesLocked}
                creating={creatingAxis}
                onChange={(patch) => setFormData((prev) => ({ ...prev, ...patch }))}
                onCreateAxis={handleCreateInlineAxis}
              />
            )}
          </div>
        )}

        {/* Step 4: Versions, Pricing & Inventory */}
        {showStep(4) && (
          <div className="space-y-6">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {isEditMode ? t("products.editMode.variantsHeading") : t("products.wizard.step4")}
              </h2>
              <p className="text-xs text-muted-foreground mt-1">
                {isEditMode
                  ? t("products.editMode.variantsHint")
                  : formData.productType === "simple"
                  ? t("products.wizard.step4SimpleHint")
                  : t("products.wizard.step4VariableHint")}
              </p>
            </div>

            {formData.productType === "simple" ? (
              <SimplePricingFields
                value={{
                  sku: formData.sku,
                  price: formData.price,
                  compareAtPrice: formData.compareAtPrice,
                  stock: formData.stock,
                  adjustmentReason: formData.adjustmentReason,
                }}
                hasChildren={hasChildren}
                onChange={(patch) =>
                  setFormData((prev) => ({ ...prev, ...patch }))
                }
              />
            ) : (
              /* Variable Product Options Values Builder & Matrix Generator */
              <div className="space-y-6">
                {(productDetail?.variants?.length ?? 0) === 0 && (
                  <div className="rounded-xl border border-amber-500/30 bg-amber-500/5 p-4">
                    <p className="text-xs font-bold text-foreground">
                      {t("products.wizard.noVersionsYet")}
                    </p>
                    <p className="mt-0.5 text-[11px] text-muted-foreground">
                      {t("products.wizard.noVersionsYetHint")}
                    </p>
                  </div>
                )}

                <div className="rounded-2xl border border-border bg-secondary/15 p-5 space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">
                        {t("products.variants.variationAxes")}
                      </h3>
                      <p className="text-[11px] text-muted-foreground mt-0.5">
                        {t("products.wizard.step3OptionsHint")}
                      </p>
                    </div>
                    <span className="text-xs font-semibold px-2.5 py-1 rounded-lg bg-accent-brand/10 text-accent-brand">
                      {t("products.variants.optionsCount", { count: activeAxes.length })}
                    </span>
                  </div>

                  <VariantAxisValuesFields
                    axes={activeAxes}
                    values={axisDraftValues}
                    definitions={attributeByKey}
                    stock={matrixStock}
                    onAddValue={(key, value) =>
                      setAxisDraftValues((prev) => ({
						...prev,
						[key]: [...(prev[key] || []), value],
					}))
					}
                    onRemoveValue={(key, value) =>
                      setAxisDraftValues((prev) => ({
						...prev,
						[key]: (prev[key] || []).filter((v) => v !== value),
					}))
					}
                    onChangeStock={(v) => setMatrixStock(v)}
                    onCreateMissing={handleCreateMissingCombinations}
                    creating={matrixSaving}
                    ready={axisValuesReady}
                    count={matrixCombinationCount}
                  />
                </div>

                {/* Generated Variants List */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">
                      {t("products.variants.listTitle")} ({productDetail?.variants?.length ?? 0})
                    </h3>
                    <button
                      type="button"
                      onClick={openVariantCreate}
                      className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl border border-border text-xs font-semibold text-foreground hover:bg-secondary transition-all"
                    >
                      <Plus className="size-3.5" />
                      <span>{t("products.variants.add")}</span>
                    </button>
                  </div>

                  {detailLoading ? (
                    <div className="flex items-center gap-2 py-6 text-xs text-muted-foreground justify-center">
                      <Loader2 className="size-4 animate-spin text-accent-brand" />
                      <span>{t("products.variants.loading")}</span>
                    </div>
                  ) : !productDetail?.variants?.length ? (
                    <div className="p-8 text-center border border-dashed border-border rounded-2xl bg-secondary/10">
                      <Package className="size-8 mx-auto text-muted-foreground opacity-50 mb-2" />
                      <p className="text-xs text-muted-foreground font-medium">{t("products.variants.empty")}</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 gap-2.5">
                      {productDetail.variants.map((variant: any, vIdx: number) => {
                        const thumbUrl = getVariantThumbnailUrl(variant);
                        const mediaCount = Array.isArray(variant.media) ? variant.media.length : 0;

                        return (
                          <div
                            key={variant.id ? `var-item-${variant.id}` : `var-item-${vIdx}`}
                            className="rounded-xl border border-border bg-secondary/20 p-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-secondary/40 transition-colors"
                          >
                            <div className="flex items-center gap-3 min-w-0">
                              {/* Variant Thumbnail Preview / Quick Media Opener */}
                              <button
                                type="button"
                                onClick={() => setSelectedVariantForMedia(variant)}
                                className="size-12 rounded-xl border border-border bg-background overflow-hidden shrink-0 flex items-center justify-center hover:border-accent-brand/60 transition-all group relative shadow-2xs"
                                title={t("products.variants.manageMedia")}
                              >
                                {thumbUrl ? (
                                  <img
                                    src={thumbUrl}
                                    alt={variant.title}
                                    className="size-full object-cover"
                                    onError={(e) => {
                                      (e.currentTarget as HTMLImageElement).src = "/placeholder-product.png";
                                    }}
                                  />
                                ) : (
                                  <Camera className="size-5 text-muted-foreground group-hover:text-accent-brand transition-colors" />
                                )}
                                <span className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity text-white">
                                  <Upload className="size-4" />
                                </span>
                              </button>

                              <div className="min-w-0">
                                <div className="flex items-center gap-2">
                                  <span className="font-semibold text-xs text-foreground truncate">{variant.title}</span>
                                  {variant.isDefault && (
                                    <span className="rounded-full bg-accent-brand/10 text-accent-brand px-2 py-0.5 text-[10px] font-bold">
                                      {t("products.variants.defaultBadge")}
                                    </span>
                                  )}
                                </div>
                                <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground mt-1 font-mono">
                                  {variant.sku && (
                                    <>
                                      <span>SKU: {variant.sku}</span>
                                      <span>·</span>
                                    </>
                                  )}
                                  <span className="font-semibold text-foreground">EGP {variant.extractedPrice ?? 0}</span>
                                  {variant.extractedOriginalPrice != null && variant.extractedOriginalPrice > 0 && (
                                    <span className="text-muted-foreground line-through text-[10px]">
                                      EGP {variant.extractedOriginalPrice}
                                    </span>
                                  )}
                                  <span>·</span>
                                  <button
                                    type="button"
                                    onClick={() => {
                                      setQuickStockVariant(variant);
                                      setQuickStockAmount(String(variant.stock?.availableQuantity ?? 0));
                                      setQuickStockReason("restock");
                                    }}
                                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-lg border border-border bg-background hover:border-accent-brand hover:bg-accent-brand/10 text-foreground text-[10px] font-semibold transition-all cursor-pointer shadow-2xs"
                                    title={t("products.variants.adjustStock")}
                                  >
                                    <Boxes className="size-3 text-accent-brand" />
                                    <span>{t("products.units", { count: variant.stock?.availableQuantity ?? 0 })}</span>
                                  </button>
                                  {mediaCount > 0 && (
                                    <>
                                      <span>·</span>
                                      <span className="text-accent-brand font-medium">
                                        {t("products.variants.imagesCount", { count: mediaCount })}
                                      </span>
                                    </>
                                  )}
                                </div>
                              </div>
                            </div>

                            <div className="flex items-center gap-1.5 shrink-0 self-end sm:self-center">
                              {/* Images & Previews Modal Button */}
                              <button
                                type="button"
                                onClick={() => setSelectedVariantForMedia(variant)}
                                className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl border border-border bg-background hover:border-accent-brand/50 hover:bg-accent-brand/10 text-muted-foreground hover:text-accent-brand transition-all text-xs font-semibold shadow-2xs"
                                title={t("products.variants.manageMedia")}
                              >
                                <ImageIcon className="size-3.5 text-accent-brand" />
                                <span>{t("products.variants.manageMedia")}</span>
                                {mediaCount > 0 && (
                                  <span className="px-1.5 py-0.2 rounded-full text-[9px] font-bold bg-accent-brand/20 text-accent-brand">
                                    {mediaCount}
                                  </span>
                                )}
                              </button>

                              {!variant.isDefault && (
                                <button
                                  type="button"
                                  onClick={() => handleSetDefaultVariant(variant)}
                                  className="p-2 rounded-xl border border-border bg-background text-muted-foreground hover:text-accent-brand hover:border-accent-brand/50 transition-colors shadow-2xs"
                                  title={t("products.variants.setDefault")}
                                >
                                  <Star className="size-4" />
                                </button>
                              )}
                              <button
                                type="button"
                                onClick={() => openVariantEdit(variant)}
                                className="p-2 rounded-xl border border-border bg-background text-muted-foreground hover:text-accent-brand hover:border-accent-brand/50 transition-colors shadow-2xs"
                                title={t("products.variants.edit")}
                              >
                                <Edit2 className="size-4" />
                              </button>
                              <button
                                type="button"
                                onClick={() => handleDeleteVariant(variant)}
                                className="p-2 rounded-xl border border-border bg-background text-muted-foreground hover:text-destructive hover:border-destructive/40 hover:bg-destructive/10 transition-colors shadow-2xs"
                                title={t("products.variants.delete")}
                              >
                                <Trash2 className="size-4" />
                              </button>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>

                {/* Inline Variant Form modal/drawer */}
                {variantForm && (
                  <div className="space-y-4 rounded-2xl border border-accent-brand/40 bg-secondary/30 p-5 shadow-xs">
                    <div className="flex items-center justify-between border-b border-border pb-2.5">
                      <h4 className="text-xs font-bold text-foreground">
                        {variantForm.mode === "create" ? t("products.variants.newTitle") : t("products.variants.editTitle")}
                      </h4>
                      <button
                        type="button"
                        onClick={() => setVariantForm(null)}
                        className="p-1 rounded-lg text-muted-foreground hover:text-foreground"
                      >
                        <X className="size-4" />
                      </button>
                    </div>

                    <VariantBasicsFields
                      mode={variantForm.mode}
                      value={{
                        title: variantForm.title,
                        sku: variantForm.sku,
                        price: variantForm.price,
                        originalPrice: variantForm.originalPrice,
                        stock: variantForm.stock,
                        adjustmentReason: variantForm.adjustmentReason || "restock",
                      }}
                      onChange={(patch) =>
                        setVariantForm((prev) => (prev ? { ...prev, ...patch } : prev))
                      }
                    />

                    {activeAxes.length > 0 && (
                      <VariantAxisInputsFields
                        axes={activeAxes}
                        attributes={variantForm.attributes}
                        definitions={attributeByKey}
                        draftValues={axisDraftValues}
                        onAttributeChange={(key, val) => {
                          const nextAttrs = { ...variantForm.attributes, [key]: val };
                          if (val.trim() && !(axisDraftValues[key] || []).includes(val.trim())) {
                            setAxisDraftValues((prev) => ({
                              ...prev,
                              [key]: [...(prev[key] || []), val.trim()],
                            }));
                          }
                          const parts = activeAxes.map((k) => nextAttrs[k]).filter(Boolean);
                          setVariantForm((prev) =>
                            prev
                              ? {
									...prev,
									attributes: nextAttrs,
									title:
										prev.mode === "create" && parts.length > 0
											? `${formData.name.trim() || productDetail?.title || "Variant"} - ${parts.join(", ")}`
											: prev.title,
							  }
                              : prev,
                          );
                        }}
                      />
                    )}

                    <div className="flex justify-end gap-2 pt-2 border-t border-border">
                      <button
                        type="button"
                        onClick={() => setVariantForm(null)}
                        className="rounded-xl border border-border px-4 py-2 text-xs font-semibold text-foreground hover:bg-secondary"
                      >
                        {t("actions.cancel")}
                      </button>
                      <button
                        type="button"
                        onClick={handleSaveVariant}
                        disabled={savingVariant}
                        className="flex items-center gap-1.5 rounded-xl bg-accent-brand px-4 py-2 text-xs font-semibold text-white hover:bg-accent-brand/90 disabled:opacity-60 shadow-xs"
                      >
                        {savingVariant && <Loader2 className="size-3.5 animate-spin" />}
                        {t("products.variants.save")}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Step 5: Media, Summary Review & Publishing */}
        {showStep(5) && (
          <div className="space-y-6">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {isEditMode ? t("products.editMode.publishingHeading") : t("products.wizard.step5")}
              </h2>
              <p className="text-xs text-muted-foreground mt-1">
                {isEditMode
                  ? t("products.editMode.publishingHint")
                  : t("products.wizard.step5Hint")}
              </p>
            </div>

            {/* Product Master Thumbnail Uploader */}
            <div className="space-y-3 max-w-2xl bg-secondary/20 border border-border rounded-2xl p-5 shadow-xs">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-xs font-bold text-foreground uppercase tracking-wider flex items-center gap-1.5">
                    <Camera className="size-3.5 text-accent-brand" />
                    <span>{t("products.form.productThumbnail")}</span>
                  </h3>
                  <p className="text-[11px] text-muted-foreground mt-0.5">
                    {t("products.form.productThumbnailHint")}
                  </p>
                </div>
                {formData.image && (
                  <button
                    type="button"
                    onClick={handleDeleteProductThumbnail}
                    disabled={uploadingProductThumbnail}
                    className="flex items-center gap-1 text-[11px] font-semibold text-destructive hover:underline disabled:opacity-60"
                  >
                    <Trash2 className="size-3.5" />
                    <span>{t("products.form.removeThumbnail")}</span>
                  </button>
                )}
              </div>

              <div className="flex flex-col sm:flex-row items-center gap-5 pt-2">
                <div className="size-28 rounded-2xl border-2 border-border overflow-hidden bg-background shrink-0 flex items-center justify-center shadow-xs relative group">
                  {formData.image ? (
                    <img
                      src={normalizeMediaUrl(formData.image)}
                      alt={formData.name}
                      className="size-full object-cover"
                      onError={(e) => {
                        (e.currentTarget as HTMLImageElement).src = "/placeholder-product.png";
                      }}
                    />
                  ) : (
                    <ImageIcon className="size-10 text-muted-foreground opacity-40" />
                  )}
                  {uploadingProductThumbnail && (
                    <div className="absolute inset-0 bg-background/80 backdrop-blur-2xs flex flex-col items-center justify-center gap-1.5">
                      <Loader2 className="size-6 animate-spin text-accent-brand" />
                      <span className="text-[10px] font-bold text-accent-brand">{t("products.form.uploadingImage")}</span>
                    </div>
                  )}
                </div>

                <div className="flex-1 w-full space-y-3">
                  <input
                    ref={productFileInputRef}
                    type="file"
                    accept="image/png,image/jpeg,image/webp,image/avif"
                    className="hidden"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      if (file) handleUploadProductThumbnail(file);
                    }}
                  />

                  <div
                    onClick={() => productFileInputRef.current?.click()}
                    className="border-2 border-dashed border-border hover:border-accent-brand/60 rounded-xl p-4 text-center cursor-pointer transition-all bg-background/50 hover:bg-secondary/40 flex flex-col items-center justify-center gap-1.5"
                  >
                    <Upload className="size-5 text-accent-brand" />
                    <span className="text-xs font-semibold text-foreground">
                      {t("products.form.uploadThumbnail")}
                    </span>
                    <span className="text-[11px] text-muted-foreground">
                      {t("products.form.dropToUpload")}
                    </span>
                  </div>

                  <div className="flex items-center gap-2">
                    <span className="text-[10px] text-muted-foreground shrink-0">{t("products.form.imageUrl")}:</span>
                    <input
                      type="text"
                      value={formData.image}
                      onChange={(e) => setFormData({ ...formData, image: e.target.value })}
                      placeholder="https://..."
                      className="flex-1 px-3 py-1.5 bg-background border border-border rounded-lg text-xs text-foreground focus:outline-hidden focus:ring-1 focus:ring-accent-brand/40 font-mono text-[11px]"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Product Summary Review Card */}
            <div className="rounded-2xl border border-border bg-secondary/20 p-5 space-y-4 max-w-2xl">
              <div className="flex items-center justify-between border-b border-border pb-3">
                <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">{t("products.wizard.summary")}</h3>
                <span
                  className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                    isPublished
                      ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/30"
                      : productStatus === "archived"
                      ? "bg-slate-500/10 text-slate-600 border border-slate-500/30"
                      : "bg-amber-500/10 text-amber-600 border border-amber-500/30"
                  }`}
                >
                  {statusLabel}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-[11px] text-muted-foreground block">{t("products.colProduct")}</span>
                  <span className="font-bold text-foreground">{formData.name || "—"}</span>
                </div>
                <div>
                  <span className="text-[11px] text-muted-foreground block">{t("products.colCategory")}</span>
                  <span className="font-bold text-foreground">{currentCategoryName}</span>
                </div>
                <div>
                  <span className="text-[11px] text-muted-foreground block">{t("products.form.brand")}</span>
                  <span className="font-bold text-foreground">{formData.brand || "—"}</span>
                </div>
                <div>
                  <span className="text-[11px] text-muted-foreground block">{t("products.form.productType")}</span>
                  <span className="font-bold text-foreground capitalize">
                    {formData.productType === "variable"
                      ? `${t("products.variants.variableBadge")} (${productDetail?.variants?.length ?? 0} variants)`
                      : t("products.variants.simpleBadge")}
                  </span>
                </div>
                <div>
                  <span className="text-[11px] text-muted-foreground block">{t("products.colStock")}</span>
                  <span className="font-bold text-foreground">
                    {t("products.units", {
                      count:
                        (productDetail?.variants || []).reduce(
                          (acc: number, v: any) => acc + (v.stock?.availableQuantity ?? 0),
                          0
                        ) || parseInt(formData.stock, 10) || 0,
                    })}
                  </span>
                </div>
                <div>
                  <span className="text-[11px] text-muted-foreground block">{t("products.colPrice")}</span>
                  <span className="font-bold text-foreground">
                    EGP {formData.price || "0.00"}
                  </span>
                </div>
              </div>

              <div className="pt-3 border-t border-border flex items-center gap-2">
                <CheckCircle2 className="size-4 text-emerald-500 shrink-0" />
                <p className="text-[11px] text-muted-foreground">
                  {(productDetail?.variants?.length ?? 0) > 0
                    ? t("products.wizard.readyToPublish")
                    : t("products.wizard.noVariantsWarning")}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Creation Wizard Footer Controls */}
        {!isEditMode && (
        <div className="flex flex-wrap items-center justify-between gap-3 pt-6 border-t border-border">
          <div>
            {wizardStep > 1 && (
              <button
                type="button"
                onClick={goBackStage}
                className="flex items-center gap-1.5 px-4 py-2.5 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground transition-all"
              >
                <ChevronLeft className="size-4 rtl:rotate-180" />
                <span>{t("products.wizard.back")}</span>
              </button>
            )}
          </div>

          <div className="flex items-center gap-2">
            {wizardStep === 1 && (
              <button
                type="button"
                onClick={handleStep1Next}
                disabled={isSavingStep || !basicsComplete}
                title={!basicsComplete ? t("products.toast.requiredFields") : undefined}
                className="flex items-center gap-1.5 px-5 py-2.5 text-xs font-semibold rounded-xl bg-accent-brand text-white shadow-xs hover:bg-accent-brand/90 disabled:opacity-60 transition-all"
              >
                {isSavingStep && <Loader2 className="size-3.5 animate-spin" />}
                <span>{t("products.wizard.next")}</span>
                <ChevronRight className="size-4 rtl:rotate-180" />
              </button>
            )}

            {wizardStep === 2 && (
              <button
                type="button"
                onClick={handleStep2Next}
                disabled={isSavingStep || !basicsComplete || !productId}
                title={!basicsComplete ? t("products.toast.requiredFields") : undefined}
                className="flex items-center gap-1.5 px-5 py-2.5 text-xs font-semibold rounded-xl bg-accent-brand text-white shadow-xs hover:bg-accent-brand/90 disabled:opacity-60 transition-all"
              >
                {isSavingStep && <Loader2 className="size-3.5 animate-spin" />}
                <span>{t("products.wizard.next")}</span>
                <ChevronRight className="size-4 rtl:rotate-180" />
              </button>
            )}

            {wizardStep === 3 && (
              <button
                type="button"
                onClick={handleOptionsNext}
                disabled={isSavingStep || formData.variationAxes.length === 0}
                title={formData.variationAxes.length === 0 ? t("products.form.axesRequired") : undefined}
                className="flex items-center gap-1.5 px-5 py-2.5 text-xs font-semibold rounded-xl bg-accent-brand text-white shadow-xs hover:bg-accent-brand/90 disabled:opacity-60 transition-all"
              >
                {isSavingStep && <Loader2 className="size-3.5 animate-spin" />}
                <span>{t("products.wizard.next")}</span>
                <ChevronRight className="size-4 rtl:rotate-180" />
              </button>
            )}

            {wizardStep === 4 && (
              <button
                type="button"
                onClick={handleVersionsNext}
                disabled={isSavingStep || (formData.productType === "variable" && (productDetail?.variants?.length ?? 0) === 0)}
                title={formData.productType === "variable" && (productDetail?.variants?.length ?? 0) === 0 ? t("products.wizard.noVariantsWarning") : undefined}
                className="flex items-center gap-1.5 px-5 py-2.5 text-xs font-semibold rounded-xl bg-accent-brand text-white shadow-xs hover:bg-accent-brand/90 disabled:opacity-60 transition-all"
              >
                {isSavingStep && <Loader2 className="size-3.5 animate-spin" />}
                <span>{t("products.wizard.next")}</span>
                <ChevronRight className="size-4 rtl:rotate-180" />
              </button>
            )}

            {wizardStep === 5 && (
              <>
                <button
                  type="button"
                  onClick={handleSaveDraft}
                  disabled={isSavingStep}
                  className="flex items-center gap-1.5 px-4 py-2.5 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground transition-all"
                >
                  {isSavingStep && <Loader2 className="size-3.5 animate-spin" />}
                  <Save className="size-3.5" />
                  <span>{t("products.wizard.saveDraft")}</span>
                </button>

                <button
                  type="button"
                  onClick={handlePublishProduct}
                  disabled={publishing || isSavingStep || !productId || (productDetail?.variants?.length ?? 0) === 0}
                  className="flex items-center gap-1.5 px-5 py-2.5 text-xs font-semibold rounded-xl bg-emerald-600 text-white shadow-xs hover:bg-emerald-700 disabled:opacity-60 transition-all"
                >
                  {publishing && <Loader2 className="size-3.5 animate-spin" />}
                  <Sparkles className="size-3.5" />
                  <span>{t("products.wizard.publishNow")}</span>
                </button>
              </>
            )}
          </div>
        </div>
        )}

        {isEditMode && (
          <div className="flex flex-wrap items-center justify-between gap-3 pt-6 border-t border-border">
            <p className="text-[11px] text-muted-foreground">
              {t("products.editMode.footerHint")}
            </p>
            <div className="flex items-center gap-2">
              {productStatus === "published" ? (
                <button
                  type="button"
                  onClick={handleArchiveProduct}
                  disabled={archiving || isSavingStep}
                  className="flex items-center gap-1.5 px-4 py-2.5 text-xs font-semibold rounded-xl border border-red-200 text-red-600 hover:bg-red-50 disabled:opacity-60 transition-all"
                >
                  {archiving && <Loader2 className="size-3.5 animate-spin" />}
                  <Archive className="size-3.5" />
                  <span>{t("products.editMode.archive")}</span>
                </button>
              ) : null}
              <button
                type="button"
                onClick={handleSaveDraft}
                disabled={isSavingStep}
                className="flex items-center gap-1.5 px-4 py-2.5 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground transition-all"
              >
                {isSavingStep && <Loader2 className="size-3.5 animate-spin" />}
                <Save className="size-3.5" />
                <span>{t("products.editMode.saveChanges")}</span>
              </button>
              <button
                type="button"
                onClick={handlePublishProduct}
                disabled={publishing || isSavingStep || !productId || (productDetail?.variants?.length ?? 0) === 0}
                className="flex items-center gap-1.5 px-5 py-2.5 text-xs font-semibold rounded-xl bg-emerald-600 text-white shadow-xs hover:bg-emerald-700 disabled:opacity-60 transition-all"
              >
                {publishing && <Loader2 className="size-3.5 animate-spin" />}
                <Sparkles className="size-3.5" />
                <span>{isPublished ? t("products.editMode.republish") : t("products.publish")}</span>
              </button>
            </div>
          </div>
        )}
      </div>
      {selectedVariantForMedia && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4 animate-in fade-in duration-200">
          <div className="bg-background border border-border rounded-3xl max-w-2xl w-full p-6 shadow-2xl max-h-[90vh] overflow-y-auto space-y-6">
            {/* Modal Header */}
            <div className="flex items-start justify-between border-b border-border pb-4">
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-base font-bold text-foreground">
                    {t("products.variants.manageMedia")}
                  </h3>
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/30">
                    {selectedVariantForMedia.title}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground mt-1 font-mono">
                  SKU: {selectedVariantForMedia.sku || "—"} · EGP {selectedVariantForMedia.extractedPrice ?? 0}
                  {selectedVariantForMedia.extractedOriginalPrice != null && selectedVariantForMedia.extractedOriginalPrice > 0 && (
                    <span className="line-through text-muted-foreground/70 ml-1.5">
                      EGP {selectedVariantForMedia.extractedOriginalPrice}
                    </span>
                  )}
                </p>
              </div>

              <button
                type="button"
                onClick={() => setSelectedVariantForMedia(null)}
                className="p-1.5 rounded-xl border border-border hover:bg-secondary text-muted-foreground hover:text-foreground transition-all"
              >
                <X className="size-4" />
              </button>
            </div>

            {/* Section 1: Variant Specific Thumbnail */}
            <div className="space-y-3 bg-secondary/20 border border-border rounded-2xl p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-xs font-bold text-foreground uppercase tracking-wider flex items-center gap-1.5">
                    <Camera className="size-3.5 text-accent-brand" />
                    <span>{t("products.variants.thumbnail")}</span>
                  </h4>
                  <p className="text-[11px] text-muted-foreground mt-0.5">
                    {t("products.variants.thumbnailHint")}
                  </p>
                </div>

                {getVariantThumbnailUrl(selectedVariantForMedia) && (
                  <button
                    type="button"
                    onClick={() => handleDeleteVariantThumbnail(selectedVariantForMedia.id)}
                    disabled={uploadingVariantThumb}
                    className="flex items-center gap-1 text-[11px] font-semibold text-destructive hover:underline disabled:opacity-60"
                  >
                    <Trash2 className="size-3" />
                    <span>{t("products.variants.deleteThumbnail")}</span>
                  </button>
                )}
              </div>

              <div className="flex items-center gap-4 pt-1">
                <div className="size-20 rounded-xl border-2 border-border overflow-hidden bg-background shrink-0 flex items-center justify-center relative shadow-xs">
                  {getVariantThumbnailUrl(selectedVariantForMedia) ? (
                    <img
                      src={getVariantThumbnailUrl(selectedVariantForMedia)!}
                      alt={selectedVariantForMedia.title}
                      className="size-full object-cover"
                      onError={(e) => {
                        (e.currentTarget as HTMLImageElement).src = "/placeholder-product.png";
                      }}
                    />
                  ) : (
                    <ImageIcon className="size-8 text-muted-foreground opacity-30" />
                  )}
                  {uploadingVariantThumb && (
                    <div className="absolute inset-0 bg-background/80 backdrop-blur-2xs flex items-center justify-center">
                      <Loader2 className="size-5 animate-spin text-accent-brand" />
                    </div>
                  )}
                </div>

                <div className="flex-1 space-y-2">
                  <input
                    ref={variantThumbInputRef}
                    type="file"
                    accept="image/png,image/jpeg,image/webp,image/avif"
                    className="hidden"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      if (file) handleUploadVariantThumbnail(selectedVariantForMedia.id, file);
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => variantThumbInputRef.current?.click()}
                    disabled={uploadingVariantThumb}
                    className="flex items-center gap-2 px-3.5 py-2 bg-accent-brand text-white text-xs font-semibold rounded-xl hover:bg-accent-brand/90 transition-all shadow-2xs disabled:opacity-60"
                  >
                    {uploadingVariantThumb ? (
                      <Loader2 className="size-3.5 animate-spin" />
                    ) : (
                      <Upload className="size-3.5" />
                    )}
                    <span>{t("products.variants.uploadThumbnail")}</span>
                  </button>
                  <p className="text-[10px] text-muted-foreground">
                    PNG, JPG, WebP (max 10MB)
                  </p>
                </div>
              </div>
            </div>

            {/* Section 2: Variant Previews Gallery */}
            <div className="space-y-3 bg-secondary/20 border border-border rounded-2xl p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-xs font-bold text-foreground uppercase tracking-wider flex items-center gap-1.5">
                    <ImageIcon className="size-3.5 text-accent-brand" />
                    <span>{t("products.variants.previews")}</span>
                    <span className="text-muted-foreground ml-1">
                      ({selectedVariantForMedia.media?.length || 0})
                    </span>
                  </h4>
                  <p className="text-[11px] text-muted-foreground mt-0.5">
                    {t("products.variants.previewsHint")}
                  </p>
                </div>

                <div>
                  <input
                    ref={variantMediaInputRef}
                    type="file"
                    multiple
                    accept="image/png,image/jpeg,image/webp,image/avif"
                    className="hidden"
                    onChange={(e) => {
                      if (e.target.files && e.target.files.length > 0) {
                        handleUploadVariantMedia(selectedVariantForMedia.id, e.target.files);
                      }
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => variantMediaInputRef.current?.click()}
                    disabled={uploadingVariantMedia}
                    className="flex items-center gap-1.5 px-3 py-1.5 bg-background border border-border hover:border-accent-brand/50 text-foreground hover:text-accent-brand text-xs font-semibold rounded-xl transition-all shadow-2xs disabled:opacity-60"
                  >
                    {uploadingVariantMedia ? (
                      <Loader2 className="size-3.5 animate-spin text-accent-brand" />
                    ) : (
                      <Plus className="size-3.5 text-accent-brand" />
                    )}
                    <span>{t("products.variants.uploadPreview")}</span>
                  </button>
                </div>
              </div>

              {/* Previews Gallery Grid */}
              {Array.isArray(selectedVariantForMedia.media) && selectedVariantForMedia.media.length > 0 ? (
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
                  {selectedVariantForMedia.media.map((med: any, idx: number) => (
                    <div
                      key={med.id || `med-${idx}`}
                      className="group relative rounded-xl border border-border bg-background overflow-hidden aspect-square shadow-2xs"
                    >
                      <img
                        src={normalizeMediaUrl(med.url)}
                        alt={`Preview ${idx + 1}`}
                        className="size-full object-cover"
                        onError={(e) => {
                          (e.currentTarget as HTMLImageElement).src = "/placeholder-product.png";
                        }}
                      />
                      <span className="absolute top-1.5 left-1.5 bg-black/60 text-white text-[9px] font-bold px-1.5 py-0.2 rounded-md">
                        #{idx + 1}
                      </span>
                      <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleDeleteVariantMedia(selectedVariantForMedia.id, med.id)}
                          className="p-2 rounded-xl bg-destructive text-white hover:bg-destructive/90 transition-all shadow-sm"
                          title={t("products.variants.deletePreview")}
                        >
                          <Trash2 className="size-4" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div
                  onClick={() => variantMediaInputRef.current?.click()}
                  className="border-2 border-dashed border-border hover:border-accent-brand/60 rounded-xl p-6 text-center cursor-pointer transition-all bg-background/40 hover:bg-secondary/40 flex flex-col items-center justify-center gap-1.5"
                >
                  <Upload className="size-5 text-muted-foreground" />
                  <span className="text-xs font-medium text-muted-foreground">
                    {t("products.variants.noPreviews")}
                  </span>
                  <span className="text-[11px] text-accent-brand font-semibold">
                    {t("products.variants.uploadPreview")}
                  </span>
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="flex justify-end pt-3 border-t border-border">
              <button
                type="button"
                onClick={() => setSelectedVariantForMedia(null)}
                className="px-5 py-2 bg-accent-brand text-white text-xs font-semibold rounded-xl hover:bg-accent-brand/90 transition-all shadow-xs"
              >
                {t("actions.close") || "Done"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Quick Variant Inventory Adjustment Modal */}
      {quickStockVariant && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-sm w-full p-5 shadow-xl space-y-4">
            <div className="flex items-start justify-between border-b border-border pb-3">
              <div>
                <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
                  <Boxes className="size-4 text-accent-brand" />
                  <span>{t("products.variants.adjustStock")}</span>
                </h3>
                <p className="text-xs text-muted-foreground mt-0.5 font-medium">
                  {quickStockVariant.title} &bull; <span className="font-mono">{quickStockVariant.sku || "—"}</span>
                </p>
              </div>
              <button
                type="button"
                onClick={() => setQuickStockVariant(null)}
                className="p-1 rounded-lg text-muted-foreground hover:text-foreground"
              >
                <X className="size-4" />
              </button>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between p-2.5 rounded-xl bg-secondary/30 border border-border text-xs">
                <span className="text-muted-foreground">{t("products.variants.currentStock")}</span>
                <span className="font-bold text-foreground">
                  {t("products.units", { count: quickStockVariant.stock?.availableQuantity ?? 0 })}
                </span>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">
                  {t("products.form.stockQty")} *
                </label>
                <input
                  type="number"
                  min="0"
                  required
                  value={quickStockAmount}
                  onChange={(e) => setQuickStockAmount(e.target.value)}
                  className="w-full px-3 py-2 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  placeholder="New stock quantity"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">
                  {t("products.form.auditReason")}
                </label>
                <select
                  value={quickStockReason}
                  onChange={(e) => setQuickStockReason(e.target.value)}
                  className="w-full px-3 py-2 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                >
                  <option value="restock">{t("products.form.reasonRestock")}</option>
                  <option value="adjustment">{t("products.form.reasonAdjustment")}</option>
                  <option value="return">{t("products.form.reasonReturn")}</option>
                  <option value="sale">{t("products.form.reasonSale")}</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2 border-t border-border">
              <button
                type="button"
                onClick={() => setQuickStockVariant(null)}
                className="px-4 py-2 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground transition-all"
              >
                {t("actions.cancel")}
              </button>
              <button
                type="button"
                onClick={handleApplyQuickStock}
                disabled={savingQuickStock}
                className="flex items-center gap-1.5 px-4 py-2 bg-accent-brand text-white text-xs font-semibold rounded-xl hover:bg-accent-brand/90 transition-all shadow-xs disabled:opacity-60"
              >
                {savingQuickStock && <Loader2 className="size-3.5 animate-spin" />}
                <span>{t("actions.save")}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
