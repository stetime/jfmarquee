/** Shapes returned by the Go API. Keep in sync with server/internal/domain. */

export type Network = "lan" | "wan" | "tailscale" | "unknown";
export type SessionKind = "movie" | "episode" | "track" | "live" | "other";
export type PlayState = "playing" | "paused";
export type StreamMethod = "direct" | "remux" | "transcode";
export type OS = "windows" | "macos" | "linux" | "android" | "ios" | "tv" | "unknown";

export interface SessionImage {
	id: string;
	tag: string;
}

export interface SessionClient {
	name: string;
	device?: string;
	os: OS;
}

export interface SessionStream {
	method: StreamMethod;
	hw?: string;
	videoCodec?: string;
	audioCodec?: string;
	bitrate?: number;
}

export interface Session {
	id: string;
	user?: string;
	title: string;
	subtitle?: string;
	kind: SessionKind;
	year?: number;
	image?: SessionImage;
	state: PlayState;
	progress?: number;
	duration?: number;
	ip?: string;
	network: Network;
	client: SessionClient;
	stream: SessionStream;
}

export interface Snapshot {
	sessions: Session[];
	updatedAt: number;
	servingKbps: number;
	transcoding: number;
	paused: number;
}

export const EMPTY_SNAPSHOT: Snapshot = {
	sessions: [],
	updatedAt: 0,
	servingKbps: 0,
	transcoding: 0,
	paused: 0,
};
