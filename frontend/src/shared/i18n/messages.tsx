import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import type { ReactNode } from "react";
import {
	BASE_MESSAGE_PATHS,
	GROUP_MESSAGE_PATHS,
	pickMessages,
	type MessageGroup,
	type MessageMap,
} from "./pickMessages";

export async function getRouteMessages(
	group: MessageGroup | "base" = "base",
): Promise<MessageMap> {
	const allMessages = (await getMessages()) as unknown as MessageMap;
	const paths =
		group === "base"
			? BASE_MESSAGE_PATHS
			: [...BASE_MESSAGE_PATHS, ...GROUP_MESSAGE_PATHS[group]];
	return pickMessages(allMessages, paths);
}

export async function RouteMessages({
	group,
	children,
}: {
	group: MessageGroup;
	children: ReactNode;
}) {
	const messages = await getRouteMessages(group);
	return (
		<NextIntlClientProvider messages={messages}>
			{children}
		</NextIntlClientProvider>
	);
}
