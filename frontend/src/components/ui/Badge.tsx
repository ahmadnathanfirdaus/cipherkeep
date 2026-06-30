import { cn } from "@/utils/cn";

type BadgeVariant = "neutral" | "accent" | "success" | "warning" | "danger";

export interface BadgeProps {
  variant?: BadgeVariant;
  children: React.ReactNode;
  className?: string;
}

const VARIANT_CLASSES: Record<BadgeVariant, string> = {
  neutral: "bg-neutral-100 text-neutral-700 ring-neutral-200",
  accent: "bg-accent-50 text-accent-700 ring-accent-200",
  success: "bg-success-50 text-success-700 ring-success-100",
  warning: "bg-warning-50 text-warning-700 ring-warning-100",
  danger: "bg-danger-50 text-danger-700 ring-danger-100",
};

export const Badge = ({
  variant = "neutral",
  className,
  children,
}: BadgeProps) => (
  <span
    className={cn(
      "inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset",
      VARIANT_CLASSES[variant],
      className,
    )}
  >
    {children}
  </span>
);
