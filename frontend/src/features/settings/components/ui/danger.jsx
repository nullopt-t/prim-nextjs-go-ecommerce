import { CustomButton } from "@/components/ui";

export default function DangerZone() {
	const dangerActions = [
		{
			id: "danger-logout",
			buttonText: "Logout",
		},
		{
			id: "danger-delete-acc",
			buttonText: "Delete Account",
		},
	];

	return (
		<div className="flex flex-col gap-2.5 border-t border-b border-destructive border-dashed pt-5 pb-5">
			{dangerActions.map((action) => (
				<div
					key={action.id}
					className="mr-auto bg-destructive text-destructive-foreground rounded-md hover:scale-90 w-fit"
				>
					<CustomButton text={action.buttonText} />
				</div>
			))}
		</div>
	);
}
