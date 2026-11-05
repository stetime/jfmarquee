import { Circle, Clock3, Pause, Play, Sparkles, Tv } from "lucide-react";

import { cn } from "@/lib/utils";
import type { Network, Session, SessionKind, StreamMethod } from "@/lib/types";

/** Formats seconds as "42m" or "1h 12m". */
export function formatDuration(seconds: number): string {
	const total = Math.floor(seconds);
	const hours = Math.floor(total / 3600);
	const minutes = Math.floor((total % 3600) / 60);
	return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
}

const kindStyles: Record<SessionKind, { label: string; icon: typeof Play }> = {
	movie: { label: "Movie", icon: Play },
	episode: { label: "Episode", icon: Circle },
	track: { label: "Track", icon: Sparkles },
	live: { label: "Live", icon: Tv },
	other: { label: "Media", icon: Clock3 },
};

const networkLabel: Record<Network, string | null> = {
	lan: "LAN",
	tailscale: "Tailscale",
	wan: "WAN",
	unknown: null,
};

const acceleratorLabel: Record<string, string> = {
	qsv: "QSV",
	nvenc: "NVENC",
	vaapi: "VAAPI",
	amf: "AMF",
	videoToolbox: "VideoToolbox",
	videotoolbox: "VideoToolbox",
	v4l2m2m: "V4L2",
	rkmpp: "RKMPP",
};

const methodLabel: Record<StreamMethod, string | null> = {
	direct: null,
	remux: "Remux",
	transcode: "Transcode",
};

const osLabel: Record<Session["client"]["os"], string | null> = {
	windows: "Windows",
	macos: "macOS",
	linux: "Linux",
	android: "Android",
	ios: "iOS",
	tv: "TV",
	unknown: null,
};

function posterUrl(session: Session): string | null {
	if (!session.image) return null;
	const params = new URLSearchParams({ tag: session.image.tag });
	return `/api/thumb/${encodeURIComponent(session.image.id)}?${params}`;
}

function Badge({
	children,
	tone = "muted",
	title,
}: {
	children: React.ReactNode;
	tone?: "muted" | "accent";
	title?: string;
}) {
	return (
		<span
			title={title}
			className={cn(
				"inline-flex h-5 shrink-0 items-center gap-1 rounded-4xl border px-2 text-[11px] font-medium tabular-nums",
				tone === "muted" && "border-border/70 text-foreground/60",
				tone === "accent" && "border-primary/30 bg-primary/10 text-primary",
			)}
		>
			{children}
		</span>
	);
}

export default function SessionCard({ session }: { session: Session }) {
	const image = posterUrl(session);
	const KindIcon = kindStyles[session.kind].icon;
	const network = networkLabel[session.network];
	const hw = session.stream.hw ? (acceleratorLabel[session.stream.hw] ?? session.stream.hw) : null;
	const method = session.stream.method === "transcode" && hw ? hw : methodLabel[session.stream.method];
	const os = osLabel[session.client.os];
	const paused = session.state === "paused";

	return (
		<article
			className={cn(
				"group relative overflow-hidden rounded-lg border border-border/50 bg-card/30 transition-colors duration-200 hover:border-border hover:bg-card/50",
				paused && "opacity-80",
			)}
		>
			<div className="relative aspect-video overflow-hidden bg-muted/50">
				{image ? (
					<img
						src={image}
						alt=""
						loading="lazy"
						className={cn(
							"h-full w-full object-cover transition-transform duration-300 ease-out group-hover:scale-[1.02]",
							paused && "grayscale",
						)}
					/>
				) : (
					<div className="flex h-full w-full items-center justify-center text-muted-foreground">
						<KindIcon className="size-7 opacity-40" aria-hidden />
					</div>
				)}

				{paused && (
					<div className="absolute inset-0 flex items-center justify-center bg-background/40">
						<Pause className="size-7 fill-current text-foreground/90" aria-label="Paused" />
					</div>
				)}

				{session.progress !== undefined && session.progress > 0 && (
					<div className="absolute inset-x-0 bottom-0 h-0.5 bg-background/50">
						<div
							className="h-full bg-primary/90"
							style={{ width: `${Math.min(100, session.progress * 100)}%` }}
						/>
					</div>
				)}
			</div>

			<div className="flex flex-col gap-1 px-3 py-2.5">
				<div className="flex items-start justify-between gap-2">
					<h3 className="truncate text-sm font-semibold text-foreground" title={session.title}>
						{session.title}
					</h3>
					<div className="flex shrink-0 items-center gap-1">
						{method && (
							<Badge tone={session.stream.method === "direct" ? "muted" : "accent"}>{method}</Badge>
						)}
						{network && <Badge>{network}</Badge>}
					</div>
				</div>

				<p className="truncate text-xs text-foreground/55">
					{session.subtitle || kindStyles[session.kind].label}
				</p>

				<div className="flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-foreground/45">
					<span className="truncate">{session.client.name}</span>
					{session.client.device && session.client.device !== session.client.name && (
						<>
							<span className="text-border select-none">·</span>
							<span className="truncate">{session.client.device}</span>
						</>
					)}
					{os && (
						<>
							<span className="text-border select-none">·</span>
							<span>{os}</span>
						</>
					)}
					{session.duration !== undefined && session.duration > 0 && (
						<>
							<span className="text-border select-none">·</span>
							<span className="tabular-nums">{formatDuration(session.duration)}</span>
						</>
					)}
				</div>
			</div>
		</article>
	);
}
