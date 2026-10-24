import type { ComponentType } from "preact";
import { useEffect, useLayoutEffect, useState } from "preact/hooks";

interface Route {
  pattern: URLPattern;
  component: ComponentType;
}

export function useLocation() {
  const [url, setUrl] = useState(() => new URL(location.href));

  useLayoutEffect(() => {
    const onLocationChange = () => setUrl(new URL(location.href));
    window.addEventListener("popstate", onLocationChange);
    onLocationChange();
    return () => window.removeEventListener("popstate", onLocationChange);
  }, []);

  return url;
}

export function goto(path: string, replace = false) {
  if (replace) {
    window.history.replaceState({}, "", path);
  } else {
    window.history.pushState({}, "", path);
  }
  window.dispatchEvent(new Event("popstate"));
}

export function Redirect({ to, replace = false }: { to: string; replace?: boolean }) {
  useEffect(() => {
    goto(to, replace);
  }, [to, replace]);

  return null;
}

export function Router({ routes }: { routes: Route[] }) {
  const url = useLocation();

  for (const route of routes) {
    if (route.pattern.exec(url.href)) {
      return <route.component />;
    }
  }

  return <PageNotFound />;
}

function PageNotFound() {
  return (
    <div class="hero min-h-screen">
      <div class="hero-content flex-col text-center">
        <h1 class="font-bold text-3xl">Page not found</h1>
        <button class="btn btn-primary" onClick={() => goto("/")}>
          Go back home
        </button>
      </div>
    </div>
  );
}
