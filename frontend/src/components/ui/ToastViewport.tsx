import { CheckCircle2, Info, X, XCircle } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { createPortal } from "react-dom";
import { cn } from "@/utils/cn";
import type { Toast, ToastVariant } from "@/context/toast-context";

interface ToastViewportProps {
  toasts: Toast[];
  onDismiss: (id: string) => void;
}

const VARIANT_META: Record<
  ToastVariant,
  { icon: LucideIcon; iconClass: string; ring: string }
> = {
  success: {
    icon: CheckCircle2,
    iconClass: "text-success-600",
    ring: "ring-success-100",
  },
  error: { icon: XCircle, iconClass: "text-danger-600", ring: "ring-danger-100" },
  info: { icon: Info, iconClass: "text-accent-600", ring: "ring-accent-100" },
};

export const ToastViewport = ({ toasts, onDismiss }: ToastViewportProps) => {
  if (toasts.length === 0) return null;

  return createPortal(
    <div className="pointer-events-none fixed bottom-4 right-4 z-[60] flex w-full max-w-sm flex-col gap-2">
      {toasts.map((toast) => {
        const meta = VARIANT_META[toast.variant];
        const Icon = meta.icon;
        return (
          <div
            key={toast.id}
            role="status"
            className={cn(
              "pointer-events-auto flex items-start gap-3 rounded-lg border border-neutral-200 bg-white p-3 shadow-overlay ring-1 animate-scale-in",
              meta.ring,
            )}
          >
            <Icon
              className={cn("mt-0.5 h-5 w-5 shrink-0", meta.iconClass)}
              aria-hidden="true"
            />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-neutral-900">
                {toast.title}
              </p>
              {toast.description && (
                <p className="mt-0.5 break-words text-sm text-neutral-500">
                  {toast.description}
                </p>
              )}
            </div>
            <button
              type="button"
              aria-label="Dismiss"
              onClick={() => onDismiss(toast.id)}
              className="flex h-6 w-6 items-center justify-center rounded text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        );
      })}
    </div>,
    document.body,
  );
};
