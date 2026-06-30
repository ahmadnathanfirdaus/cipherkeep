import { Outlet } from "react-router-dom";
import { Logo } from "@/components/Logo";

export const AuthLayout = () => (
  <div className="flex min-h-screen items-center justify-center bg-neutral-50 px-4 py-12">
    <div className="w-full max-w-sm">
      <div className="mb-8 flex flex-col items-center text-center">
        <Logo size={44} className="mb-3" />
        <h1 className="text-lg font-semibold">
          <span className="text-neutral-900">Cipher</span>
          <span className="text-accent-600">keep</span>
        </h1>
        <p className="mt-1 text-sm text-neutral-500">
          Secure secrets for your team.
        </p>
      </div>
      <Outlet />
    </div>
  </div>
);
