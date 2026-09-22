export default function OverviewContent() {
  const boxes = [
    {
      id: "order-49",
      name: "Total orders",
      value: 12,
    },
    {
      id: "pending-49",
      name: "Pending",
      value: 2,
    },
    {
      id: "wishlist-49",
      name: "Wishlist items",
      value: 6,
    },
    {
      id: "or-49",
      name: "Reviews given",
      value: 5,
    },
  ];

  const tbHeaders = [
    { id: "row1-td1", td: "# Order" },
    { id: "row1-td2", td: "Date" },
    { id: "row1-td3", td: "Items" },
    { id: "row1-td4", td: "Total" },
    { id: "row1-td5", td: "Status" },
  ];

  return (
    <div className="w-full flex flex-col gap-6">
      <div className="grid gap-3 md:gap-5 grid-cols-2 md:grid-cols-4">
        {boxes.map((value) => (
          <div key={value.id} className="bg-card p-4 rounded-lg border border-border">
            <p className="text-muted-foreground text-xs md:text-sm">
              {value.name}
            </p>
            <p className="text-card-foreground font-bold text-title-sm md:text-title-md mt-1">
              {value.value}
            </p>
          </div>
        ))}
      </div>

      <div className="mt-4">
        <h2 className="mb-4 text-foreground font-medium text-title-sm">
          Recent Orders
        </h2>
        <div className="overflow-x-auto border border-border rounded-lg">
          <table className="w-full text-left">
            <thead className="bg-card border-b border-border">
              <tr>
                {tbHeaders.map((value) => (
                  <th
                    key={value.id}
                    className="p-3 text-xs font-medium text-muted-foreground uppercase"
                  >
                    {value.td}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              <tr className="text-txt-sm">
                <td className="p-3 text-foreground font-medium">PR-00482</td>
                <td className="p-3 text-muted-foreground">Jun 10, 2026</td>
                <td className="p-3 text-foreground">3</td>
                <td className="p-3 text-foreground font-medium">$124.48</td>
                <td className="p-3 text-accent-brand font-medium">Delivered</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
