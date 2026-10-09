"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useAddresses } from "@/hooks/useUserAccount";
import {
  MapPin,
  Plus,
  Trash2,
  Edit2,
  CheckCircle2,
  Building,
  Home,
  Briefcase,
  Phone,
  AlertCircle,
  X,
  Loader2,
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

function getLabelIcon(label?: string) {
  const l = (label || "").toLowerCase();
  if (l.includes("work") || l.includes("office")) return Briefcase;
  if (l.includes("building") || l.includes("warehouse")) return Building;
  return Home;
}

export default function AddressesContent() {
  const t = useTranslations("addresses");
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

  const openAddModal = () => {
    setEditingId(null);
    setFormData({
      ...INITIAL_FORM,
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
      toast.error("Please fill in all mandatory fields (Name, Street address, City, Country)");
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
          toast.error(res.error || "Failed to update address");
        }
      } else {
        const res = await addAddress(formData);
        if (res.success) {
          toast.success(t("addedSuccess"));
          setIsModalOpen(false);
        } else {
          toast.error(res.error || "Failed to add address");
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
        toast.error(res.error || "Failed to set default address");
      }
    } finally {
      setSettingDefaultId(null);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    setDeletingId(id);
    try {
      const res = await deleteAddress(id);
      if (res.success) {
        toast.success(t("removedSuccess"));
      } else {
        toast.error(res.error || "Failed to delete address");
      }
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border/60">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {t("subtitle")}
          </p>
        </div>
        <button
          onClick={openAddModal}
          className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl bg-accent-brand text-white text-sm font-semibold hover:bg-accent-brand/90 transition-all shadow-xs cursor-pointer shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("addAddress")}</span>
        </button>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="p-4 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive text-sm flex items-center gap-3">
          <AlertCircle className="size-4.5 shrink-0" />
          <p>{error}</p>
        </div>
      )}

      {/* Loading Skeleton */}
      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-5">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-56 rounded-2xl bg-secondary/50 animate-pulse border border-border/60" />
          ))}
        </div>
      ) : addresses.length === 0 ? (
        /* Empty State */
        <div className="bg-card border border-border/80 rounded-2xl p-12 text-center flex flex-col items-center justify-center gap-4 max-w-lg mx-auto my-6 shadow-xs">
          <div className="size-16 rounded-2xl bg-secondary/80 flex items-center justify-center text-muted-foreground">
            <MapPin className="size-8" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-foreground">{t("noAddresses")}</h3>
            <p className="text-sm text-muted-foreground mt-1 max-w-sm">
              {t("noAddressesDesc")}
            </p>
          </div>
          <button
            onClick={openAddModal}
            className="mt-2 inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-accent-brand text-white text-sm font-medium hover:bg-accent-brand/90 transition-all shadow-xs cursor-pointer"
          >
            <Plus className="size-4" />
            <span>{t("addAddress")}</span>
          </button>
        </div>
      ) : (
        /* Address Grid */
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-5">
          {/* Add New Address Card Placeholder (Amazon style) */}
          <div
            onClick={openAddModal}
            className="min-h-[220px] rounded-2xl border-2 border-dashed border-border/80 hover:border-accent-brand/60 hover:bg-accent-brand/5 transition-all p-6 flex flex-col items-center justify-center text-center gap-2 cursor-pointer group"
          >
            <div className="size-12 rounded-full bg-secondary group-hover:bg-accent-brand/10 group-hover:text-accent-brand flex items-center justify-center text-muted-foreground transition-colors">
              <Plus className="size-6" />
            </div>
            <span className="font-semibold text-sm text-foreground group-hover:text-accent-brand transition-colors">
              {t("addAddress")}
            </span>
          </div>

          {addresses.map((addr) => {
            const Icon = getLabelIcon(addr.label);
            const isDefault = addr.isDefault;

            return (
              <div
                key={addr.id}
                className={`rounded-2xl border p-5 flex flex-col justify-between transition-all bg-card ${
                  isDefault
                    ? "border-accent-brand ring-2 ring-accent-brand/20 shadow-sm"
                    : "border-border/80 hover:border-border hover:shadow-xs"
                }`}
              >
                <div>
                  {/* Top Bar with Label & Default Badge */}
                  <div className="flex items-center justify-between gap-2 mb-3">
                    <span className="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg bg-secondary text-foreground capitalize">
                      <Icon className="size-3.5 text-accent-brand" />
                      {addr.label ? (addr.label.toLowerCase() === "home" ? t("labels.home") : addr.label.toLowerCase() === "office" ? t("labels.office") : addr.label) : t("labels.other")}
                    </span>
                    {isDefault && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-bold text-accent-brand bg-accent-brand/10 border border-accent-brand/20 px-2.5 py-0.5 rounded-full">
                        <CheckCircle2 className="size-3" />
                        {t("defaultBadge")}
                      </span>
                    )}
                  </div>

                  {/* Address Content */}
                  <div className="space-y-1 text-sm">
                    <p className="font-bold text-foreground text-base tracking-tight">
                      {addr.fullName}
                    </p>
                    <p className="text-muted-foreground leading-snug">
                      {addr.line1}
                      {addr.line2 && <span className="block">{addr.line2}</span>}
                    </p>
                    <p className="text-muted-foreground">
                      {addr.city}
                      {addr.state && `, ${addr.state}`}
                      {addr.postalCode && ` ${addr.postalCode}`}
                    </p>
                    <p className="text-muted-foreground font-medium">{addr.country}</p>

                    {addr.phone && (
                      <p className="text-xs text-muted-foreground/80 flex items-center gap-1.5 pt-2">
                        <Phone className="size-3 text-muted-foreground" />
                        <span>{addr.phone}</span>
                      </p>
                    )}
                  </div>
                </div>

                {/* Bottom Actions */}
                <div className="pt-4 mt-4 border-t border-border/60 flex items-center justify-between text-xs">
                  <div className="flex items-center gap-3 font-medium">
                    <button
                      onClick={() => openEditModal(addr)}
                      className="text-foreground hover:text-accent-brand inline-flex items-center gap-1 cursor-pointer transition-colors"
                    >
                      <Edit2 className="size-3.5" />
                      <span>{t("edit")}</span>
                    </button>
                    <span className="text-border">|</span>
                    <button
                      disabled={deletingId === addr.id}
                      onClick={() => handleDelete(addr.id)}
                      className="text-muted-foreground hover:text-destructive inline-flex items-center gap-1 cursor-pointer transition-colors disabled:opacity-50"
                    >
                      {deletingId === addr.id ? (
                        <Loader2 className="size-3.5 animate-spin" />
                      ) : (
                        <Trash2 className="size-3.5" />
                      )}
                      <span>{t("remove")}</span>
                    </button>
                  </div>

                  {!isDefault && (
                    <button
                      disabled={settingDefaultId === addr.id}
                      onClick={() => handleSetDefault(addr.id)}
                      className="text-accent-brand hover:underline font-semibold cursor-pointer disabled:opacity-50 inline-flex items-center gap-1"
                    >
                      {settingDefaultId === addr.id ? (
                        <Loader2 className="size-3 animate-spin" />
                      ) : null}
                      <span>{t("setAsDefault")}</span>
                    </button>
                  )}
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

              {/* Full Name & Phone */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                <div>
                  <label className="block text-xs font-semibold text-foreground mb-1">
                    {t("labels.fullName")}
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.fullName}
                    onFocus={() => setFocusedField("fullName")}
                    onBlur={() => setFocusedField((cur) => (cur === "fullName" ? null : cur))}
                    onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
                    placeholder={t("placeholders.fullName")}
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
    </div>
  );
}
