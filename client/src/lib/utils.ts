/**
 * Joins conditional class names.
 *
 * This is a deliberately minimal replacement for clsx + tailwind-merge: the
 * components here have at most a handful of variants and no third-party class
 * overrides, so last-wins conflict resolution is not needed. Do not rely on a
 * later class beating an earlier conflicting one (`p-2 p-4` is undefined).
 */
export function cn(...inputs: (string | false | null | undefined)[]): string {
	return inputs.filter(Boolean).join(" ");
}
