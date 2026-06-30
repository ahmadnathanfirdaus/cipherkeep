type ClassValue = string | number | false | null | undefined;

/** Tiny classnames joiner — filters falsy values, joins with a space. */
export const cn = (...classes: ClassValue[]): string =>
  classes.filter((c): c is string | number => Boolean(c)).join(" ");
