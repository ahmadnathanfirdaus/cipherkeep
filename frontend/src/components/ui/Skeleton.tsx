import { cn } from "@/utils/cn";

export interface SkeletonProps {
  className?: string;
}

export const Skeleton = ({ className }: SkeletonProps) => (
  <div
    className={cn(
      "relative overflow-hidden rounded-md bg-neutral-200/70",
      "before:absolute before:inset-0 before:-translate-x-full",
      "before:animate-[shimmer_1.5s_infinite] before:bg-gradient-to-r",
      "before:from-transparent before:via-white/40 before:to-transparent",
      className,
    )}
    aria-hidden="true"
  />
);

interface TableSkeletonProps {
  rows?: number;
  columns?: number;
}

export const TableSkeleton = ({
  rows = 5,
  columns = 4,
}: TableSkeletonProps) => (
  <div className="flex flex-col gap-3 p-4" aria-busy="true">
    {Array.from({ length: rows }).map((_, rowIdx) => (
      <div key={rowIdx} className="flex items-center gap-4">
        {Array.from({ length: columns }).map((__, colIdx) => (
          <Skeleton
            key={colIdx}
            className={cn("h-5", colIdx === 0 ? "w-1/3" : "flex-1")}
          />
        ))}
      </div>
    ))}
  </div>
);
