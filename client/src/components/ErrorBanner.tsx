import { Badge } from "./ui/badge";

export default function ErrorBanner() {
	return (
		<div
			role="alert"
			className="flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-2.5 text-sm"
		>
			<span className="size-1.5 shrink-0 rounded-full bg-destructive" />
			<span className="text-destructive">Can't reach the server. Reconnecting…</span>
			<Badge variant="outline" className="ml-auto shrink-0">
				offline
			</Badge>
		</div>
	);
}
