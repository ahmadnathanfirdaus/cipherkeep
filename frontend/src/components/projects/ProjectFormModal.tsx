import { useEffect, useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { isNonEmpty } from "@/utils/validators";
import type { Project } from "@/types";

export interface ProjectFormValues {
  name: string;
  description: string;
}

interface ProjectFormModalProps {
  open: boolean;
  mode: "create" | "edit";
  initial?: Pick<Project, "name" | "description">;
  isSubmitting: boolean;
  onSubmit: (values: ProjectFormValues) => void;
  onClose: () => void;
}

export const ProjectFormModal = ({
  open,
  mode,
  initial,
  isSubmitting,
  onSubmit,
  onClose,
}: ProjectFormModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setName(initial?.name ?? "");
      setDescription(initial?.description ?? "");
      setError(null);
    }
  }, [open, initial]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (!isNonEmpty(name)) {
      setError("Project name is required.");
      return;
    }
    onSubmit({ name: name.trim(), description: description.trim() });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={mode === "create" ? "New project" : "Edit project"}
      description={
        mode === "create"
          ? "Create a project to group environments and secrets."
          : undefined
      }
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            type="submit"
            form="project-form"
            isLoading={isSubmitting}
          >
            {mode === "create" ? "Create project" : "Save changes"}
          </Button>
        </>
      }
    >
      <form
        id="project-form"
        className="flex flex-col gap-4"
        onSubmit={handleSubmit}
        noValidate
      >
        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="payments-api"
          error={error ?? undefined}
          autoFocus
          required
        />
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="project-description"
            className="text-sm font-medium text-neutral-700"
          >
            Description{" "}
            <span className="font-normal text-neutral-400">(optional)</span>
          </label>
          <textarea
            id="project-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            placeholder="What is this project for?"
            className="w-full resize-none rounded-md border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-900 shadow-sm placeholder:text-neutral-400 focus:border-accent-500 focus:outline-none focus:ring-2 focus:ring-accent-500"
          />
        </div>
      </form>
    </Modal>
  );
};
