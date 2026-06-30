import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { cn } from "@/utils/cn";
import { copyToClipboard } from "@/utils/clipboard";
import { useToast } from "@/hooks/useToast";

export interface CopyButtonProps {
  value: string;
  label?: string;
  className?: string;
}

export const CopyButton = ({
  value,
  label = "Copy",
  className,
}: CopyButtonProps) => {
  const [copied, setCopied] = useState(false);
  const { notify } = useToast();

  const handleCopy = async (): Promise<void> => {
    const ok = await copyToClipboard(value);
    if (ok) {
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } else {
      notify({ variant: "error", title: "Could not copy to clipboard" });
    }
  };

  return (
    <button
      type="button"
      aria-label={copied ? "Copied" : label}
      onClick={() => void handleCopy()}
      className={cn(
        "inline-flex h-9 w-9 items-center justify-center rounded-md text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-500",
        className,
      )}
    >
      {copied ? (
        <Check className="h-4 w-4 text-success-600" />
      ) : (
        <Copy className="h-4 w-4" />
      )}
    </button>
  );
};
