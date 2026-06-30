import { useEffect, useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { isNonEmpty, isValidSecretKey } from "@/utils/validators";

export interface SecretFormValues {
  key: string;
  value: string;
}

interface SecretFormModalProps {
  open: boolean;
  mode: "create" | "edit";
  /** Existing key when editing (the key field is locked in edit mode). */
  initialKey?: string;
  isSubmitting: boolean;
  onSubmit: (values: SecretFormValues) => void;
  onClose: () => void;
}

export const SecretFormModal = ({
  open,
  mode,
  initialKey,
  isSubmitting,
  onSubmit,
  onClose,
}: SecretFormModalProps) => {
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [valueError, setValueError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setKey(mode === "edit" ? (initialKey ?? "") : "");
      setValue("");
      setKeyError(null);
      setValueError(null);
    }
  }, [open, mode, initialKey]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    let valid = true;

    if (mode === "create" && !isValidSecretKey(key)) {
      setKeyError("Use UPPER_SNAKE_CASE (e.g. DATABASE_URL).");
      valid = false;
    } else {
      setKeyError(null);
    }

    if (!isNonEmpty(value)) {
      setValueError("Value is required.");
      valid = false;
    } else {
      setValueError(null);
    }

    if (!valid) return;
    onSubmit({ key: key.trim(), value });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={mode === "create" ? "New secret" : `Update "${initialKey ?? ""}"`}
      description={
        mode === "edit"
          ? "Saving creates a new version; previous versions are kept in history."
          : undefined
      }
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="secret-form" isLoading={isSubmitting}>
            {mode === "create" ? "Create secret" : "Save new version"}
          </Button>
        </>
      }
    >
      <form
        id="secret-form"
        className="flex flex-col gap-4"
        onSubmit={handleSubmit}
        noValidate
      >
        <Input
          label="Key"
          value={key}
          onChange={(e) => setKey(e.target.value.toUpperCase())}
          placeholder="DATABASE_URL"
          error={keyError ?? undefined}
          disabled={mode === "edit"}
          autoFocus={mode === "create"}
          className="font-mono"
          required
        />
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="secret-value"
            className="text-sm font-medium text-neutral-700"
          >
            Value
          </label>
          <textarea
            id="secret-value"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            rows={4}
            placeholder="Paste the secret value"
            autoFocus={mode === "edit"}
            className="w-full resize-none rounded-md border border-neutral-300 bg-white px-3 py-2 font-mono text-sm text-neutral-900 shadow-sm placeholder:text-neutral-400 focus:border-accent-500 focus:outline-none focus:ring-2 focus:ring-accent-500"
          />
          {valueError && (
            <p className="text-sm text-danger-600">{valueError}</p>
          )}
        </div>
      </form>
    </Modal>
  );
};
