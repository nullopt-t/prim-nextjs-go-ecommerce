import "@/app/styles/index.css";
import "@/app/i18n/config";
import App from "@/app/App";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import Provider from "@/app/providers/provider";

createRoot(document.getElementById("root")).render(
	<StrictMode>
		<Provider>
			<App />
		</Provider>
	</StrictMode>,
);
