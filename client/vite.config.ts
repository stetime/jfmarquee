import path from "path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [react(), tailwindcss()],
	resolve: {
		alias: {
			"@": path.resolve(import.meta.dirname, "./src"),
		},
	},
	server: {
		proxy: {
			// Everything the Go server owns lives under /api.
			"/api": {
				target: "http://localhost:3000",
				changeOrigin: true,
			},
		},
	},
});
