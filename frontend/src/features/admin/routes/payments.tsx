import { useState } from "react";
import { RefreshCcw, Search } from "lucide-react";
import { INITIAL_PAYMENTS, getLocalAdminData, setLocalAdminData } from "../data/mockAdminData";
import { toast } from "sonner";

export function AdminPayments() {
  const [payments, setPayments] = useState(() => getLocalAdminData("payments", INITIAL_PAYMENTS));
  const [search, setSearch] = useState("");

  const savePayments = (updated) => {
    setPayments(updated);
    setLocalAdminData("payments", updated);
  };

  const handleIssueRefund = (pay) => {
    if (confirm(`Issue full refund for transaction ${pay.id}?`)) {
      const updated = payments.map((p) =>
        p.id === pay.id ? { ...p, status: "Refunded" } : p
      );
      savePayments(updated);
      toast.success(`Refund of ${pay.amount} issued for ${pay.customer}`);
    }
  };

  const filtered = payments.filter(
    (p) =>
      p.id.toLowerCase().includes(search.toLowerCase()) ||
      p.orderId.toLowerCase().includes(search.toLowerCase()) ||
      p.customer.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Payment Transactions</h1>
          <p className="text-sm text-muted-foreground mt-1">
            View credit card and payment gateway settlement logs, processor fees, and process refunds.
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search payments by ID, order, or customer..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-4">Transaction ID</th>
                <th className="py-3.5">Order Ref</th>
                <th className="py-3.5">Customer</th>
                <th className="py-3.5">Payment Method</th>
                <th className="py-3.5">Gross Amount</th>
                <th className="py-3.5">Processor Fee</th>
                <th className="py-3.5">Status</th>
                <th className="py-3.5 pr-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((pay) => (
                <tr key={pay.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-4 font-mono font-bold text-foreground">{pay.id}</td>
                  <td className="py-3.5 font-mono text-muted-foreground">{pay.orderId}</td>
                  <td className="py-3.5 font-semibold text-foreground">{pay.customer}</td>
                  <td className="py-3.5 text-muted-foreground">{pay.gateway}</td>
                  <td className="py-3.5 font-bold text-foreground">{pay.amount}</td>
                  <td className="py-3.5 text-muted-foreground">{pay.fee}</td>
                  <td className="py-3.5">
                    <span
                      className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        pay.status === "Succeeded"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                          : "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-400"
                      }`}
                    >
                      {pay.status}
                    </span>
                  </td>
                  <td className="py-3.5 pr-4 text-right">
                    {pay.status === "Succeeded" && (
                      <button
                        onClick={() => handleIssueRefund(pay)}
                        className="px-2.5 py-1 text-[11px] font-semibold text-destructive bg-destructive/10 hover:bg-destructive/20 rounded-lg transition-colors inline-flex items-center gap-1"
                      >
                        <RefreshCcw className="size-3" />
                        <span>Issue Refund</span>
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
