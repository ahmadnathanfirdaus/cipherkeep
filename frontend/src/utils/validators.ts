const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const isValidEmail = (value: string): boolean =>
  EMAIL_RE.test(value.trim());

export const isNonEmpty = (value: string): boolean => value.trim().length > 0;

/** Secret keys: uppercase letters, digits, and underscores (env-var style). */
const SECRET_KEY_RE = /^[A-Z][A-Z0-9_]*$/;

export const isValidSecretKey = (value: string): boolean =>
  SECRET_KEY_RE.test(value.trim());
