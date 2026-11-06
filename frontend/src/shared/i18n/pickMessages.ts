export type MessageMap = { [key: string]: string | MessageMap };

/**
 * Messages needed by components that live above the route-group layouts:
 * header/footer, cart sidebar, wishlist state, product cards, and the role
 * badge in the account menu.
 */
export const BASE_MESSAGE_PATHS = [
	"common",
	"components",
	"cart",
	"wishlist",
	"products",
	"notFound",
	"home.product",
	"admin.roles",
] as const;

export const GROUP_MESSAGE_PATHS = {
	shop: ["home", "shop", "about", "reviews", "addresses", "checkout"],
	dashboard: [
		"addresses",
		"orders",
		"orderDetails",
		"overview",
		"settings",
		"userNav",
		"userReviews",
	],
	admin: ["admin"],
	vendor: ["vendor"],
	auth: ["auth"],
	portal: ["portal"],
	marketing: ["marketing"],
} as const;

export type MessageGroup = keyof typeof GROUP_MESSAGE_PATHS;

function getPath(
	source: MessageMap,
	parts: string[],
): string | MessageMap | undefined {
	let current: string | MessageMap | undefined = source;
	for (const part of parts) {
		if (typeof current !== "object" || current === null) return undefined;
		current = current[part];
	}
	return current;
}

function setPath(target: MessageMap, parts: string[], value: string | MessageMap) {
	let current = target;
	for (let index = 0; index < parts.length; index += 1) {
		const part = parts[index];
		if (index === parts.length - 1) {
			current[part] = value;
			return;
		}
		const next = current[part];
		if (typeof next !== "object" || next === null) {
			current[part] = {};
		}
		current = current[part] as MessageMap;
	}
}

/**
 * Copies only the requested dotted paths from the full message tree.
 * Source: https://next-intl.dev/docs/usage/configuration#nextintlclientprovider
 */
export function pickMessages(
	messages: MessageMap,
	paths: readonly string[],
): MessageMap {
	const selected: MessageMap = {};
	for (const path of paths) {
		const parts = path.split(".");
		const value = getPath(messages, parts);
		if (value !== undefined) setPath(selected, parts, value);
	}
	return selected;
}
