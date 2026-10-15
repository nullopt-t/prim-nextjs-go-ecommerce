import { useState, useEffect } from "react";
import { Plus, Trash2, X } from "lucide-react";
import { adminService } from "@/services/admin";
import { toast } from "sonner";

export function AdminPromotions() {
  const [promotions, setPromotions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const loadPromos = async () => {
    try {
      setLoading(true);
      const res: any = await adminService.getPromotions();
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setPromotions(list.map((p: any) => ({
        id: p.id,
        code: p.code,
        type: p.discountType === "percentage" ? "Percentage" : "Fixed Amount",
        value: p.discountType === "percentage" ? `${p.discountValue}%` : `$${(p.discountValue / 100).toFixed(2)}`,
        discountPercent: p.discountType === "percentage" ? p.discountValue : 0,
        fixedDiscount: p.discountType === "fixed_amount" ? p.discountValue / 100 : 0,
        usageLimit: p.usageLimit || 500,
        currentUsage: 12,
        status: p.status ? p.status.charAt(0).toUpperCase() + p.status.slice(1) : "Active",
        startDate: p.startsAt ? p.startsAt.slice(0, 10) : "2026-09-01",
        endDate: p.expiresAt ? p.expiresAt.slice(0, 10) : "2026-12-31"
      })));
    } catch {
      toast.error("Failed to load promotions from server");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPromos();
  }, []);

  const [formData, setFormData] = useState({
    code: "",
    type: "Percentage",
    value: "15",
    usageLimit: 200,
    startDate: "2026-08-01",
    endDate: "2026-12-31"
  });

  const handleToggleStatus = async (id: string, currentStatus: string) => {
    const nextStatus = currentStatus === "Active" ? "disabled" : "active";
    try {
      await adminService.updatePromotionStatus(id, nextStatus);
      toast.success("Promotion status updated!");
      loadPromos();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to update status");
    }
  };

  const handleCreatePromo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.code) return;

    const isPct = formData.type === "Percentage";
    const numVal = parseFloat(formData.value.replace(/[^0-9.]/g, "")) || 10;
    const discountVal = isPct ? Math.min(100, Math.max(1, Math.round(numVal))) : Math.round(numVal * 100);

    const payload = {
      code: formData.code.toUpperCase().trim(),
      discountType: isPct ? "percentage" : "fixed_amount",
      discountValue: discountVal,
      currency: "USD",
      minCartAmount: 5000,
      usageLimit: Number(formData.usageLimit) || 100,
      status: "active",
      startsAt: new Date(formData.startDate).toISOString(),
      expiresAt: new Date(formData.endDate).toISOString(),
    };

    try {
      await adminService.createPromotion(payload);
      toast.success(`Coupon code ${payload.code} created!`);
      setIsModalOpen(false);
      loadPromos();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to create coupon code");
    }
  };

  const handleDelete = async (id: string) => {
    if (confirm("Delete promo code?")) {
      try {
        await adminService.deletePromotion(id);
        toast.success("Promotion deleted");
        loadPromos();
      } catch (err: any) {
        toast.error(err?.response?.data?.message || "Failed to delete promotion");
      }
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Promotions & Discount Codes</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Create checkout coupons, percentage discounts, free shipping rules, and seasonal campaigns.
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>Create Coupon Code</span>
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {promotions.map((promo) => (
          <div key={promo.id} className="bg-background border border-border rounded-2xl p-6 shadow-xs flex flex-col justify-between space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <span className="font-mono font-bold text-lg text-accent-brand tracking-wider bg-accent-brand/10 px-3 py-1 rounded-xl border border-accent-brand/20">
                  {promo.code}
                </span>
                <div className="font-bold text-sm text-foreground mt-3">{promo.value}</div>
                <div className="text-xs text-muted-foreground mt-0.5">Type: {promo.type}</div>
              </div>
              <button
                onClick={() => handleToggleStatus(promo.id, promo.status)}
                className={`px-2.5 py-1 text-[10px] font-bold rounded-full transition-colors ${
                  promo.status === "Active"
                    ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                    : "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300"
                }`}
              >
                {promo.status}
              </button>
            </div>

            <div className="space-y-2 pt-3 border-t border-border text-xs">
              <div className="flex justify-between text-muted-foreground">
                <span>Usage Progress:</span>
                <span className="font-bold text-foreground">
                  {promo.currentUsage} / {promo.usageLimit} redeemed
                </span>
              </div>
              <div className="w-full h-1.5 bg-secondary rounded-full overflow-hidden">
                <div
                  className="h-full bg-accent-brand rounded-full"
                  style={{ width: `${Math.min(100, (promo.currentUsage / promo.usageLimit) * 100)}%` }}
                ></div>
              </div>
              <div className="flex justify-between text-[10px] text-muted-foreground pt-1">
                <span>Valid: {promo.startDate}</span>
                <span>Expires: {promo.endDate}</span>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button onClick={() => handleDelete(promo.id)} className="p-1.5 text-muted-foreground hover:text-destructive">
                <Trash2 className="size-4" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">Create Coupon Code</h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <form onSubmit={handleCreatePromo} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Promo Code *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. SUMMER25"
                  value={formData.code}
                  onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs font-mono uppercase text-foreground"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">Discount Type</label>
                  <select
                    value={formData.type}
                    onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  >
                    <option value="Percentage">Percentage OFF</option>
                    <option value="Fixed Amount">Fixed $ OFF</option>
                    <option value="Free Shipping">Free Shipping</option>
                  </select>
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">Value Label</label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. 25% OFF"
                    value={formData.value}
                    onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Usage Limit (Max Uses)</label>
                <input
                  type="number"
                  value={formData.usageLimit}
                  onChange={(e) => setFormData({ ...formData, usageLimit: Number(e.target.value) || 0 })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">Cancel</button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white rounded-xl">Create Code</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
