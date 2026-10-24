import type { ComponentChildren } from "preact";
import { useRef, useState } from "preact/hooks";

interface ActionButtonProps {
  onClick: () => Promise<unknown>;
  errorMessage: string;
  children: ComponentChildren;
  loadingChildren?: ComponentChildren;
  class?: string;
  disabled?: boolean;
}

export function ActionButton({
  onClick,
  errorMessage,
  children,
  class: className = "",
  disabled = false,
}: ActionButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleClick() {
    if (loading || disabled) return;

    setLoading(true);
    setError(null);
    try {
      await onClick();
    } catch {
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      {error ? (
        <p role="alert" class="text-error text-sm w-full">
          {error}
        </p>
      ) : null}
      <button
        type="button"
        class={`btn ${className}`}
        disabled={loading || disabled}
        onClick={handleClick}
      >
        {loading ? (
          <span class="loading loading-spinner loading-sm" aria-hidden="true"></span>
        ) : null}
        {children}
      </button>
    </>
  );
}
