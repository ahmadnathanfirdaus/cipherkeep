import { useEffect, useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { isValidEmail } from "@/utils/validators";
import type { AddMemberRequest } from "@/types";

type AssignableRole = AddMemberRequest["role"];

interface AddMemberModalProps {
  open: boolean;
  isSubmitting: boolean;
  onSubmit: (payload: AddMemberRequest) => void;
  onClose: () => void;
}

const ROLE_OPTIONS = [
  { value: "member", label: "Member" },
  { value: "admin", label: "Admin" },
];

export const AddMemberModal = ({
  open,
  isSubmitting,
  onSubmit,
  onClose,
}: AddMemberModalProps) => {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AssignableRole>("member");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setEmail("");
      setRole("member");
      setError(null);
    }
  }, [open]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (!isValidEmail(email)) {
      setError("Enter a valid email address.");
      return;
    }
    onSubmit({ email: email.trim(), role });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Add member"
      description="Invite a teammate to this project by email."
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="member-form" isLoading={isSubmitting}>
            Add member
          </Button>
        </>
      }
    >
      <form
        id="member-form"
        className="flex flex-col gap-4"
        onSubmit={handleSubmit}
        noValidate
      >
        <Input
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="teammate@example.com"
          error={error ?? undefined}
          autoFocus
          required
        />
        <Select
          label="Role"
          options={ROLE_OPTIONS}
          value={role}
          onChange={(e) => setRole(e.target.value as AssignableRole)}
        />
      </form>
    </Modal>
  );
};
