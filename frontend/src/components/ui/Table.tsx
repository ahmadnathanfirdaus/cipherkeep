import { cn } from "@/utils/cn";

export const Table = ({
  className,
  children,
  ...props
}: React.TableHTMLAttributes<HTMLTableElement>) => (
  <div className="overflow-x-auto">
    <table
      className={cn("w-full border-collapse text-sm", className)}
      {...props}
    >
      {children}
    </table>
  </div>
);

export const THead = ({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <thead className={cn("border-b border-neutral-200", className)} {...props}>
    {children}
  </thead>
);

export const TBody = ({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLTableSectionElement>) => (
  <tbody className={cn("divide-y divide-neutral-100", className)} {...props}>
    {children}
  </tbody>
);

export const TR = ({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLTableRowElement>) => (
  <tr className={cn("hover:bg-neutral-50/60", className)} {...props}>
    {children}
  </tr>
);

export const TH = ({
  className,
  children,
  ...props
}: React.ThHTMLAttributes<HTMLTableCellElement>) => (
  <th
    className={cn(
      "px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wide text-neutral-500",
      className,
    )}
    {...props}
  >
    {children}
  </th>
);

export const TD = ({
  className,
  children,
  ...props
}: React.TdHTMLAttributes<HTMLTableCellElement>) => (
  <td
    className={cn("px-4 py-3 align-middle text-neutral-700", className)}
    {...props}
  >
    {children}
  </td>
);
