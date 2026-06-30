import { useEffect, useState } from "react";
import { Download, ShieldAlert } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Select } from "@/components/ui/Select";
import { Spinner } from "@/components/ui/Spinner";
import { CopyButton } from "@/components/ui/CopyButton";
import { useExportSecrets } from "@/hooks/useSecrets";
import { useToast } from "@/hooks/useToast";
import { downloadTextFile } from "@/utils/download";
import { getApiErrorMessage } from "@/utils/apiError";
import type { SecretExport, SecretFormat } from "@/types";

interface ExportSecretsModalProps {
  open: boolean;
  environmentId: string;
  secretCount: number;
  onClose: () => void;
}

const FORMAT_OPTIONS = [
  { value: "env", label: ".env (flat)" },
  { value: "json", label: "JSON (nested)" },
  { value: "yaml", label: "YAML (nested)" },
];

export const ExportSecretsModal = ({
  open,
  environmentId,
  secretCount,
  onClose,
}: ExportSecretsModalProps) => {
  const { notify } = useToast();
  const exportSecrets = useExportSecrets(environmentId);
  const [format, setFormat] = useState<SecretFormat>("env");
  const [result, setResult] = useState<SecretExport | null>(null);

  // Fetch (and convert) the raw content whenever the modal opens or the format
  // changes. Switching format re-renders the same secrets in the chosen format.
  useEffect(() => {
    if (!open) {
      setResult(null);
      return;
    }
    let active = true;
    exportSecrets.mutate(format, {
      onSuccess: (res) => {
        if (active) setResult(res);
      },
      onError: (err) => {
        if (active) {
          notify({
            variant: "error",
            title: "Could not load export",
            description: getApiErrorMessage(err),
          });
        }
      },
    });
    return () => {
      active = false;
    };
    // exportSecrets/notify are stable enough; re-run only on these inputs.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, format, environmentId]);

  const handleDownload = (): void => {
    if (!result) return;
    downloadTextFile(result.filename, result.content);
    notify({ variant: "success", title: `Downloaded ${result.filename}` });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Export secrets"
      description={`${secretCount} secret${secretCount === 1 ? "" : "s"} in this environment. JSON and YAML nest dotted keys (a.b.c).`}
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
          <Button
            leftIcon={<Download className="h-4 w-4" />}
            disabled={!result || secretCount === 0}
            onClick={handleDownload}
          >
            Download
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex items-end justify-between gap-3">
          <div className="w-44">
            <Select
              label="Format"
              value={format}
              onChange={(e) => setFormat(e.target.value as SecretFormat)}
              options={FORMAT_OPTIONS}
            />
          </div>
          {result && <CopyButton value={result.content} label="Copy raw" />}
        </div>

        <div className="min-h-[10rem] overflow-hidden rounded-md border border-neutral-200 bg-neutral-900">
          {exportSecrets.isPending && !result ? (
            <div className="flex h-40 items-center justify-center">
              <Spinner label="Loading" />
            </div>
          ) : secretCount === 0 ? (
            <div className="flex h-40 items-center justify-center text-sm text-neutral-400">
              No secrets to export.
            </div>
          ) : (
            <pre className="max-h-80 overflow-auto p-3 text-xs leading-relaxed text-neutral-100">
              <code>{result?.content ?? ""}</code>
            </pre>
          )}
        </div>

        <div className="flex items-start gap-2.5 rounded-md bg-warning-50 p-3 text-sm text-warning-700">
          <ShieldAlert
            className="mt-0.5 h-4 w-4 shrink-0 text-warning-600"
            aria-hidden="true"
          />
          <p>
            This shows decrypted secret values in plain text. Copy or download
            only to a secure location.
          </p>
        </div>
      </div>
    </Modal>
  );
};
