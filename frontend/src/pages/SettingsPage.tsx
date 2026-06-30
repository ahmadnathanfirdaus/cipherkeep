import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { KeyRound, User as UserIcon } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";
import { useAuth } from "@/hooks/useAuth";
import { useToast } from "@/hooks/useToast";
import { authService } from "@/services/auth.service";
import { getApiErrorMessage } from "@/utils/apiError";

export const SettingsPage = () => {
  const { user } = useAuth();
  const { notify } = useToast();

  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);

  const changePassword = useMutation({
    mutationFn: () =>
      authService.changePassword({
        current_password: current,
        new_password: next,
      }),
    onSuccess: () => {
      notify({ variant: "success", title: "Password changed" });
      setCurrent("");
      setNext("");
      setConfirm("");
      setError(null);
    },
    onError: (err) =>
      notify({
        variant: "error",
        title: "Could not change password",
        description: getApiErrorMessage(err),
      }),
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (next.length < 8) {
      setError("New password must be at least 8 characters.");
      return;
    }
    if (next !== confirm) {
      setError("New password and confirmation do not match.");
      return;
    }
    setError(null);
    changePassword.mutate();
  };

  return (
    <div className="flex max-w-xl flex-col gap-6">
      <h1 className="text-xl font-semibold tracking-tight text-neutral-900">
        Settings
      </h1>

      <Card>
        <div className="flex flex-col gap-4 p-5">
          <div className="flex items-center gap-2 text-neutral-900">
            <UserIcon className="h-4 w-4 text-neutral-500" aria-hidden="true" />
            <span className="font-medium">Profile</span>
          </div>
          <Input label="Name" value={user?.name ?? ""} disabled />
          <Input label="Email" value={user?.email ?? ""} disabled />
        </div>
      </Card>

      <Card>
        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-4 p-5"
          noValidate
        >
          <div className="flex items-center gap-2 text-neutral-900">
            <KeyRound className="h-4 w-4 text-neutral-500" aria-hidden="true" />
            <span className="font-medium">Change password</span>
          </div>
          <Input
            label="Current password"
            type="password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            autoComplete="current-password"
            required
          />
          <Input
            label="New password"
            type="password"
            value={next}
            onChange={(e) => setNext(e.target.value)}
            hint="At least 8 characters."
            autoComplete="new-password"
            required
          />
          <Input
            label="Confirm new password"
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            error={error ?? undefined}
            autoComplete="new-password"
            required
          />
          <p className="text-xs text-neutral-500">
            Changing your password signs out your other sessions.
          </p>
          <div>
            <Button type="submit" isLoading={changePassword.isPending}>
              Update password
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
};
