function OverviewContent() {
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
		{
			id: "row1-td1",
			td: "# Order",
		},
		{
			id: "row1-td2",
			td: "Date",
		},
		{
			id: "row1-td3",
			td: "Items",
		},
		{
			id: "row1-td4",
			td: "Total",
		},
		{
			id: "row1-td5",
			td: "Status",
		},
	];

	return (
		<div className="w-full flex flex-col gap-5">
			<div className="grid gap-2.5 md:gap-5 grid-cols-2 md:grid-cols-4">
				{boxes.map((value) => (
					<div key={value.id} className="bg-card">
						<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
							{value.name}
						</p>
						<p className="text-card-foreground font-medium text-title-sm md:text-title-md lg:text-title-lg">
							{value.value}
						</p>
					</div>
				))}
			</div>
			<div className="">
				<p className="mb-2.5 text-foreground font-medium text-title-sm md:text-title-md">
					Recent Orders
				</p>
				<table className="w-full">
					<thead className="">
						<tr className="border-b border-border">
							{tbHeaders.map((value) => (
								<td
									key={value.id}
									className="pt-2.5 pb-2.5 text-muted-foreground"
								>
									{value.td}
								</td>
							))}
						</tr>
					</thead>
					<tbody className="">
						<tr className="border-b border-border text-txt-sm md:text-txt-md lg:text-txt-lg">
							<td className="pt-2.5 pb-2.5 text-foreground">
								PR-00482
							</td>
							<td className="pt-2.5 pb-2.5 text-muted-foreground">
								Jun 10, 2026
							</td>
							<td className="pt-2.5 pb-2.5 text-foreground">3</td>
							<td className="pt-2.5 pb-2.5 text-foreground">
								$124.48
							</td>
							<td className="pt-2.5 pb-2.5 text-foreground">
								Delivered
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	);
}

export default OverviewContent;
