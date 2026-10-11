"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useAddresses } from "@/hooks/useUserAccount";
import { useAuthContext } from "@/context/AuthContext";
import {
  X,
  Loader2,
  Plus,
  MapPin,
  Pencil,
  Trash2,
  Check,
  Gift,
  RotateCcw,
  AlertTriangle,
  Copy,
} from "lucide-react";
import { toast } from "sonner";
import { CountrySelect } from "@/components/ui/CountrySelect";

interface AddressFormData {
  label: string;
  fullName: string;
  phone: string;
  line1: string;
  line2: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  isDefault: boolean;
}

const INITIAL_FORM: AddressFormData = {
  label: "Home",
  fullName: "",
  phone: "",
  line1: "",
  line2: "",
  city: "",
  state: "",
  postalCode: "",
  country: "Egypt",
  isDefault: false,
};

export default function AddressesContent() {
  const t = useTranslations("addresses");
  const { user } = useAuthContext();
  const {
    addresses,
    loading,
    error,
    addAddress,
    updateAddress,
    setDefaultAddress,
    deleteAddress,
  } = useAddresses();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<AddressFormData>(INITIAL_FORM);
  const [focusedField, setFocusedField] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [settingDefaultId, setSettingDefaultId] = useState<string | null>(null);
  const [isGiftMode, setIsGiftMode] = useState(false);
  const [addressToDelete, setAddressToDelete] = useState<any | null>(null);
  const [copiedPhoneId, setCopiedPhoneId] = useState<string | null>(null);

  const copyPhoneToClipboard = (id: string, phone: string) => {
    if (!phone) return;
    navigator.clipboard.writeText(phone);
    setCopiedPhoneId(id);
    toast.success(t("phoneCopied"));
    setTimeout(() => {
      setCopiedPhoneId((curr) => (curr === id ? null : curr));
    }, 2000);
  };

  const openAddModal = () => {
    setEditingId(null);
    setIsGiftMode(false);
    setFormData({
      ...INITIAL_FORM,
      fullName: user?.name || "",
      phone: user?.phone || "",
      isDefault: addresses.length === 0,
    });
    setIsModalOpen(true);
  };

  const openEditModal = (addr: any) => {
    setEditingId(addr.id);
    setFormData({
      label: addr.label || "Home",
      fullName: addr.fullName || "",
      phone: addr.phone || "",
      line1: addr.line1 || "",
      line2: addr.line2 || "",
      city: addr.city || "",
      state: addr.state || "",
      postalCode: addr.postalCode || "",
      country: addr.country || "Egypt",
      isDefault: addr.isDefault || false,
    });
    setIsModalOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.fullName.trim() || !formData.line1.trim() || !formData.city.trim() || !formData.country.trim()) {
      toast.error(t("validationError"));
      return;
    }

    setSaving(true);
    try {
      if (editingId) {
        const res = await updateAddress(editingId, formData);
        if (res.success) {
          toast.success(t("updatedSuccess"));
          setIsModalOpen(false);
        } else {
          toast.error(res.error || t("failedToUpdate"));
        }
      } else {
        const res = await addAddress(formData);
        if (res.success) {
          toast.success(t("addedSuccess"));
          setIsModalOpen(false);
        } else {
          toast.error(res.error || t("failedToAdd"));
        }
      }
    } finally {
      setSaving(false);
    }
  };

  const handleSetDefault = async (id: string) => {
    setSettingDefaultId(id);
    try {
      const res = await setDefaultAddress(id);
      if (res.success) {
        toast.success(t("defaultSuccess"));
      } else {
        toast.error(res.error || t("failedToSetDefault"));
      }
    } finally {
      setSettingDefaultId(null);
    }
  };

  const promptDelete = (addr: any) => {
    setAddressToDelete(addr);
  };

  const confirmDeleteAddress = async () => {
    if (!addressToDelete) return;
    const id = addressToDelete.id;
    setDeletingId(id);
    try {
      const res = await deleteAddress(id);
      if (res.success) {
        toast.success(t("removedSuccess"));
        setAddressToDelete(null);
      } else {
        toast.error(res.error || t("failedToDelete"));
      }
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-border/60">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {t("subtitle")}
          </p>
        </div>
        <button
          onClick={openAddModal}
          className="inline-flex items-center justify-center gap-1.5 px-4 py-2 rounded-xl bg-accent-brand text-white text-xs font-semibold hover:bg-accent-brand/90 transition-all shadow-xs cursor-pointer shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("addAddress")}</span>
        </button>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="p-4 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive text-sm">
          <p>{error}</p>
        </div>
      )}

      {/* Loading Skeleton */}
      {loading ? (
        <div className="bg-card border border-border rounded-xl divide-y divide-border overflow-hidden">
          {[1, 2, 3].map((i) => (
            <div key={i} className="p-4 flex items-center justify-between gap-4 animate-pulse">
              <div className="space-y-2 flex-1">
                <div className="h-4 bg-secondary/70 rounded w-1/4" />
                <div className="h-3 bg-secondary/50 rounded w-1/2" />
              </div>
              <div className="h-8 bg-secondary/60 rounded w-20" />
            </div>
          ))}
        </div>
      ) : addresses.length === 0 ? (
        /* Empty State */
        <div className="bg-card border border-border rounded-xl p-8 sm:p-12 text-center flex flex-col items-center justify-center gap-3 shadow-2xs">
          <div className="size-10 rounded-xl bg-secondary/80 flex items-center justify-center text-muted-foreground">
            <MapPin className="size-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-foreground">{t("noAddresses")}</h3>
            <p className="text-xs text-muted-foreground mt-0.5 max-w-sm">
              {t("noAddressesDesc")}
            </p>
          </div>
          <button
            onClick={openAddModal}
            className="mt-1 inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-accent-brand text-white text-xs font-semibold hover:bg-accent-brand/90 transition-all shadow-2xs cursor-pointer"
          >
            <Plus className="size-3.5" />
            <span>{t("addAddress")}</span>
          </button>
        </div>
      ) : (
        /* Address Cards Grid Styled like Auth Cart Preview */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3.5">
          {addresses.map((addr) => {
            const isDefault = addr.isDefault;

            return (
              <div
                key={addr.id}
                className={`group relative flex flex-col justify-between p-3.5 sm:p-4 rounded-xl border transition-all duration-200 shadow-2xs ${
                  isDefault
                    ? "bg-secondary/40 border-accent-brand/40 shadow-xs"
                    : "bg-secondary/30 hover:bg-secondary/50 border-border/50 hover:border-accent-brand/30"
                }`}
              >
                {/* Top: Icon Thumbnail, Recipient & Label, Default Badge */}
                <div className="flex items-start gap-3">
                  {/* Address Icon Box (matching product thumbnail from auth cart) */}
                  <div className="size-11 rounded-lg bg-background border border-border/60 overflow-hidden shrink-0 flex items-center justify-center p-1 text-muted-foreground shadow-2xs group-hover:text-accent-brand transition-colors">
                    <MapPin className="size-5" />
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 min-w-0">
                        <span className="font-semibold text-foreground text-xs sm:text-sm truncate">
                          {addr.fullName}
                        </span>
                        <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-md bg-secondary text-muted-foreground uppercase shrink-0">
                          {(() => {
                            const labelKey = addr.label ? addr.label.toLowerCase() : "other";
                            return labelKey === "home"
                              ? t("labels.home")
                              : labelKey === "office" || labelKey === "work"
                              ? t("labels.office")
                              : labelKey === "other"
                              ? t("labels.other")
                              : addr.label;
                          })()}
                        </span>
                      </div>

                      {isDefault && (
                        <span className="text-[10px] font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded-full inline-flex items-center gap-1 shrink-0">
                          <Check className="size-2.5" />
                          <span>{t("defaultBadge")}</span>
                        </span>
                      )}
                    </div>

                    {/* Street & Location */}
                    <p className="text-xs text-foreground/90 font-medium line-clamp-1 mt-1">
                      {[addr.line1, addr.line2].filter(Boolean).join(", ")}
                    </p>
                    <p className="text-[11px] text-muted-foreground line-clamp-1 mt-0.5">
                      {[addr.city, addr.state, addr.postalCode, addr.country].filter(Boolean).join("، ")}
                    </p>

                    {/* Phone with copy */}
                    {addr.phone && (
                      <div className="mt-1">
                        <button
                          type="button"
                          onClick={() => copyPhoneToClipboard(addr.id, addr.phone)}
                          className="inline-flex items-center gap-1 text-muted-foreground hover:text-foreground cursor-pointer text-[11px] transition-colors"
                        >
                          <span className="font-mono tabular-nums [direction:ltr] [unicode-bidi:isolate]">
                            {addr.phone}
                          </span>
                          {copiedPhoneId === addr.id ? (
                            <Check className="size-3 text-emerald-500 shrink-0" />
                          ) : (
                            <Copy className="size-3 opacity-60 hover:opacity-100 shrink-0" />
                          )}
                        </button>
                      </div>
                    )}
                  </div>
                </div>

                {/* Bottom Actions Bar */}
                <div className="mt-3 pt-2.5 border-t border-border/50 flex items-center justify-between gap-2">
                  <div>
                    {!isDefault && (
                      <button
                        type="button"
                        disabled={settingDefaultId === addr.id}
                        onClick={() => handleSetDefault(addr.id)}
                        className="text-xs font-semibold text-accent-brand hover:underline cursor-pointer disabled:opacity-50 inline-flex items-center gap-1"
                      >
                        {settingDefaultId === addr.id && (
                          <Loader2 className="size-3 animate-spin" />
                        )}
                        <span>{t("setAsDefault")}</span>
                      </button>
                    )}
                  </div>

                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      onClick={() => openEditModal(addr)}
                      className="px-2.5 py-1 rounded-md text-xs font-medium text-foreground hover:bg-secondary cursor-pointer inline-flex items-center gap-1 transition-colors"
                    >
                      <Pencil className="size-3" />
                      <span>{t("edit")}</span>
                    </button>
                    <button
                      type="button"
                      disabled={deletingId === addr.id}
                      onClick={() => promptDelete(addr)}
                      className="px-2.5 py-1 rounded-md text-xs font-medium text-destructive hover:bg-destructive/10 cursor-pointer inline-flex items-center gap-1 transition-colors disabled:opacity-40"
                    >
                      {deletingId === addr.id ? (
                        <Loader2 className="size-3 animate-spin" />
                      ) : (
                        <Trash2 className="size-3" />
                      )}
                      <span>{t("remove")}</span>
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Add / Edit Address Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="bg-card border border-border w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
            {/* Modal Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-border/70">
              <h2 className="text-lg font-bold text-foreground">
                {editingId ? t("editAddress") : t("addNewDelivery")}
              </h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="size-8 rounded-lg flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer transition-colors"
              >
                <X className="size-4.5" />
              </button>
            </div>

            {/* Modal Body / Form */}
            <form onSubmit={handleSubmit} className="p-6 overflow-y-auto space-y-4.5">
              {/* Friendly Label Selector */}
              <div>
                <label className="block text-xs font-semibold text-foreground mb-1.5">
                  {t("labels.addressLabel")}
                </label>
                <div className="grid grid-cols-3 gap-2">
                  {[
                    { key: "Home", text: t("labels.home") },
                    { key: "Office", text: t("labels.office") },
                    { key: "Other", text: t("labels.other") },
                  ].map((lbl) => (
                    <button
                      key={lbl.key}
                      type="button"
                      onClick={() => setFormData({ ...formData, label: lbl.key })}
                      className={`py-2 px-3 rounded-xl border text-xs font-medium cursor-pointer transition-all ${
                        formData.label.toLowerCase() === lbl.key.toLowerCase()
                          ? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold"
                          : "border-border/80 bg-secondary/40 text-muted-foreground hover:text-foreground"
                      }`}
                    >
                      {lbl.text}
                    </button>
                  ))}
                </div>
              </div>

              {/* Recipient / Delivery Persona Selector */}
              <div className="p-3 rounded-xl bg-secondary/30 border border-border/70 space-y-2.5">
                <div className="flex items-center justify-between gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      const next = !isGiftMode;
                      setIsGiftMode(next);
                      if (next) {
                        // Switch to gift mode: clear recipient name so they can enter the gift receiver
                        setFormData((prev) => ({
                          ...prev,
                          fullName: "",
                          phone: "",
                          label: "Other",
                        }));
                      } else {
                        // Switch back to ordering for self
                        setFormData((prev) => ({
                          ...prev,
                          fullName: user?.name || "",
                          phone: user?.phone || "",
                          label: "Home",
                        }));
                      }
                    }}
                    className={`inline-flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-xs font-semibold cursor-pointer transition-all ${
                      isGiftMode
                        ? "bg-amber-500/15 text-amber-700 dark:text-amber-400 border border-amber-500/30"
                        : "bg-secondary text-foreground hover:bg-secondary/80 border border-border/80"
                    }`}
                  >
                    <Gift className="size-3.5" />
                    <span>{t("labels.isGift")}</span>
                  </button>

                  {user?.name && (
                    <button
                      type="button"
                      onClick={() => {
                        setIsGiftMode(false);
                        setFormData((prev) => ({
                          ...prev,
                          fullName: user.name || "",
                          phone: user.phone || "",
                        }));
                      }}
                      className="text-[11px] text-accent-brand hover:underline inline-flex items-center gap-1 cursor-pointer font-medium"
                      title={t("labels.fillMyInfo")}
                    >
                      <RotateCcw className="size-3" />
                      <span>{t("labels.fillMyInfo")}</span>
                    </button>
                  )}
                </div>

                {isGiftMode && (
                  <p className="text-[11px] text-amber-700 dark:text-amber-400/90 leading-tight">
                    {t("labels.isGiftDesc")}
                  </p>
                )}
              </div>

              {/* Full Name & Phone */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                <div>
                  <label className="block text-xs font-semibold text-foreground mb-1">
                    {isGiftMode ? `${t("labels.recipient")} *` : t("labels.fullName")}
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.fullName}
                    onFocus={() => setFocusedField("fullName")}
                    onBlur={() => setFocusedField((cur) => (cur === "fullName" ? null : cur))}
                    onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
                    placeholder={isGiftMode ? "e.g. Gift Receiver Name" : t("placeholders.fullName")}
                    className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                  />
                  <div
                    className={`overflow-hidden transition-all duration-200 ease-out ${
                      focusedField === "fullName" ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
                    }`}
                  >
                    <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
                      {t("hints.fullName")}
                    </p>
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-foreground mb-1">
                    {t("labels.phone")}
                  </label>
                  <input
                    type="tel"
                    value={formData.phone}
                    onFocus={() => setFocusedField("phone")}
                    onBlur={() => setFocusedField((cur) => (cur === "phone" ? null : cur))}
                    onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                    placeholder={t("placeholders.phone")}
                    className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                  />
                  <div
                    className={`overflow-hidden transition-all duration-200 ease-out ${
                      focusedField === "phone" ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
                    }`}
                  >
                    <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
                      {t("hints.phone")}
                    </p>
                  </div>
                </div>
              </div>

              {/* Address Line 1 */}
              <div>
                <label className="block text-xs font-semibold text-foreground mb-1">
                  {t("labels.line1")}
                </label>
                <input
                  type="text"
                  required
                  value={formData.line1}
                  onFocus={() => setFocusedField("line1")}
                  onBlur={() => setFocusedField((cur) => (cur === "line1" ? null : cur))}
                  onChange={(e) => setFormData({ ...formData, line1: e.target.value })}
                  placeholder={t("placeholders.line1")}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                />
                <div
                  className={`overflow-hidden transition-all duration-200 ease-out ${
                    focusedField === "line1" ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
                  }`}
                >
                  <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
                    {t("hints.line1")}
                  </p>
                </div>
              </div>

              {/* Address Line 2 */}
              <div>
                <label className="block text-xs font-semibold text-foreground mb-1">
                  {t("labels.line2")}
                </label>
                <input
                  type="text"
                  value={formData.line2}
                  onFocus={() => setFocusedField("line2")}
                  onBlur={() => setFocusedField((cur) => (cur === "line2" ? null : cur))}
                  onChange={(e) => setFormData({ ...formData, line2: e.target.value })}
                  placeholder={t("placeholders.line2")}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                />
                <div
                  className={`overflow-hidden transition-all duration-200 ease-out ${
                    focusedField === "line2" ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
                  }`}
                >
                  <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
                    {t("hints.line2")}
                  </p>
                </div>
              </div>

              {/* City, State, Postal Code */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3.5">
                <div>
                  <label className="block text-xs font-semibold text-foreground mb-1">
                    {t("labels.city")}
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.city}
                    onFocus={() => setFocusedField("city")}
                    onBlur={() => setFocusedField((cur) => (cur === "city" ? null : cur))}
                    onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                    placeholder={t("placeholders.city")}
                    className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                  />
                  <div
                    className={`overflow-hidden transition-all duration-200 ease-out ${
                      focusedField === "city" ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
                    }`}
                  >
                    <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
                      {t("hints.city")}
                    </p>
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-foreground mb-1">
                    {t("labels.state")}
                  </label>
                  <input
                    type="text"
                    value={formData.state}
                    onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                    placeholder={t("placeholders.state")}
                    className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-foreground mb-1">
                    {t("labels.postalCode")}
                  </label>
                  <input
                    type="text"
                    value={formData.postalCode}
                    onFocus={() => setFocusedField("postalCode")}
                    onBlur={() => setFocusedField((cur) => (cur === "postalCode" ? null : cur))}
                    onChange={(e) => setFormData({ ...formData, postalCode: e.target.value })}
                    placeholder={t("placeholders.postalCode")}
                    className="w-full px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm focus:border-accent-brand focus:outline-hidden transition-colors"
                  />
                  <div
                    className={`overflow-hidden transition-all duration-200 ease-out ${
                      focusedField === "postalCode" ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
                    }`}
                  >
                    <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
                      {t("hints.postalCode")}
                    </p>
                  </div>
                </div>
              </div>

              {/* Country */}
              <div>
                <label className="block text-xs font-semibold text-foreground mb-1">
                  {t("labels.country")}
                </label>
                <CountrySelect
                  value={formData.country}
                  onChange={(countryName) =>
                    setFormData({ ...formData, country: countryName })
                  }
                />
              </div>

              {/* Default checkbox */}
              <label className="flex items-center gap-2.5 pt-1 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={formData.isDefault}
                  onChange={(e) => setFormData({ ...formData, isDefault: e.target.checked })}
                  className="size-4 rounded-md border-border text-accent-brand focus:ring-accent-brand cursor-pointer"
                />
                <span className="text-xs font-medium text-foreground">
                  {t("labels.makeDefault")}
                </span>
              </label>

              {/* Modal Footer Buttons */}
              <div className="pt-4 flex items-center justify-end gap-2.5 border-t border-border/60">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 rounded-xl border border-border/80 text-xs font-semibold text-muted-foreground hover:bg-secondary cursor-pointer transition-colors"
                >
                  {t("cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="px-5 py-2 rounded-xl bg-accent-brand text-white text-xs font-bold hover:bg-accent-brand/90 cursor-pointer transition-colors disabled:opacity-50 inline-flex items-center gap-1.5 shadow-xs"
                >
                  {saving && <Loader2 className="size-3.5 animate-spin" />}
                  <span>{editingId ? t("saveChanges") : t("submitAdd")}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Custom Delete Confirmation Modal */}
      {addressToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="bg-card border border-border w-full max-w-md rounded-2xl shadow-2xl overflow-hidden p-6 space-y-4 animate-in zoom-in-95 duration-150">
            <div className="flex items-start gap-3.5">
              <div className="size-10 rounded-xl bg-destructive/10 text-destructive flex items-center justify-center shrink-0">
                <AlertTriangle className="size-5" />
              </div>
              <div className="space-y-1">
                <h3 className="text-base font-bold text-foreground">
                  {t("deleteDialogTitle")}
                </h3>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {t("deleteDialogDesc")}
                </p>
                <div className="mt-2.5 p-2.5 rounded-lg bg-secondary/50 border border-border/60 text-xs space-y-0.5">
                  <span className="font-semibold text-foreground block">
                    {addressToDelete.fullName}
                  </span>
                  <span className="text-muted-foreground block truncate">
                    {[addressToDelete.line1, addressToDelete.city, addressToDelete.country].filter(Boolean).join(", ")}
                  </span>
                </div>
              </div>
            </div>

            <div className="pt-2 flex items-center justify-end gap-2.5">
              <button
                type="button"
                disabled={Boolean(deletingId)}
                onClick={() => setAddressToDelete(null)}
                className="px-4 py-2 rounded-xl border border-border/80 text-xs font-semibold text-muted-foreground hover:bg-secondary cursor-pointer transition-colors"
              >
                {t("cancel")}
              </button>
              <button
                type="button"
                disabled={Boolean(deletingId)}
                onClick={confirmDeleteAddress}
                className="px-4 py-2 rounded-xl bg-destructive text-destructive-foreground text-xs font-bold hover:bg-destructive/90 cursor-pointer transition-colors disabled:opacity-50 inline-flex items-center gap-1.5 shadow-xs"
              >
                {deletingId && <Loader2 className="size-3.5 animate-spin" />}
                <span>{t("confirmDeleteAction")}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
