import { useEffect, useRef, useState } from "react";
import { Upload } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Select } from "@/components/ui/Select";
import { readFileAsText } from "@/utils/download";
import type { ImportSecretsRequest, SecretFormat } from "@/types";

interface ImportSecretsModalProps {
  open: boolean;
  isSubmitting: boolean;
  onSubmit: (payload: ImportSecretsRequest) => void;
  onClose: () => void;
}

const FORMAT_OPTIONS = [
  { value: "env", label: ".env" },
  { value: "json", label: "JSON" },
  { value: "yaml", label: "YAML" },
];

const PLACEHOLDERS: Record<SecretFormat, string> = {
  env: "DATABASE_URL=postgres://...\nAPI_KEY=sk_live_...",
  json: '{\n  "database": {\n    "host": "localhost",\n    "port": 5432\n  }\n}',
  yaml: "database:\n  host: localhost\n  port: 5432",
};

const formatFromFilename = (name: string): SecretFormat | null => {
  const lower = name.toLowerCase();
  if (lower.endsWith(".json")) return "json";
  if (lower.endsWith(".yaml") || lower.endsWith(".yml")) return "yaml";
  if (lower.endsWith(".env") || lower.startsWith(".env")) return "env";
  return null;
};

export const ImportSecretsModal = ({
  open,
  isSubmitting,
  onSubmit,
  onClose,
}: ImportSecretsModalProps) => {
  const [format, setFormat] = useState<SecretFormat>("env");
  const [content, setContent] = useState("");
  const [overwrite, setOverwrite] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setFormat("env");
      setContent("");
      setOverwrite(true);
      setError(null);
    }
  }, [open]);

  const handleFileChange = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ): Promise<void> => {
    const file = event.target.files?.[0];
    if (!file) return;
    try {
      const text = await readFileAsText(file);
      setContent(text);
      const inferred = formatFromFilename(file.name);
      if (inferred) setFormat(inferred);
      setError(null);
    } catch {
      setError("Could not read the selected file.");
    } finally {
      // Allow re-selecting the same file later.
      event.target.value = "";
    }
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (content.trim() === "") {
      setError("Paste content or upload a file to import.");
      return;
    }
    onSubmit({ format, content, overwrite });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Import secrets"
      description="Bulk-add secrets from a .env, JSON, or YAML file. Nested JSON/YAML is flattened to dotted keys (database.host)."
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="import-form" isLoading={isSubmitting}>
            Import
          </Button>
        </>
      }
    >
      <form
        id="import-form"
        className="flex flex-col gap-4"
        onSubmit={(e) => void handleSubmit(e)}
        noValidate
      >
        <div className="flex items-end gap-3">
          <div className="w-40">
            <Select
              label="Format"
              value={format}
              onChange={(e) => setFormat(e.target.value as SecretFormat)}
              options={FORMAT_OPTIONS}
            />
          </div>
          <Button
            type="button"
            variant="secondary"
            leftIcon={<Upload className="h-4 w-4" />}
            onClick={() => fileInputRef.current?.click()}
          >
            Upload file
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".env,.json,.yaml,.yml,text/plain"
            className="hidden"
            onChange={(e) => void handleFileChange(e)}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="import-content"
            className="text-sm font-medium text-neutral-700"
          >
            Content
          </label>
          <textarea
            id="import-content"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={10}
            spellCheck={false}
            placeholder={PLACEHOLDERS[format]}
            className="w-full resize-y rounded-md border border-neutral-300 bg-white px-3 py-2 font-mono text-sm text-neutral-900 shadow-sm placeholder:text-neutral-400 focus:border-accent-500 focus:outline-none focus:ring-2 focus:ring-accent-500"
          />
          {error && <p className="text-sm text-danger-600">{error}</p>}
        </div>

        <label className="flex items-start gap-2.5 text-sm text-neutral-700">
          <input
            type="checkbox"
            checked={overwrite}
            onChange={(e) => setOverwrite(e.target.checked)}
            className="mt-0.5 h-4 w-4 rounded border-neutral-300 text-accent-600 focus:ring-accent-500"
          />
          <span>
            Overwrite existing keys
            <span className="block text-xs text-neutral-500">
              When on, existing keys are updated (a new version is saved). When
              off, they are skipped.
            </span>
          </span>
        </label>
      </form>
    </Modal>
  );
};
