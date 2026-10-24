import { useEffect, useState } from "preact/hooks";

type AsyncState<T> = {
  loading: boolean;
  data: T | null;
  error: Error | null;
};

export function useAsync<T>(
  asyncFn: (signal: AbortSignal) => Promise<T>,
  deps: readonly unknown[] = [],
): AsyncState<T> {
  const [state, setState] = useState<AsyncState<T>>({
    loading: true,
    data: null,
    error: null,
  });

  useEffect(() => {
    const controller = new AbortController();

    void (async () => {
      setState({ loading: true, data: null, error: null });

      try {
        const data = await asyncFn(controller.signal);
        if (!controller.signal.aborted) {
          setState({ loading: false, data, error: null });
        }
      } catch (cause: unknown) {
        if (controller.signal.aborted) return;
        setState({
          loading: false,
          data: null,
          error: cause instanceof Error ? cause : new Error(String(cause)),
        });
      }
    })();

    return () => controller.abort();
  }, deps);

  return state;
}
