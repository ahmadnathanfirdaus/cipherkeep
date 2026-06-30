import { useId } from "react";

interface LogoProps {
  size?: number;
  className?: string;
}

/**
 * Cipherkeep mark: three nested rounded squares with a keyhole — the envelope
 * encryption layers (KEK → DEK → secret) guarding the innermost value.
 */
export const Logo = ({ size = 28, className }: LogoProps) => {
  const maskId = `ck-${useId().replace(/:/g, "")}`;
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 56 56"
      fill="none"
      role="img"
      aria-label="Cipherkeep"
      className={className}
    >
      <mask id={maskId}>
        <rect x="23" y="23" width="10" height="10" rx="3" fill="#fff" />
        <circle cx="28" cy="27.5" r="1.9" fill="#000" />
        <rect x="26.8" y="28.5" width="2.4" height="3.4" fill="#000" />
      </mask>
      <rect x="9" y="9" width="38" height="38" rx="11" stroke="#2dd4bf" strokeWidth="2.5" />
      <rect x="16.5" y="16.5" width="23" height="23" rx="7" stroke="#0d9488" strokeWidth="2.5" />
      <rect x="23" y="23" width="10" height="10" rx="3" fill="#0d9488" mask={`url(#${maskId})`} />
    </svg>
  );
};
