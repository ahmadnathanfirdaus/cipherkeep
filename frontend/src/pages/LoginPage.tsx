import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Lock, Mail } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useToast } from "@/hooks/useToast";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card, CardBody } from "@/components/ui/Card";
import { getApiErrorMessage } from "@/utils/apiError";
import { isValidEmail, isNonEmpty } from "@/utils/validators";

interface LocationState {
  from?: { pathname?: string };
}

export const LoginPage = () => {
  const { login } = useAuth();
  const { notify } = useToast();
  const navigate = useNavigate();
  const location = useLocation();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (
    event: React.FormEvent<HTMLFormElement>,
  ): Promise<void> => {
    event.preventDefault();
    setError(null);

    if (!isValidEmail(email) || !isNonEmpty(password)) {
      setError("Enter a valid email and password.");
      return;
    }

    setIsSubmitting(true);
    try {
      await login({ email: email.trim(), password });
      const state = location.state as LocationState | null;
      const redirectTo = state?.from?.pathname ?? "/projects";
      navigate(redirectTo, { replace: true });
    } catch (err) {
      const message = getApiErrorMessage(err);
      setError(message);
      notify({ variant: "error", title: "Sign in failed", description: message });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardBody>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => void handleSubmit(e)}
          noValidate
        >
          <Input
            label="Email"
            type="email"
            name="email"
            autoComplete="email"
            placeholder="you@example.com"
            leftIcon={<Mail className="h-4 w-4" />}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <Input
            label="Password"
            type="password"
            name="password"
            autoComplete="current-password"
            placeholder="••••••••"
            leftIcon={<Lock className="h-4 w-4" />}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={error ?? undefined}
            required
          />
          <Button type="submit" isLoading={isSubmitting} className="mt-1">
            Sign in
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-neutral-500">
          Don&apos;t have an account?{" "}
          <Link
            to="/register"
            className="font-medium text-accent-700 hover:text-accent-800"
          >
            Create one
          </Link>
        </p>
      </CardBody>
    </Card>
  );
};
