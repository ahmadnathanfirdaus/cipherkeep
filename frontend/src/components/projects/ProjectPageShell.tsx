import { Link } from "react-router-dom";
import { ChevronRight } from "lucide-react";
import { ProjectTabs } from "@/components/projects/ProjectTabs";
import { RoleBadge } from "@/components/ui/RoleBadge";
import { ErrorState } from "@/components/ui/ErrorState";
import { Skeleton } from "@/components/ui/Skeleton";
import { useProject } from "@/hooks/useProjects";
import type { Project } from "@/types";

interface ProjectPageShellProps {
  projectId: string;
  /** Receives the loaded project so sub-pages get the caller's role/details. */
  children: (project: Project) => React.ReactNode;
}

/**
 * Shared chrome for the project sub-pages: breadcrumb, title, role badge, tabs.
 * Loads the project itself so each sub-page doesn't have to.
 */
export const ProjectPageShell = ({
  projectId,
  children,
}: ProjectPageShellProps) => {
  const { data: project, isLoading, isError, error, refetch } =
    useProject(projectId);

  return (
    <div className="flex flex-col gap-5">
      <nav
        aria-label="Breadcrumb"
        className="flex items-center gap-1 text-sm text-neutral-500"
      >
        <Link to="/projects" className="hover:text-neutral-900">
          Projects
        </Link>
        <ChevronRight className="h-4 w-4" aria-hidden="true" />
        <span className="font-medium text-neutral-700">
          {project?.name ?? "…"}
        </span>
      </nav>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          <Skeleton className="h-7 w-48" />
          <Skeleton className="h-9 w-full max-w-md" />
        </div>
      ) : isError || !project ? (
        <ErrorState error={error} onRetry={() => void refetch()} />
      ) : (
        <>
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-semibold tracking-tight text-neutral-900">
              {project.name}
            </h1>
            <RoleBadge role={project.role} />
          </div>
          {project.description && (
            <p className="-mt-3 text-sm text-neutral-500">
              {project.description}
            </p>
          )}
          <ProjectTabs projectId={projectId} role={project.role} />
          {children(project)}
        </>
      )}
    </div>
  );
};
