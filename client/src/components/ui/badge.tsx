import * as React from "react";

import { cn } from "@/lib/utils";

type BadgeVariant = "default" | "secondary" | "destructive" | "outline" | "ghost" | "link";

const badgeVariants: Record<BadgeVariant, string> = {
	default: "bg-primary text-primary-foreground",
	secondary: "bg-secondary text-secondary-foreground",
	destructive:
		"bg-destructive/10 text-destructive dark:bg-destructive/20 dark:focus-visible:ring-destructive/40",
	outline: "border-border text-foreground",
	ghost: "hover:bg-muted hover:text-muted-foreground dark:hover:bg-muted/50",
	link: "text-primary underline-offset-4 hover:underline",
};

const base =
	"group/badge inline-flex h-5 w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-4xl border border-transparent px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-all focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 [&>svg]:pointer-events-none [&>svg]:size-3!";

function Badge({
	className,
	variant = "default",
	...props
}: React.ComponentProps<"span"> & { variant?: BadgeVariant }) {
	return (
		<span
			data-slot="badge"
			data-variant={variant}
			className={cn(base, badgeVariants[variant], className)}
			{...props}
		/>
	);
}

export { Badge };
