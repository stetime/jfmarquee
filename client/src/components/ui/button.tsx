import * as React from "react";

import { cn } from "@/lib/utils";

type ButtonVariant = "default" | "outline" | "secondary" | "ghost" | "destructive" | "link";
type ButtonSize = "default" | "sm" | "xs" | "lg" | "icon" | "icon-xs";

const variants: Record<ButtonVariant, string> = {
	default: "bg-primary text-primary-foreground hover:bg-primary/80",
	outline:
		"border-border bg-background hover:bg-muted hover:text-foreground aria-expanded:bg-muted dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
	secondary: "bg-secondary text-secondary-foreground aria-expanded:bg-secondary",
	ghost: "hover:bg-muted hover:text-foreground aria-expanded:bg-muted dark:hover:bg-muted/50",
	destructive:
		"bg-destructive/10 text-destructive hover:bg-destructive/20 focus-visible:border-destructive/40 dark:bg-destructive/20",
	link: "text-primary underline-offset-4 hover:underline",
};

const sizes: Record<ButtonSize, string> = {
	default: "h-8 gap-1.5 px-2.5",
	xs: "h-6 gap-1 rounded-[min(var(--radius-md),10px)] px-2 text-xs [&_svg:not([class*='size-'])]:size-3",
	sm: "h-7 gap-1 rounded-[min(var(--radius-md),12px)] px-2.5 text-[0.8rem] [&_svg:not([class*='size-'])]:size-3.5",
	lg: "h-9 gap-1.5 px-2.5",
	icon: "size-8",
	"icon-xs": "size-6 rounded-[min(var(--radius-md),10px)] [&_svg:not([class*='size-'])]:size-3",
};

const base =
	"group/button relative inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4";

function Button({
	className,
	variant = "default",
	size = "default",
	type = "button",
	...props
}: React.ComponentProps<"button"> & { variant?: ButtonVariant; size?: ButtonSize }) {
	return (
		<button
			type={type}
			data-slot="button"
			className={cn(base, variants[variant], sizes[size], className)}
			{...props}
		/>
	);
}

export { Button };
