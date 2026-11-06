import type { ReactNode } from "react";

/**
 * Consistent page heading for admin and seller workspaces.
 *
 * Keep this intentionally plain: a title, an optional short description and a
 * trailing action group. Decorative icons and marketing-style subtitles do not
 * belong here.
 */
export function PageHeader({
	title,
	description,
	actions,
}: {
	title: string;
	description?: string;
	actions?: ReactNode;
}) {
	return (
		<div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
			<div className="min-w-0">
				<h1 className="text-xl font-semibold tracking-tight text-foreground">
					{title}
				</h1>
				{description ? (
					<p className="mt-1 max-w-2xl text-sm text-muted-foreground">
						{description}
					</p>
				) : null}
			</div>
			{actions ? (
				<div className="flex shrink-0 flex-wrap items-center gap-2">
					{actions}
				</div>
			) : null}
		</div>
	);
}
