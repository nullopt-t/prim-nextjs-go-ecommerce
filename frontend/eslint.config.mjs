import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

// Next.js 16 removed `next lint`; ESLint is run directly (see
// node_modules/next/dist/docs/01-app/03-api-reference/05-config/03-eslint.md).
const eslintConfig = defineConfig([
	...nextVitals,
	...nextTs,
	{
		rules: {
			// Pre-existing debt. Keep these visible without failing the gate while
			// the migrations in AGENTS.md are in flight.
			"@next/next/no-img-element": "warn",
			"@typescript-eslint/no-explicit-any": "warn",
			"react/no-unescaped-entities": "off",
			// React Compiler-oriented rules. The existing effects genuinely fetch
			// data and a few components use render-time randomness for form
			// defaults. Keep them visible; they are tracked for the module split.
			"react-hooks/set-state-in-effect": "warn",
			"react-hooks/purity": "warn",
			"react-hooks/preserve-manual-memoization": "warn",
		},
	},
	globalIgnores([
		".next/**",
		"out/**",
		"build/**",
		"dist/**",
		".vite/**",
		"next-env.d.ts",
	]),
]);

export default eslintConfig;
