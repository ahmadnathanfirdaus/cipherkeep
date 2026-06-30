import { History } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import { Badge } from "@/components/ui/Badge";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { Spinner } from "@/components/ui/Spinner";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/Table";
import { useSecretVersions } from "@/hooks/useSecrets";
import { formatDateTime } from "@/utils/format";

interface SecretVersionsModalProps {
  open: boolean;
  environmentId: string;
  secretKey: string | null;
  onClose: () => void;
}

export const SecretVersionsModal = ({
  open,
  environmentId,
  secretKey,
  onClose,
}: SecretVersionsModalProps) => {
  const { data, isLoading, isError, error, refetch } = useSecretVersions(
    open ? environmentId : undefined,
    open ? (secretKey ?? undefined) : undefined,
  );

  const latest = data && data.length > 0 ? data[0]?.version : undefined;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Version history"
      description={secretKey ?? undefined}
      size="lg"
    >
      {isLoading ? (
        <div className="flex justify-center py-8">
          <Spinner label="Loading versions" />
        </div>
      ) : isError ? (
        <ErrorState error={error} onRetry={() => void refetch()} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon={History}
          title="No version history"
          description="Updates to this secret will appear here."
        />
      ) : (
        <Table>
          <THead>
            <TR>
              <TH>Version</TH>
              <TH>Created</TH>
              <TH>By</TH>
            </TR>
          </THead>
          <TBody>
            {data.map((version) => (
              <TR key={version.version}>
                <TD>
                  <span className="inline-flex items-center gap-2">
                    <span className="font-mono text-neutral-900">
                      v{version.version}
                    </span>
                    {version.version === latest && (
                      <Badge variant="accent">Current</Badge>
                    )}
                  </span>
                </TD>
                <TD className="text-neutral-500">
                  {formatDateTime(version.created_at)}
                </TD>
                <TD className="text-neutral-500">{version.created_by}</TD>
              </TR>
            ))}
          </TBody>
        </Table>
      )}
    </Modal>
  );
};
