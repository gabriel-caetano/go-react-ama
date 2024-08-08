import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app";
import "./index.css";

function render() {
	const container = document.getElementById("root");
	if (!container) return;
	createRoot(container).render(
		<StrictMode>
			<App />
		</StrictMode>,
	);
}

render();
