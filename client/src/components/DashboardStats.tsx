import { cn } from "@/lib/utils";
import type { Session, SessionKind } from "@/lib/types";

export function bitrateLabel(kbps: number): string {
	if (kbps <= 0) return "0 kbps";
	if (kbps >= 1_000_000) return `${(kbps / 1_000_000).toFixed(1)} Gbps`;
	if (kbps >= 1000) return `${(kbps / 1000).toFixed(1)} Mbps`;
	return `${Math.round(kbps)} kbps`;
}

const kindOrder: SessionKind[] = ["movie", "episode", "track", "live", "other"];
const kindName: Record<SessionKind, [string, string]> = {
	movie: ["movie", "movies"],
	episode: ["episode", "episodes"],
	track: ["track", "tracks"],
	live: ["live", "live"],
	other: ["item", "items"],
};

export default function DashboardStats({
	sessions,
	servingKbps,
	transcoding,
	paused,
	stale,
}: {
	sessions: Session[];
	servingKbps: number;
	transcoding: number;
	paused: number;
	stale: boolean;
}) {
	const counts = sessions.reduce<Partial<Record<SessionKind, number>>>((acc, s) => {
		acc[s.kind] = (acc[s.kind] ?? 0) + 1;
		return acc;
	}, {});

	return (
		<div className="flex flex-wrap items-baseline gap-x-4 gap-y-1 text-sm">
			<span className="flex items-center gap-2 font-semibold text-foreground tabular-nums">
				<span
					className={cn(
						"size-1.5 rounded-full",
						stale ? "bg-amber-500" : "animate-pulse bg-primary",
					)}
					aria-hidden
				/>
				{sessions.length} streaming
			</span>

			{kindOrder.map((kind) => {
				const n = counts[kind];
				if (!n) return null;
				const [one, many] = kindName[kind];
				return (
					<span key={kind} className="text-foreground/55 tabular-nums">
						{n} {n === 1 ? one : many}
					</span>
				);
			})}

			{servingKbps > 0 && (
				<span className="text-foreground/40 tabular-nums">
					Serving <span className="text-foreground/70">{bitrateLabel(servingKbps)}</span>
				</span>
			)}

			{transcoding > 0 && (
				<span className="text-foreground/40 tabular-nums">
					<span className="text-foreground/70">{transcoding}</span> transcoding
				</span>
			)}

			{paused > 0 && (
				<span className="text-foreground/40 tabular-nums">
					<span className="text-foreground/70">{paused}</span> paused
				</span>
			)}
		</div>
	);
}
