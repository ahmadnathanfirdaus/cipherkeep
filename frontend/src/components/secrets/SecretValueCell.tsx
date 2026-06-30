import { useState } from "react";
import { Eye, EyeOff, Loader2 } from "lucide-react";
import { cn } from "@/utils/cn";
import { CopyButton } from "@/components/ui/CopyButton";
import { secretService } from "@/services/secret.service";
import { useToast } from "@/hooks/useToast";
import { getApiErrorMessage } from "@/utils/apiError";

interface SecretValueCellProps {
  environmentId: string;
  secretKey: string;
}

const MASK = "••••••••••••";

/**
 * Secret value is masked by default. On reveal we fetch the decrypted value
 * (which writes an audit log server-side) and keep it only in local state.
 */
export const SecretValueCell = ({
  environmentId,
  secretKey,
}: SecretValueCellProps) => {
  const { notify } = useToast();
  const [value, setValue] = useState<string | null>(null);
  const [revealed, setRevealed] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleToggle = async (): Promise<void> => {
    if (revealed) {
      setRevealed(false);
      return;
    }
    if (value !== null) {
      setRevealed(true);
      return;
    }
    setIsLoading(true);
    try {
      const secret = await secretService.reveal(environmentId, secretKey);
      setValue(secret.value);
      setRevealed(true);
    } catch (err) {
      notify({
        variant: "error",
        title: "Could not reveal secret",
        description: getApiErrorMessage(err),
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-1">
      <code
        className={cn(
          "min-w-0 flex-1 truncate rounded bg-neutral-50 px-2 py-1 font-mono text-xs",
          revealed ? "text-neutral-900" : "tracking-wider text-neutral-400",
        )}
      >
        {revealed && value !== null ? value : MASK}
      </code>
      <button
        type="button"
        onClick={() => void handleToggle()}
        aria-label={revealed ? "Hide value" : "Reveal value"}
        className="inline-flex h-9 w-9 items-center justify-center rounded-md text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-500"
      >
        {isLoading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : revealed ? (
          <EyeOff className="h-4 w-4" />
        ) : (
          <Eye className="h-4 w-4" />
        )}
      </button>
      {revealed && value !== null && (
        <CopyButton value={value} label="Copy value" />
      )}
    </div>
  );
};
