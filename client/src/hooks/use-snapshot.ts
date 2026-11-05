import { useEffect, useRef, useState } from "react";

import { EMPTY_SNAPSHOT, type Snapshot } from "@/lib/types";

export type SSEStatus = "connecting" | "connected" | "disconnected";

/**
 * Subscribes to the server's SSE snapshot stream.
 *
 * EventSource handles reconnection on its own, so the only jobs here are
 * tracking status, replacing the snapshot wholesale, and closing on unmount.
 */
export function useSnapshot(url = "/api/sessions/stream") {
	const [snapshot, setSnapshot] = useState<Snapshot>(EMPTY_SNAPSHOT);
	const [status, setStatus] = useState<SSEStatus>("connecting");
	const sourceRef = useRef<EventSource | null>(null);

	useEffect(() => {
		const es = new EventSource(url);
		sourceRef.current = es;

		es.addEventListener("open", () => setStatus("connected"));

		es.addEventListener("snapshot", (event) => {
			try {
				setSnapshot(JSON.parse((event as MessageEvent<string>).data) as Snapshot);
				setStatus("connected");
			} catch {
				// A malformed frame is not worth tearing the stream down for.
			}
		});

		es.addEventListener("error", () => setStatus("disconnected"));

		return () => {
			es.close();
			sourceRef.current = null;
		};
	}, [url]);

	return { snapshot, status };
}
