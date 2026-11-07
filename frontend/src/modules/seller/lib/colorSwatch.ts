// Maps common color names (as category attributes usually define them) to
// swatch hex values. Unknown names resolve to null and render as text.

const COLOR_HEX: Record<string, string> = {
	black: "#171717",
	white: "#fafafa",
	silver: "#c0c0c0",
	gold: "#d4af37",
	bronze: "#cd7f32",
	red: "#dc2626",
	blue: "#2563eb",
	green: "#16a34a",
	yellow: "#eab308",
	orange: "#ea580c",
	purple: "#7c3aed",
	pink: "#ec4899",
	brown: "#92400e",
	beige: "#d6c7a8",
	cream: "#fff8e7",
	ivory: "#fffff0",
	navy: "#1e3a5f",
	teal: "#0d9488",
	turquoise: "#2dd4bf",
	maroon: "#7f1d1d",
	olive: "#6b7c2a",
	khaki: "#b7a642",
	lavender: "#c4b5fd",
	cyan: "#06b6d4",
	magenta: "#d946ef",
	platinum: "#d8dce3",
	graphite: "#454a52",
	charcoal: "#36454f",
	space: "#23272f",
	midnight: "#0f1e38",
	starlight: "#ece7dd",
	moonstone: "#b9b3a8",
	"space black": "#17181c",
	"space gray": "#8e8e93",
	"space grey": "#8e8e93",
	"dark metallic moon": "#3f4347",
	"platinum silver": "#cfd3d8",
	"midnight green": "#1b3a34",
	"midnight blue": "#101c30",
	"forest green": "#1e4620",
	"rose gold": "#e6b8a2",
	"rose": "#e11d48",
	"sky blue": "#7dd3fc",
	"royal blue": "#1d4ed8",
	"emerald": "#059669",
	"mint": "#a7f3d0",
	"peach": "#ffdab9",
	"coral": "#ff7f50",
	"burgundy": "#800020",
	"wine": "#722f37",
	"champagne": "#f4e5c3",
	"copper": "#b87333",
	"amber": "#f59e0b",
	"lime": "#84cc16",
	"indigo": "#4f46e5",
	"violet": "#8b5cf6",
	"fuchsia": "#c026d3",
	"gray": "#6b7280",
	"grey": "#6b7280",
	"light gray": "#d1d5db",
	"light grey": "#d1d5db",
	"dark gray": "#374151",
	"dark grey": "#374151",
	"charcoal gray": "#36454f",
	"charcoal grey": "#36454f",
	"off white": "#f5f5f4",
	"off-white": "#f5f5f4",
};

// Single-token keyword hints used to recognize color axes even when some
// values are unknown.
const COLOR_KEYWORDS = [
	"black", "white", "silver", "gold", "bronze", "red", "blue", "green",
	"yellow", "orange", "purple", "pink", "brown", "beige", "cream", "ivory",
	"navy", "teal", "turquoise", "maroon", "olive", "khaki", "lavender",
	"cyan", "magenta", "platinum", "graphite", "charcoal", "midnight",
	"starlight", "moonstone", "sky", "royal", "emerald", "mint", "peach",
	"coral", "burgundy", "wine", "champagne", "copper", "amber", "lime",
	"indigo", "violet", "fuchsia", "rose", "space", "forest",
];

const HEX_PATTERN = /^#(?:[0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/i;

export function colorHexFor(value: string): string | null {
	const normalized = value.trim().toLowerCase();
	if (HEX_PATTERN.test(normalized)) return normalized;
	if (COLOR_HEX[normalized]) return COLOR_HEX[normalized];
	// Multi-word names ("Platinum Silver", "Space Black") match on their
	// leading keyword when the full name is unknown.
	for (const word of normalized.split(/[\s-]+/)) {
		if (COLOR_HEX[word]) return COLOR_HEX[word];
	}
	return null;
}

export function isColorValue(value: string): boolean {
	return colorHexFor(value) !== null;
}

export function isColorAttribute(key: string, allowedValues?: string[]): boolean {
	if (/colou?r|shade|hue/i.test(key)) return true;
	if (!allowedValues?.length) return false;
	const known = allowedValues.filter(isColorValue).length;
	return known >= Math.ceil(allowedValues.length / 2);
}
