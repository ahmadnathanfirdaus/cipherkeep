import { AlertTriangle, RotateCw } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { getApiErrorMessage } from "@/utils/apiError";

export interface ErrorStateProps {
  title?: string;
  error: unknown;
  onRetry?: () => void;
}

export const ErrorState = ({
  title = "Something went wrong",
  error,
  onRetry,
}: ErrorStateProps) => (
  <div
    role="alert"
    className="flex flex-col items-center justify-center px-6 py-12 text-center"
  >
    <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-danger-50 text-danger-600">
      <AlertTriangle className="h-6 w-6" aria-hidden="true" />
    </div>
    <h3 className="text-sm font-semibold text-neutral-900">{title}</h3>
    <p className="mt-1 max-w-sm text-sm text-neutral-500">
      {getApiErrorMessage(error)}
    </p>
    {onRetry && (
      <Button
        variant="secondary"
        size="sm"
        className="mt-4"
        leftIcon={<RotateCw className="h-4 w-4" />}
        onClick={onRetry}
      >
        Try again
      </Button>
    )}
  </div>
);
