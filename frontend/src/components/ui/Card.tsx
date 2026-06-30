import { cn } from "@/utils/cn";

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

export const Card = ({ className, children, ...props }: CardProps) => (
  <div
    className={cn(
      "rounded-lg border border-neutral-200 bg-white shadow-card",
      className,
    )}
    {...props}
  >
    {children}
  </div>
);

export const CardHeader = ({ className, children, ...props }: CardProps) => (
  <div
    className={cn(
      "flex items-center justify-between gap-3 border-b border-neutral-200 px-5 py-4",
      className,
    )}
    {...props}
  >
    {children}
  </div>
);

export const CardTitle = ({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) => (
  <h2
    className={cn("text-base font-semibold text-neutral-900", className)}
    {...props}
  >
    {children}
  </h2>
);

export const CardBody = ({ className, children, ...props }: CardProps) => (
  <div className={cn("px-5 py-4", className)} {...props}>
    {children}
  </div>
);
