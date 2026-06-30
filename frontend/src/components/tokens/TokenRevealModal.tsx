import { ShieldAlert } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { CopyButton } from "@/components/ui/CopyButton";
import { env } from "@/lib/env";

interface TokenRevealModalProps {
  open: boolean;
  plaintext: string;
  environmentId: string;
  onClose: () => void;
}

export const TokenRevealModal = ({
  open,
  plaintext,
  environmentId,
  onClose,
}: TokenRevealModalProps) => {
  const usage = [
    `export SECRETS_API=${env.apiBaseUrl}`,
    `export SECRETS_ENV_ID=${environmentId}`,
    `export SECRETS_TOKEN=${plaintext}`,
    "",
    `curl -fsS "$SECRETS_API/environments/$SECRETS_ENV_ID/export?format=env" \\`,
    `  -H "Authorization: Bearer $SECRETS_TOKEN"`,
  ].join("\n");

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Copy your token now"
      description="This is the only time the token is shown. Store it securely."
      size="lg"
      footer={<Button onClick={onClose}>Done</Button>}
    >
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2 rounded-md border border-neutral-200 bg-neutral-50 p-2">
          <code className="flex-1 overflow-x-auto whitespace-nowrap px-1 font-mono text-sm text-neutral-900">
            {plaintext}
          </code>
          <CopyButton value={plaintext} label="Copy token" />
        </div>

        <div className="flex items-start gap-2.5 rounded-md bg-warning-50 p-3 text-sm text-warning-700">
          <ShieldAlert
            className="mt-0.5 h-4 w-4 shrink-0 text-warning-600"
            aria-hidden="true"
          />
          <p>
            Treat this token like a password. It grants read access to this
            environment's secrets. If it leaks, revoke it and create a new one.
          </p>
        </div>

        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-neutral-700">
              Use it from an app or CI
            </span>
            <CopyButton value={usage} label="Copy snippet" />
          </div>
          <pre className="overflow-x-auto rounded-md bg-neutral-900 p-3 text-xs leading-relaxed text-neutral-100">
            <code>{usage}</code>
          </pre>
        </div>
      </div>
    </Modal>
  );
};
