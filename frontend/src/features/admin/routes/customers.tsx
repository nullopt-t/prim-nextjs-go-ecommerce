import { useState } from "react";
import { Search, Edit2, X } from "lucide-react";
import { INITIAL_CUSTOMERS, getLocalAdminData, setLocalAdminData } from "../data/mockAdminData";
import { toast } from "sonner";

export function AdminCustomers() {
  const [customers, setCustomers] = useState(() => getLocalAdminData("customers", INITIAL_CUSTOMERS));
  const [search, setSearch] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState(null);

  const [formData, setFormData] = useState({ segment: "Regular", status: "Active" });

  const saveCustomers = (updated) => {
    setCustomers(updated);
    setLocalAdminData("customers", updated);
  };

  const handleOpenEdit = (cust) => {
    setEditingCustomer(cust);
    setFormData({ segment: cust.segment, status: cust.status });
    setIsModalOpen(true);
  };

  const handleSave = (e) => {
    e.preventDefault();
    const updated = customers.map((c) =>
      c.id === editingCustomer.id ? { ...c, ...formData } : c
    );
    saveCustomers(updated);
    setIsModalOpen(false);
    toast.success(`Customer ${editingCustomer.name} updated!`);
  };

  const filtered = customers.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Customer Accounts</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage registered store users, VIP customer segmentation, and total lifetime spend.
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search customers by name or email..."
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
                <th className="py-3.5 pl-4">Customer Name</th>
                <th className="py-3.5">Contact Email</th>
                <th className="py-3.5">Total Orders</th>
                <th className="py-3.5">Lifetime Spend</th>
                <th className="py-3.5">Segment</th>
                <th className="py-3.5">Status</th>
                <th className="py-3.5 pr-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((cust) => (
                <tr key={cust.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-4">
                    <div className="font-semibold text-foreground">{cust.name}</div>
                    <div className="text-[10px] text-muted-foreground">Joined: {cust.joinedDate}</div>
                  </td>
                  <td className="py-3.5 text-muted-foreground">{cust.email}</td>
                  <td className="py-3.5 font-bold text-foreground">{cust.ordersCount} orders</td>
                  <td className="py-3.5 font-bold text-foreground">{cust.totalSpent}</td>
                  <td className="py-3.5">
                    <span
                      className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        cust.segment === "VIP"
                          ? "bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-400"
                          : "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-400"
                      }`}
                    >
                      {cust.segment}
                    </span>
                  </td>
                  <td className="py-3.5">
                    <span
                      className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        cust.status === "Active"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                          : "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-400"
                      }`}
                    >
                      {cust.status}
                    </span>
                  </td>
                  <td className="py-3.5 pr-4 text-right">
                    <button onClick={() => handleOpenEdit(cust)} className="p-1.5 text-muted-foreground hover:text-accent-brand">
                      <Edit2 className="size-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {isModalOpen && editingCustomer && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-sm w-full p-6 shadow-xl space-y-4">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">Edit Account: {editingCustomer.name}</h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>
            <form onSubmit={handleSave} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Customer Segment</label>
                <select
                  value={formData.segment}
                  onChange={(e) => setFormData({ ...formData, segment: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                >
                  <option value="New">New</option>
                  <option value="Regular">Regular</option>
                  <option value="VIP">VIP</option>
                </select>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Account Status</label>
                <select
                  value={formData.status}
                  onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                >
                  <option value="Active">Active</option>
                  <option value="Suspended">Suspended</option>
                </select>
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">Cancel</button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white rounded-xl">Save</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
