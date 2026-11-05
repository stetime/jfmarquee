import { useTheme } from "./theme-provider";
import { Button } from "./ui/button";

export default function ThemeToggle() {
	const { theme, setTheme } = useTheme();
	const dark = theme !== "light";
	const next = dark ? "light" : "dark";

	return (
		<Button
			variant="ghost"
			size="icon"
			onClick={() => setTheme(next)}
			aria-label={`Switch to ${next} mode`}
			title={`Switch to ${next} mode`}
		>
			<span className="relative grid size-[18px] place-items-center">
				<svg
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					strokeWidth="1.75"
					strokeLinecap="round"
					className="absolute size-[18px] transition-all duration-200 dark:rotate-90 dark:scale-0 dark:opacity-0"
					aria-hidden
				>
					<circle cx="12" cy="12" r="4" />
					<path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
				</svg>
				<svg
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					strokeWidth="1.75"
					strokeLinecap="round"
					className="absolute size-[18px] -rotate-90 scale-0 opacity-0 transition-all duration-200 dark:rotate-0 dark:scale-100 dark:opacity-100"
					aria-hidden
				>
					<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8Z" />
				</svg>
			</span>
		</Button>
	);
}
