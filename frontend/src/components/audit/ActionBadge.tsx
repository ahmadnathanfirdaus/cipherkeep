import { Badge } from "@/components/ui/Badge";
import type { BadgeProps } from "@/components/ui/Badge";

interface ActionBadgeProps {
  action: string;
}

/** Map an audit action verb to a badge color based on its effect. */
const variantFor = (action: string): BadgeProps["variant"] => {
  const lower = action.toLowerCase();
  if (lower.includes("delete") || lower.includes("remove")) return "danger";
  if (lower.includes("create") || lower.includes("add")) return "success";
  if (lower.includes("update") || lower.includes("rotate")) return "warning";
  if (lower.includes("read") || lower.includes("reveal")) return "accent";
  return "neutral";
};

export const ActionBadge = ({ action }: ActionBadgeProps) => (
  <Badge variant={variantFor(action)}>
    <span className="font-mono">{action}</span>
  </Badge>
);
