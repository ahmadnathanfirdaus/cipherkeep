import { Plus, Trash2 } from "lucide-react";
import { cn } from "@/utils/cn";
import { Button } from "@/components/ui/Button";
import type { Environment, Role } from "@/types";

interface EnvironmentSwitcherProps {
  environments: Environment[];
  activeId: string | null;
  role: Role;
  onSelect: (environmentId: string) => void;
  onCreate: () => void;
  onDelete: (environmentId: string) => void;
}

export const EnvironmentSwitcher = ({
  environments,
  activeId,
  role,
  onSelect,
  onCreate,
  onDelete,
}: EnvironmentSwitcherProps) => (
  <div className="flex items-center gap-2 overflow-x-auto">
    <div
      role="tablist"
      aria-label="Environments"
      className="flex items-center gap-1 rounded-lg bg-neutral-100 p-1"
    >
      {environments.map((env) => {
        const isActive = env.id === activeId;
        return (
          <button
            key={env.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            onClick={() => onSelect(env.id)}
            className={cn(
              "h-8 rounded-md px-3 text-sm font-medium transition-colors",
              isActive
                ? "bg-white text-neutral-900 shadow-sm"
                : "text-neutral-600 hover:text-neutral-900",
            )}
          >
            {env.name}
          </button>
        );
      })}
    </div>
    {role !== "member" && (
      <Button
        variant="ghost"
        size="sm"
        leftIcon={<Plus className="h-4 w-4" />}
        onClick={onCreate}
      >
        Environment
      </Button>
    )}
    {role !== "member" && activeId !== null && (
      <Button
        variant="ghost"
        size="sm"
        aria-label="Delete environment"
        title="Delete environment"
        className="text-neutral-500 hover:bg-danger-50 hover:text-danger-600"
        onClick={() => onDelete(activeId)}
      >
        <Trash2 className="h-4 w-4" />
      </Button>
    )}
  </div>
);
