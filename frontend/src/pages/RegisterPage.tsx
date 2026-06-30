import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Lock, Mail, User as UserIcon } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useToast } from "@/hooks/useToast";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card, CardBody } from "@/components/ui/Card";
import { getApiErrorMessage } from "@/utils/apiError";
import { isNonEmpty, isValidEmail } from "@/utils/validators";

const MIN_PASSWORD_LENGTH = 8;

export const RegisterPage = () => {
  const { register, login } = useAuth();
  const { notify } = useToast();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (
    event: React.FormEvent<HTMLFormElement>,
  ): Promise<void> => {
    event.preventDefault();
    setError(null);

    if (!isNonEmpty(name)) {
      setError("Please enter your name.");
      return;
    }
    if (!isValidEmail(email)) {
      setError("Please enter a valid email address.");
      return;
    }
    if (password.length < MIN_PASSWORD_LENGTH) {
      setError(`Password must be at least ${MIN_PASSWORD_LENGTH} characters.`);
      return;
    }

    setIsSubmitting(true);
    try {
      await register({ name: name.trim(), email: email.trim(), password });
      // Convenience: log in immediately after registering.
      await login({ email: email.trim(), password });
      notify({ variant: "success", title: "Account created" });
      navigate("/projects", { replace: true });
    } catch (err) {
      const message = getApiErrorMessage(err);
      setError(message);
      notify({
        variant: "error",
        title: "Registration failed",
        description: message,
      });
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
            label="Name"
            type="text"
            name="name"
            autoComplete="name"
            placeholder="Ada Lovelace"
            leftIcon={<UserIcon className="h-4 w-4" />}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
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
            autoComplete="new-password"
            placeholder="At least 8 characters"
            leftIcon={<Lock className="h-4 w-4" />}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={error ?? undefined}
            required
          />
          <Button type="submit" isLoading={isSubmitting} className="mt-1">
            Create account
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-neutral-500">
          Already have an account?{" "}
          <Link
            to="/login"
            className="font-medium text-accent-700 hover:text-accent-800"
          >
            Sign in
          </Link>
        </p>
      </CardBody>
    </Card>
  );
};
