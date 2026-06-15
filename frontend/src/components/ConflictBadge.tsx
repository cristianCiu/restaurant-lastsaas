interface ConflictBadgeProps {
  count: number;
}

export default function ConflictBadge({ count }: ConflictBadgeProps) {
  if (count === 0) return null;

  return (
    <div className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full bg-amber-500/20 border border-amber-500/30">
      <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
      <span className="text-xs text-amber-300">
        {count} conflict{count !== 1 ? 's' : ''} — last-writer-wins applied
      </span>
    </div>
  );
}
