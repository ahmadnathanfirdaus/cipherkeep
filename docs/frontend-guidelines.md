# Frontend Guidelines

React + Vite + TypeScript (strict), Tailwind CSS, lucide-react icons. The goal is a
**clean, professional, intentional** interface — not generic "AI slop".

## Hard rules

- **TypeScript strict, no `any`.** `tsconfig` has `"strict": true`, `noImplicitAny`,
  `noUncheckedIndexedAccess`. If a type is truly unknown, use `unknown` + narrowing,
  never `any`. ESLint rule `@typescript-eslint/no-explicit-any: error`.
- **Arrow functions everywhere** — components, handlers, helpers. No `function`
  declarations for components.
- **Component props via `interface`** (not inline types, not `type` alias):
  ```tsx
  interface ButtonProps {
    variant?: "primary" | "secondary" | "ghost" | "danger";
    isLoading?: boolean;
    onClick?: () => void;
    children: React.ReactNode;
  }

  const Button = ({ variant = "primary", isLoading, ...props }: ButtonProps) => {
    // ...
  };
  ```
- **Real icons only — lucide-react.** Never use emoji as UI icons. Import named icons
  (`import { Key, Plus, Trash2 } from "lucide-react"`).
- **Tailwind for styling.** No inline `style` except for truly dynamic values. Use
  design tokens (configured in `tailwind.config.ts`), not arbitrary hex values scattered
  in JSX.

## Layer separation

| Folder            | Contains                                                              | Must NOT contain                     |
| ----------------- | -------------------------------------------------------------------- | ------------------------------------ |
| `components/`     | Reusable presentational + composite UI. Receive data via props.      | API calls, route logic               |
| `pages/`          | Route-level screens; compose components, call hooks/services.        | Low-level markup that belongs in components |
| `layouts/`        | App shell (sidebar, topbar), auth layout.                            | Feature logic                        |
| `services/`       | API client modules (one per resource), return typed data.            | React/JSX                            |
| `hooks/`          | Custom hooks (data fetching via services, auth, ui state).          | JSX                                  |
| `context/`        | React contexts (e.g. `AuthContext`).                                 | Unrelated business logic             |
| `types/`          | Shared interfaces mirroring API DTOs (`api-spec.md`).                | Runtime code                         |
| `utils/`          | Pure functions (formatters, validators, clipboard helper).           | React, side effects                  |
| `lib/`            | Configured third-party clients (axios instance, query client).       | Feature logic                        |
| `routes/`         | Route table + `ProtectedRoute` wrapper.                              | —                                    |

Rule of thumb: **a component never calls axios directly.** Components → hooks →
services → API. Types flow from `types/` and match the backend contract exactly.

## Data fetching

- Use **TanStack Query (React Query)** for server state; axios for the transport.
- A single configured axios instance in `lib/api.ts` with interceptors:
  - attach `Authorization: Bearer <access>`;
  - on `401`, attempt one refresh via `/auth/refresh`, then retry; on failure, log out.
- Services return typed promises, e.g. `getSecret(envId, key): Promise<Secret>`.
- Query keys are centralized and typed.

## Design direction (avoid "AI slop")

- **Restrained palette**: one neutral gray scale + a single brand/accent color. Avoid
  purple-to-blue gradients, glassmorphism, and rainbow buttons.
- **Clear hierarchy**: consistent type scale, generous whitespace, aligned grids.
- **Density appropriate for a tool**: this is a developer dashboard — tables, compact
  rows, keyboard-friendly. Not a marketing landing page.
- **Consistent components**: buttons, inputs, badges, cards share one system. Define
  primitives once in `components/ui/` and reuse.
- **Meaningful icons**: lucide icons that match the action (Key for secrets, FolderOpen
  for projects, History for versions, Eye/EyeOff for reveal, Copy for clipboard).
- **States are designed, not afterthoughts**: every list/detail has explicit loading
  (skeletons), empty, and error states.
- **Accessibility**: semantic HTML, labelled inputs, focus rings, sufficient contrast,
  keyboard navigation. Touch/click targets ≥ 36px.
- **Secret values masked by default**, revealed on explicit action, with copy-to-clipboard.

## Tailwind setup

- Define tokens in `tailwind.config.ts`: `colors` (neutral + accent), `fontFamily`
  (a clean sans like Inter), spacing scale if extended.
- Prefer composition via small components over long duplicated class strings; extract
  repeated class sets into components, not `@apply` soup.

## File/naming conventions

- Components: `PascalCase.tsx`, one component per file, colocated styles via Tailwind.
- Hooks: `useThing.ts`. Services: `thing.service.ts`. Types: `thing.ts` in `types/`.
- Exports: named exports preferred; default export only for route page components if desired.
- No `console.log` in committed code; use a small `logger` util gated by env if needed.

## Tooling

- ESLint + Prettier configured; `@typescript-eslint` with `no-explicit-any`,
  `no-floating-promises`. Lint must pass clean.
- `npm run build` (tsc + vite) must succeed with zero type errors.
