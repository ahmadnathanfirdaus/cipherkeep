import type { LucideIcon } from "lucide-react";

export interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export const EmptyState = ({
  icon: Icon,
  title,
  description,
  action,
}: EmptyStateProps) => (
  <div className="flex flex-col items-center justify-center px-6 py-12 text-center">
    <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-neutral-100 text-neutral-400">
      <Icon className="h-6 w-6" aria-hidden="true" />
    </div>
    <h3 className="text-sm font-semibold text-neutral-900">{title}</h3>
    {description && (
      <p className="mt-1 max-w-sm text-sm text-neutral-500">{description}</p>
    )}
    {action && <div className="mt-4">{action}</div>}
  </div>
);
