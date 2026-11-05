function SkeletonCard() {
  return (
    <div className="overflow-hidden rounded-lg border border-border/50 bg-card/30">
      <div className="aspect-[16/9] animate-pulse bg-muted/50" />
      <div className="flex flex-col gap-2 px-3 py-2.5">
        <div className="flex items-start justify-between gap-2">
          <div className="h-4 w-2/3 animate-pulse rounded bg-muted/60" />
          <div className="h-4 w-10 shrink-0 animate-pulse rounded bg-muted/60" />
        </div>
        <div className="h-3 w-1/2 animate-pulse rounded bg-muted/50" />
        <div className="h-3 w-2/5 animate-pulse rounded bg-muted/50" />
      </div>
    </div>
  );
}

export default function SessionGridSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading sessions"
      className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
    >
      <SkeletonCard />
      <SkeletonCard />
      <SkeletonCard />
    </div>
  );
}
