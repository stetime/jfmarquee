import { useEffect } from "react";

import { useSnapshot } from "@/hooks/use-snapshot";
import DashboardStats from "@/components/DashboardStats";
import EmptyState from "@/components/EmptyState";
import ErrorBanner from "@/components/ErrorBanner";
import SessionCard from "@/components/SessionCard";
import SessionGridSkeleton from "@/components/SessionGridSkeleton";
import ThemeToggle from "@/components/ThemeToggle";
import { ThemeProvider } from "@/components/theme-provider";

function relativeTime(ts: number): string {
	if (!ts) return "";
	const seconds = Math.max(0, Math.round((Date.now() - ts) / 1000));
	if (seconds < 10) return "just now";
	if (seconds < 60) return `${seconds}s ago`;
	const minutes = Math.round(seconds / 60);
	if (minutes < 60) return `${minutes}m ago`;
	return `${Math.round(minutes / 60)}h ago`;
}

function Dashboard() {
	const { snapshot, status } = useSnapshot();

	const sessions = snapshot.sessions;
	const hasSessions = sessions.length > 0;
	const connected = status === "connected";
	const disconnected = status === "disconnected";

	const loading = status === "connecting" && snapshot.updatedAt === 0;
	const empty = connected && snapshot.updatedAt > 0 && !hasSessions;

	// Keep the document title useful in a pinned tab.
	useEffect(() => {
		document.title = hasSessions ? `Marquee · ${sessions.length} streaming` : "Marquee";
	}, [hasSessions, sessions.length]);

	return (
		<div className="mx-auto min-h-screen max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
			<header className="mb-6 flex flex-col gap-3">
				<div className="flex items-center justify-between">
					<h1 className="text-lg font-semibold tracking-tight text-foreground">
						Marquee
						{snapshot.updatedAt > 0 && (
							<span className="ml-2 align-middle text-xs font-normal text-foreground/35 tabular-nums">
								{relativeTime(snapshot.updatedAt)}
							</span>
						)}
					</h1>
					<ThemeToggle />
				</div>

				{hasSessions ? (
					<DashboardStats
						sessions={sessions}
						servingKbps={snapshot.servingKbps}
						transcoding={snapshot.transcoding}
						paused={snapshot.paused}
						stale={disconnected}
					/>
				) : (
					<div className="h-5" />
				)}

				{disconnected && (
					<div className="-mx-1">
						<ErrorBanner />
					</div>
				)}
			</header>

			{loading && <SessionGridSkeleton />}
			{empty && <EmptyState />}

			{hasSessions && (
				<div className="grid animate-[fadeIn_0.3s_ease-out] grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
					{sessions.map((session) => (
						<SessionCard key={session.id} session={session} />
					))}
				</div>
			)}

			{hasSessions && (
				<footer className="mt-8 text-center text-xs text-muted-foreground/60">
					{connected ? "Live" : "Reconnecting…"}
				</footer>
			)}
		</div>
	);
}

export default function App() {
	return (
		<ThemeProvider defaultTheme="dark" storageKey="marquee-theme">
			<Dashboard />
		</ThemeProvider>
	);
}
