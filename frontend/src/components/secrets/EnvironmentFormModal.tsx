import { useEffect, useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { isNonEmpty } from "@/utils/validators";

interface EnvironmentFormModalProps {
  open: boolean;
  isSubmitting: boolean;
  onSubmit: (name: string) => void;
  onClose: () => void;
}

export const EnvironmentFormModal = ({
  open,
  isSubmitting,
  onSubmit,
  onClose,
}: EnvironmentFormModalProps) => {
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setName("");
      setError(null);
    }
  }, [open]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (!isNonEmpty(name)) {
      setError("Environment name is required.");
      return;
    }
    onSubmit(name.trim());
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="New environment"
      description="For example: development, staging, or production."
      size="sm"
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="environment-form" isLoading={isSubmitting}>
            Create environment
          </Button>
        </>
      }
    >
      <form id="environment-form" onSubmit={handleSubmit} noValidate>
        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="production"
          error={error ?? undefined}
          autoFocus
          required
        />
      </form>
    </Modal>
  );
};
