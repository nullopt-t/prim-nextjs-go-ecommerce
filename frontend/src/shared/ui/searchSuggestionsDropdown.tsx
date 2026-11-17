"use client";

import { useEffect, useMemo, useRef, useState, type RefObject } from "react";
import { ArrowRight, Clock, SearchX } from "lucide-react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/shared/i18n/navigation";
import { cn } from "@/shared/lib/cn";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { addRecentSearch, clearRecentSearches, getRecentSearches } from "@/shared/lib/recentSearches";
import { useSearchSuggestions } from "@/shared/lib/useSearchSuggestions";
import { CatalogImage } from "@/shared/ui/catalogImage";

type Item =
	| { kind: "recent"; term: string }
	| { kind: "product"; id: string; title: string; thumbnail: string }
	| { kind: "all" };

interface Props {
	query: string;
	/** Wraps the input and the dropdown so clicking the input does not close it. */
	containerRef: RefObject<HTMLElement | null>;
	onClose: () => void;
	onRecentSelect: (term: string) => void;
	className?: string;
}

export function SearchSuggestionsDropdown({ query, containerRef, onClose, onRecentSelect, className }: Props) {
	const t = useTranslations("common.header");
	const router = useRouter();
	const dropdownRef = useRef<HTMLDivElement>(null);
	const term = query.trim();
	const showRecents = term.length < 2;
	const { suggestions, loading, error } = useSearchSuggestions(query);
	const [recents, setRecents] = useState<string[]>(() => getRecentSearches());
	const [active, setActive] = useState({ term, index: -1 });
	const activeIndex = active.term === term ? active.index : -1;
	const setActiveIndex = (index: number) => setActive({ term, index });

	const items = useMemo<Item[]>(() => {
		if (showRecents) return recents.map((recent) => ({ kind: "recent", term: recent }));
		const products: Item[] = suggestions.map((item) => ({ kind: "product", ...item }));
		return error ? products : [...products, { kind: "all" }];
	}, [showRecents, recents, suggestions, error]);

	useEffect(() => {
		const handlePointerDown = (event: PointerEvent) => {
			if (!containerRef.current?.contains(event.target as Node)) onClose();
		};
		document.addEventListener("pointerdown", handlePointerDown);
		return () => document.removeEventListener("pointerdown", handlePointerDown);
	}, [containerRef, onClose]);

	const activate = (item: Item) => {
		if (item.kind === "recent") {
			addRecentSearch(item.term);
			onRecentSelect(item.term);
			router.push(`/products?q=${encodeURIComponent(item.term)}`);
		} else if (item.kind === "product") {
			addRecentSearch(term);
			router.push(`/products/${item.id}`);
		} else {
			addRecentSearch(term);
			router.push(`/products?q=${encodeURIComponent(term)}`);
		}
		onClose();
	};

	useEffect(() => {
		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				event.stopPropagation();
				onClose();
				return;
			}
			if (!items.length) return;
			if (event.key === "ArrowDown" || event.key === "ArrowUp") {
				event.preventDefault();
				const step = event.key === "ArrowDown" ? 1 : -1;
				setActiveIndex((activeIndex + step + items.length) % items.length);
			} else if (event.key === "Enter" && activeIndex >= 0) {
				event.preventDefault();
				activate(items[activeIndex]);
			}
		};
		document.addEventListener("keydown", handleKeyDown);
		return () => document.removeEventListener("keydown", handleKeyDown);
	});

	const handleClearRecents = () => {
		clearRecentSearches();
		setRecents([]);
	};

	if (showRecents && recents.length === 0) return null;

	return (
		<div
			ref={dropdownRef}
			role="listbox"
			aria-label={t("searchProducts")}
			className={cn(
				"absolute top-full left-0 right-0 z-50 mt-2 overflow-hidden rounded-xl border border-border/80 bg-background shadow-xl",
				className
			)}
		>
			{showRecents ? (
				<>
					<div className="flex items-center justify-between gap-2 px-3 py-2 text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
						<span className="flex items-center gap-1.5">
							<Clock className="size-3.5" aria-hidden />
							{t("recentSearches")}
						</span>
						<button
							type="button"
							onClick={handleClearRecents}
							className="text-[11px] font-medium normal-case tracking-normal text-accent-brand hover:underline cursor-pointer"
						>
							{t("clearRecent")}
						</button>
					</div>
					{items.map((item, index) =>
						item.kind === "recent" ? (
							<button
								key={item.term}
								type="button"
								role="option"
								aria-selected={activeIndex === index}
								onMouseEnter={() => setActiveIndex(index)}
								onClick={() => activate(item)}
								className={cn(
									"flex w-full items-center gap-2.5 px-3 py-2 text-start text-xs text-foreground hover:bg-secondary/70 cursor-pointer",
									activeIndex === index && "bg-secondary/70"
								)}
							>
								<Clock className="size-3.5 shrink-0 text-muted-foreground" aria-hidden />
								<span className="truncate">{item.term}</span>
							</button>
						) : null
					)}
				</>
			) : loading ? (
				<div className="space-y-2 p-3" aria-busy="true" aria-label={t("suggestionsLoading")}>
					{Array.from({ length: 3 }).map((_, index) => (
						<div key={index} className="flex items-center gap-2.5">
							<div className="size-9 shrink-0 animate-pulse rounded-md bg-muted" />
							<div className="h-3 flex-1 animate-pulse rounded bg-muted" />
						</div>
					))}
				</div>
			) : (
				<>
					{error ? (
						<p role="alert" className="px-3 py-4 text-center text-xs text-muted-foreground">
							{t("suggestionsError")}
						</p>
					) : suggestions.length === 0 ? (
						<p role="status" className="flex items-center justify-center gap-2 px-3 py-4 text-xs text-muted-foreground">
							<SearchX className="size-4" aria-hidden />
							{t("noResults")}
						</p>
					) : (
						<ul>
							{suggestions.map((product, index) => (
								<li key={product.id}>
									<button
										type="button"
										role="option"
										aria-selected={activeIndex === index}
										onMouseEnter={() => setActiveIndex(index)}
										onClick={() => activate({ kind: "product", ...product })}
										className={cn(
											"flex w-full items-center gap-2.5 px-3 py-2 text-start hover:bg-secondary/70 cursor-pointer",
											activeIndex === index && "bg-secondary/70"
										)}
									>
										<CatalogImage
											src={normalizeMediaUrl(product.thumbnail)}
											alt=""
											width={36}
											height={36}
											className="size-9 shrink-0 rounded-md object-cover"
										/>
										<span className="line-clamp-2 text-xs font-medium text-foreground">{product.title}</span>
									</button>
								</li>
							))}
						</ul>
					)}
					<button
						type="button"
						role="option"
						aria-selected={activeIndex === items.length - 1}
						onMouseEnter={() => setActiveIndex(items.length - 1)}
						onClick={() => activate({ kind: "all" })}
						className="flex w-full items-center justify-between gap-2 border-t border-border/70 px-3 py-2.5 text-start text-xs font-semibold text-accent-brand hover:bg-secondary/70 cursor-pointer"
					>
						<span>{t("viewAllResults", { query: term })}</span>
						<ArrowRight className="size-3.5 shrink-0 rtl:rotate-180" aria-hidden />
					</button>
				</>
			)}
		</div>
	);
}
