import { NavLink } from "react-router-dom";
import { KeyRound, KeySquare, ScrollText, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { cn } from "@/utils/cn";
import type { Role } from "@/types";

interface ProjectTabsProps {
  projectId: string;
  role: Role;
}

interface TabDef {
  to: string;
  label: string;
  icon: LucideIcon;
  end: boolean;
  adminOnly: boolean;
}

export const ProjectTabs = ({ projectId, role }: ProjectTabsProps) => {
  const tabs: TabDef[] = [
    {
      to: `/projects/${projectId}`,
      label: "Secrets",
      icon: KeyRound,
      end: true,
      adminOnly: false,
    },
    {
      to: `/projects/${projectId}/members`,
      label: "Members",
      icon: Users,
      end: false,
      adminOnly: false,
    },
    {
      to: `/projects/${projectId}/tokens`,
      label: "Tokens",
      icon: KeySquare,
      end: false,
      adminOnly: true,
    },
    {
      to: `/projects/${projectId}/audit`,
      label: "Audit log",
      icon: ScrollText,
      end: false,
      adminOnly: true,
    },
  ];

  const visible = tabs.filter(
    (tab) => !tab.adminOnly || role === "owner" || role === "admin",
  );

  return (
    <nav className="flex items-center gap-1 border-b border-neutral-200">
      {visible.map((tab) => {
        const Icon = tab.icon;
        return (
          <NavLink
            key={tab.to}
            to={tab.to}
            end={tab.end}
            className={({ isActive }) =>
              cn(
                "-mb-px flex items-center gap-2 border-b-2 px-3 py-2.5 text-sm font-medium transition-colors",
                isActive
                  ? "border-accent-600 text-accent-700"
                  : "border-transparent text-neutral-500 hover:text-neutral-900",
              )
            }
          >
            <Icon className="h-4 w-4" aria-hidden="true" />
            {tab.label}
          </NavLink>
        );
      })}
    </nav>
  );
};
