"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAddresses } from "@/modules/account/model/useUserAccount";
import {
	createAddressSchema,
	INITIAL_ADDRESS,
	type AddressFormValues,
} from "@/modules/account/model/addressSchema";
import { useAuthContext } from "@/modules/auth/model/authContext";
import {
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
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { copyTextToClipboard } from "@/shared/lib/clipboard";
import { CountrySelect } from "@/shared/ui/CountrySelect";
import {
  Form,
  FormControl,
  FormField,
  FormInput,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/shared/ui/form/form";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/form/dialog";
import { Checkbox } from "@/shared/ui/form/checkbox";

function FieldHint({ active, text }: { active: boolean; text: string }) {
  return (
    <div
      className={`overflow-hidden transition-all duration-200 ease-out ${
        active ? "max-h-12 opacity-100 mt-1" : "max-h-0 opacity-0"
      }`}
    >
      <p className="text-[11px] text-accent-brand/90 dark:text-accent-brand leading-tight">
        {text}
      </p>
    </div>
  );
}

export default function AddressesContent() {
  const t = useTranslations("addresses");
  const tCommon = useTranslations("common");
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

  const addressSchema = useMemo(() => createAddressSchema(t), [t]);

  const form = useForm<AddressFormValues>({
    resolver: zodResolver(addressSchema),
    defaultValues: INITIAL_ADDRESS,
  });

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [modalError, setModalError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [focusedField, setFocusedField] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [settingDefaultId, setSettingDefaultId] = useState<string | null>(null);
  const [isGiftMode, setIsGiftMode] = useState(false);
  const [addressToDelete, setAddressToDelete] = useState<any | null>(null);
  const [copiedPhoneId, setCopiedPhoneId] = useState<string | null>(null);

  const copyPhoneToClipboard = async (id: string, phone: string) => {
    if (!phone) return;
    const copied = await copyTextToClipboard(phone);
    if (!copied) {
      notify.error(tCommon("errors.copyFailed"));
      return;
    }
    setCopiedPhoneId(id);
    notify.success(t("phoneCopied"));
    setTimeout(() => {
      setCopiedPhoneId((curr) => (curr === id ? null : curr));
    }, 2000);
  };

  const openAddModal = () => {
    setEditingId(null);
    setModalError(null);
    setIsGiftMode(false);
    form.reset({
      ...INITIAL_ADDRESS,
      fullName: user?.name || "",
      phone: user?.phone || "",
      isDefault: addresses.length === 0,
    });
    setIsModalOpen(true);
  };

  const openEditModal = (addr: any) => {
    setEditingId(addr.id);
    setModalError(null);
    form.reset({
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

  const submitAddress = async (values: AddressFormValues) => {
    setSaving(true);
    setModalError(null);
    try {
      if (editingId) {
        const res = await updateAddress(editingId, values);
        if (res.success) {
          setModalError(null);
          setIsModalOpen(false);
        } else {
          const msg = apiErrorMessage(res, t("failedToUpdate"));
          setModalError(msg);
          notify.error(msg);
        }
      } else {
        const res = await addAddress(values);
        if (res.success) {
          setModalError(null);
          setIsModalOpen(false);
        } else {
          const msg = apiErrorMessage(res, t("failedToAdd"));
          setModalError(msg);
          notify.error(msg);
        }
      }
    } catch (err) {
      const msg = apiErrorMessage(err, editingId ? t("failedToUpdate") : t("failedToAdd"));
      setModalError(msg);
      notify.error(msg);
    } finally {
      setSaving(false);
    }
  };

  const handleSetDefault = async (id: string) => {
    setSettingDefaultId(id);
    setActionError(null);
    try {
      const res = await setDefaultAddress(id);
      if (res.success) {
        setActionError(null);
      } else {
        const msg = apiErrorMessage(res, t("failedToSetDefault"));
        setActionError(msg);
        notify.error(msg);
      }
    } catch (err) {
      const msg = apiErrorMessage(err, t("failedToSetDefault"));
      setActionError(msg);
      notify.error(msg);
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
    setActionError(null);
    try {
      const res = await deleteAddress(id);
      if (res.success) {
        setActionError(null);
        setAddressToDelete(null);
      } else {
        const msg = apiErrorMessage(res, t("failedToDelete"));
        setActionError(msg);
        notify.error(msg);
      }
    } catch (err) {
      const msg = apiErrorMessage(err, t("failedToDelete"));
      setActionError(msg);
      notify.error(msg);
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
      {(error || actionError) && (
        <div className="p-4 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive text-sm flex items-center justify-between gap-2">
          <p>{actionError || error}</p>
          {actionError && (
            <button
              type="button"
              onClick={() => setActionError(null)}
              className="text-xs underline hover:no-underline font-medium"
            >
              {tCommon("dismiss", { defaultMessage: "Dismiss" })}
            </button>
          )}
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
                          onClick={() => addr.phone && copyPhoneToClipboard(addr.id, addr.phone)}
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
      <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
        <DialogContent className="max-w-lg gap-0 p-0">
          <DialogHeader className="px-6 py-4 border-b border-border/70">
            <DialogTitle className="text-lg">
              {editingId ? t("editAddress") : t("addNewDelivery")}
            </DialogTitle>
          </DialogHeader>

          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(submitAddress)}
              className="p-6 overflow-y-auto space-y-4.5"
              noValidate
            >
              {modalError && (
                <div
                  role="alert"
                  className="rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-xs font-semibold text-destructive flex items-center gap-2"
                >
                  <AlertTriangle className="size-4 shrink-0" aria-hidden="true" />
                  <span>{modalError}</span>
                </div>
              )}

              {/* Friendly Label Selector */}
              <div>
                <span className="block text-xs font-semibold text-foreground mb-1.5">
                  {t("labels.addressLabel")}
                </span>
                <div className="grid grid-cols-3 gap-2">
                  {[
                    { key: "Home", text: t("labels.home") },
                    { key: "Office", text: t("labels.office") },
                    { key: "Other", text: t("labels.other") },
                  ].map((lbl) => {
                    const isActive =
                      form.watch("label").toLowerCase() === lbl.key.toLowerCase();
                    return (
                      <button
                        key={lbl.key}
                        type="button"
                        onClick={() =>
                          form.setValue("label", lbl.key, { shouldDirty: true })
                        }
                        className={`py-2 px-3 rounded-xl border text-xs font-medium cursor-pointer transition-all ${
                          isActive
                            ? "border-accent-brand bg-accent-brand/10 text-accent-brand font-semibold"
                            : "border-border/80 bg-secondary/40 text-muted-foreground hover:text-foreground"
                        }`}
                      >
                        {lbl.text}
                      </button>
                    );
                  })}
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
                        form.setValue("fullName", "");
                        form.setValue("phone", "");
                        form.setValue("label", "Other");
                      } else {
                        form.setValue("fullName", user?.name || "");
                        form.setValue("phone", user?.phone || "");
                        form.setValue("label", "Home");
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
                        form.setValue("fullName", user.name || "");
                        form.setValue("phone", user.phone || "");
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
                <FormField
                  control={form.control}
                  name="fullName"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {isGiftMode
                          ? `${t("labels.recipient")} *`
                          : t("labels.fullName")}
                      </FormLabel>
                      <FormControl>
                        <FormInput
                          {...field}
                          onFocus={() => setFocusedField("fullName")}
                          onBlur={() => {
                            field.onBlur();
                            setFocusedField((cur) =>
                              cur === "fullName" ? null : cur,
                            );
                          }}
                          placeholder={
                            isGiftMode
                              ? t("placeholders.giftReceiver")
                              : t("placeholders.fullName")
                          }
                        />
                      </FormControl>
                      <FieldHint
                        active={focusedField === "fullName"}
                        text={t("hints.fullName")}
                      />
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="phone"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("labels.phone")}</FormLabel>
                      <FormControl>
                        <FormInput
                          {...field}
                          type="tel"
                          dir="ltr"
                          onFocus={() => setFocusedField("phone")}
                          onBlur={() => {
                            field.onBlur();
                            setFocusedField((cur) =>
                              cur === "phone" ? null : cur,
                            );
                          }}
                          placeholder={t("placeholders.phone")}
                        />
                      </FormControl>
                      <FieldHint
                        active={focusedField === "phone"}
                        text={t("hints.phone")}
                      />
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {/* Address Line 1 */}
              <FormField
                control={form.control}
                name="line1"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("labels.line1")}</FormLabel>
                    <FormControl>
                      <FormInput
                        {...field}
                        onFocus={() => setFocusedField("line1")}
                        onBlur={() => {
                          field.onBlur();
                          setFocusedField((cur) =>
                            cur === "line1" ? null : cur,
                          );
                        }}
                        placeholder={t("placeholders.line1")}
                      />
                    </FormControl>
                    <FieldHint
                      active={focusedField === "line1"}
                      text={t("hints.line1")}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Address Line 2 */}
              <FormField
                control={form.control}
                name="line2"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("labels.line2")}</FormLabel>
                    <FormControl>
                      <FormInput
                        {...field}
                        onFocus={() => setFocusedField("line2")}
                        onBlur={() => {
                          field.onBlur();
                          setFocusedField((cur) =>
                            cur === "line2" ? null : cur,
                          );
                        }}
                        placeholder={t("placeholders.line2")}
                      />
                    </FormControl>
                    <FieldHint
                      active={focusedField === "line2"}
                      text={t("hints.line2")}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* City, State, Postal Code */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3.5">
                <FormField
                  control={form.control}
                  name="city"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("labels.city")}</FormLabel>
                      <FormControl>
                        <FormInput
                          {...field}
                          onFocus={() => setFocusedField("city")}
                          onBlur={() => {
                            field.onBlur();
                            setFocusedField((cur) =>
                              cur === "city" ? null : cur,
                            );
                          }}
                          placeholder={t("placeholders.city")}
                        />
                      </FormControl>
                      <FieldHint
                        active={focusedField === "city"}
                        text={t("hints.city")}
                      />
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="state"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("labels.state")}</FormLabel>
                      <FormControl>
                        <FormInput
                          {...field}
                          placeholder={t("placeholders.state")}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="postalCode"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("labels.postalCode")}</FormLabel>
                      <FormControl>
                        <FormInput
                          {...field}
                          onFocus={() => setFocusedField("postalCode")}
                          onBlur={() => {
                            field.onBlur();
                            setFocusedField((cur) =>
                              cur === "postalCode" ? null : cur,
                            );
                          }}
                          placeholder={t("placeholders.postalCode")}
                        />
                      </FormControl>
                      <FieldHint
                        active={focusedField === "postalCode"}
                        text={t("hints.postalCode")}
                      />
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {/* Country */}
              <FormField
                control={form.control}
                name="country"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("labels.country")}</FormLabel>
                    <CountrySelect
                      value={field.value}
                      onChange={(countryName) => field.onChange(countryName)}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Default checkbox */}
              <FormField
                control={form.control}
                name="isDefault"
                render={({ field }) => (
                  <FormItem className="flex-col">
                    <div className="flex items-center gap-2.5 pt-1">
                      <FormControl>
                        <Checkbox
                          checked={field.value}
                          onCheckedChange={(checked) =>
                            field.onChange(checked === true)
                          }
                        />
                      </FormControl>
                      <FormLabel className="cursor-pointer select-none font-medium">
                        {t("labels.makeDefault")}
                      </FormLabel>
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

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
          </Form>
        </DialogContent>
      </Dialog>

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
