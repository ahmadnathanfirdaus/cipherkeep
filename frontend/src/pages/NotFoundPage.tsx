import { Link } from "react-router-dom";
import { Button } from "@/components/ui/Button";

export const NotFoundPage = () => (
  <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-neutral-50 px-4 text-center">
    <p className="text-sm font-semibold text-accent-700">404</p>
    <h1 className="text-2xl font-semibold text-neutral-900">Page not found</h1>
    <p className="max-w-sm text-sm text-neutral-500">
      The page you&apos;re looking for doesn&apos;t exist or has been moved.
    </p>
    <Link to="/projects">
      <Button>Back to projects</Button>
    </Link>
  </div>
);
