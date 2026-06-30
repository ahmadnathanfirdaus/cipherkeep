import { NavLink } from "react-router-dom";
import { FolderOpen } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Logo } from "@/components/Logo";
import { cn } from "@/utils/cn";

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
}

const NAV_ITEMS: NavItem[] = [
  { to: "/projects", label: "Projects", icon: FolderOpen },
];

export const Sidebar = () => (
  <aside className="hidden w-60 shrink-0 flex-col border-r border-neutral-200 bg-white md:flex">
    <div className="flex h-14 items-center gap-2 border-b border-neutral-200 px-5">
      <Logo size={22} />
      <span className="text-sm font-semibold">
        <span className="text-neutral-900">Cipher</span>
        <span className="text-accent-600">keep</span>
      </span>
    </div>
    <nav className="flex flex-1 flex-col gap-1 p-3">
      {NAV_ITEMS.map((item) => {
        const Icon = item.icon;
        return (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                "flex h-10 items-center gap-2.5 rounded-md px-3 text-sm font-medium transition-colors",
                isActive
                  ? "bg-accent-50 text-accent-700"
                  : "text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900",
              )
            }
          >
            <Icon className="h-4 w-4" aria-hidden="true" />
            {item.label}
          </NavLink>
        );
      })}
    </nav>
  </aside>
);
