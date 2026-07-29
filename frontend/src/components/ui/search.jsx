import { Search, Mic } from "lucide-react";
import { useTranslation } from "react-i18next";

export function SearchBar() {
	const { t } = useTranslation("common");

	return (
		<div className="flex items-center relative bg-input-background rounded-lg border-2 border-border group focus-within:border-accent-brand">
			<Search className="absolute top-1/2 -translate-1/2 left-5 text-muted-foreground z-10 group-focus-within:text-accent-brand" />
			<input
				type="text"
				className="focus:text-accent-brand focus:placeholder:text-accent-brand w-full text-txt-sm md:text-txt-md lg:text-txt-lg text-black placeholder:text-muted-foreground p-2 pl-10"
				placeholder={t("header.search")}
			/>
			<Mic className="text-muted-foreground cursor-pointer hover:text-accent-brand" />
		</div>
	);
}
