import { Component, type ErrorInfo, type ReactNode } from "react";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/Button";

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

/**
 * Catches unhandled render errors and shows a recoverable fallback instead of a
 * blank screen. React error boundaries must be class components.
 */
export class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error("Unhandled UI error:", error, info.componentStack);
  }

  private readonly handleReload = (): void => {
    window.location.reload();
  };

  render(): ReactNode {
    if (this.state.hasError) {
      return (
        <div className="flex h-screen flex-col items-center justify-center gap-4 bg-neutral-50 px-6 text-center">
          <AlertTriangle
            className="h-10 w-10 text-danger-500"
            aria-hidden="true"
          />
          <div className="flex flex-col gap-1">
            <h1 className="text-lg font-semibold text-neutral-900">
              Something went wrong
            </h1>
            <p className="text-sm text-neutral-500">
              An unexpected error occurred. Reloading usually fixes it.
            </p>
          </div>
          <Button onClick={this.handleReload}>Reload page</Button>
        </div>
      );
    }
    return this.props.children;
  }
}
