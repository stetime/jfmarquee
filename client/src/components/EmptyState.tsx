export default function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <p className="text-lg font-medium text-foreground">
        Nothing playing right now
      </p>
      <p className="mt-1.5 max-w-[48ch] text-sm text-foreground/55">
        When someone starts streaming, it'll show up here.
      </p>
    </div>
  );
}
