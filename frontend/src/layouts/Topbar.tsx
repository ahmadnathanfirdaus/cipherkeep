import { useState } from "react";
import { Link } from "react-router-dom";
import { LogOut, Settings, User as UserIcon } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/Button";

export const Topbar = () => {
  const { user, logout } = useAuth();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const handleLogout = async (): Promise<void> => {
    setIsLoggingOut(true);
    try {
      await logout();
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <header className="flex h-14 items-center justify-end gap-4 border-b border-neutral-200 bg-white px-5">
      {user && (
        <Link
          to="/settings"
          className="flex items-center gap-2 rounded-md px-2 py-1 text-sm transition-colors hover:bg-neutral-100"
          title="Settings"
        >
          <span className="flex h-8 w-8 items-center justify-center rounded-full bg-neutral-100 text-neutral-500">
            <UserIcon className="h-4 w-4" aria-hidden="true" />
          </span>
          <div className="hidden flex-col leading-tight sm:flex">
            <span className="font-medium text-neutral-900">{user.name}</span>
            <span className="text-xs text-neutral-500">{user.email}</span>
          </div>
          <Settings
            className="h-4 w-4 text-neutral-400 sm:ml-1"
            aria-hidden="true"
          />
        </Link>
      )}
      <Button
        variant="ghost"
        size="sm"
        leftIcon={<LogOut className="h-4 w-4" />}
        onClick={() => void handleLogout()}
        isLoading={isLoggingOut}
      >
        Sign out
      </Button>
    </header>
  );
};
