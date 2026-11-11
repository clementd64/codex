import { createContext } from "preact";
import { useState } from "preact/hooks";
import {
  type Credentials,
  clearRefreshToken,
  isAuthenticated,
  refresh,
} from "design-system/lib/auth";
import { useAsync } from "design-system/lib/hook";
import { Redirect, Router, useLocation } from "design-system/components/router";
import { Callback, Login } from "design-system/components/login";

export const Auth = createContext<[Credentials, (credentials: Credentials) => void]>([
  { token: "", expiration: 0n, userId: "" },
  () => {},
]);

export function AuthProvider({ children }: { children: preact.ComponentChildren }) {
  const { pathname } = useLocation();
  const session = useState<Credentials>({ token: "", expiration: 0n, userId: "" });
  const authenticated = isAuthenticated(session[0]);
  const loginRoute = pathname === "/login" || pathname === "/login/callback";

  const { loading } = useAsync(async (signal) => {
    if (!authenticated) return;

    try {
      const credentials = await refresh();
      if (!signal.aborted) session[1](credentials);
    } catch {
      if (!signal.aborted) {
        clearRefreshToken();
        session[1]({ token: "", expiration: 0n, userId: "" });
      }
    }
  }, []);

  if (loading) {
    return (
      <div class="hero min-h-screen">
        <div class="hero-content">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      </div>
    );
  }

  if (!authenticated && !loginRoute) return <Redirect to="/login" replace />;
  if (authenticated && loginRoute) return <Redirect to="/" replace />;

  return (
    <Auth.Provider value={session}>
      {authenticated ? (
        children
      ) : (
        <Router
          routes={[
            {
              pattern: new URLPattern({ pathname: "/login" }),
              component: Login,
            },
            {
              pattern: new URLPattern({ pathname: "/login/callback" }),
              component: Callback,
            },
          ]}
        ></Router>
      )}
    </Auth.Provider>
  );
}
