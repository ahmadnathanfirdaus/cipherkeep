import { useEffect, useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { isNonEmpty } from "@/utils/validators";
import type { CreateServiceTokenRequest, Environment } from "@/types";

interface CreateTokenModalProps {
  open: boolean;
  environments: Environment[];
  isSubmitting: boolean;
  onSubmit: (payload: CreateServiceTokenRequest) => void;
  onClose: () => void;
}

// Maps a UI choice to expires_in_days (0 = never).
const EXPIRY_OPTIONS = [
  { value: "30", label: "30 days" },
  { value: "90", label: "90 days" },
  { value: "365", label: "1 year" },
  { value: "0", label: "Never" },
];

export const CreateTokenModal = ({
  open,
  environments,
  isSubmitting,
  onSubmit,
  onClose,
}: CreateTokenModalProps) => {
  const [name, setName] = useState("");
  const [environmentId, setEnvironmentId] = useState("");
  const [expiry, setExpiry] = useState("90");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setName("");
      setEnvironmentId(environments[0]?.id ?? "");
      setExpiry("90");
      setError(null);
    }
  }, [open, environments]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (!isNonEmpty(name)) {
      setError("Token name is required.");
      return;
    }
    if (!environmentId) {
      setError("Select an environment.");
      return;
    }
    onSubmit({
      name: name.trim(),
      environment_id: environmentId,
      expires_in_days: Number(expiry),
    });
  };

  const envOptions = environments.map((env) => ({
    value: env.id,
    label: env.name,
  }));

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create service token"
      description="A read-only API key scoped to one environment, for apps and CI."
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="create-token-form" isLoading={isSubmitting}>
            Create token
          </Button>
        </>
      }
    >
      <form
        id="create-token-form"
        className="flex flex-col gap-4"
        onSubmit={handleSubmit}
        noValidate
      >
        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="ci-backend-prod"
          autoFocus
          required
        />
        {environments.length === 0 ? (
          <p className="text-sm text-danger-600">
            Create an environment first; tokens are scoped to one environment.
          </p>
        ) : (
          <Select
            label="Environment"
            value={environmentId}
            onChange={(e) => setEnvironmentId(e.target.value)}
            options={envOptions}
          />
        )}
        <Select
          label="Expires"
          value={expiry}
          onChange={(e) => setExpiry(e.target.value)}
          options={EXPIRY_OPTIONS}
        />
        {error && <p className="text-sm text-danger-600">{error}</p>}
      </form>
    </Modal>
  );
};
