import { Outlet } from "react-router-dom";
import { Sidebar } from "@/layouts/Sidebar";
import { Topbar } from "@/layouts/Topbar";

export const AppLayout = () => (
  <div className="flex h-screen w-full overflow-hidden bg-neutral-50">
    <Sidebar />
    <div className="flex min-w-0 flex-1 flex-col">
      <Topbar />
      <main className="flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-6xl px-5 py-6 sm:px-8">
          <Outlet />
        </div>
      </main>
    </div>
  </div>
);
