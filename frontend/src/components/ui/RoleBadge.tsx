import { Badge } from "@/components/ui/Badge";
import type { Role } from "@/types";

interface RoleBadgeProps {
  role: Role;
}

const ROLE_VARIANT = {
  owner: "accent",
  admin: "neutral",
  member: "neutral",
} as const;

export const RoleBadge = ({ role }: RoleBadgeProps) => (
  <Badge variant={ROLE_VARIANT[role]} className="capitalize">
    {role}
  </Badge>
);
