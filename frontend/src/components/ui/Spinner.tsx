import { Loader2 } from "lucide-react";
import { cn } from "@/utils/cn";

export interface SpinnerProps {
  className?: string;
  label?: string;
}

export const Spinner = ({ className, label = "Loading" }: SpinnerProps) => (
  <span role="status" className="inline-flex items-center gap-2">
    <Loader2
      className={cn("h-5 w-5 animate-spin text-accent-600", className)}
      aria-hidden="true"
    />
    <span className="sr-only">{label}</span>
  </span>
);
