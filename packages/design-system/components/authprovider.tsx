import { createContext } from "preact";
import { useState } from "preact/hooks";
import { type Credentials, isAuthenticated } from "design-system/lib/auth";
import { Redirect, Router, useLocation } from "design-system/components/router";
import { Callback, Login } from "design-system/components/login";

export const Auth = createContext<[Credentials, (credentials: Credentials) => void]>([
  { token: "", expiration: 0n },
  () => {},
]);

export function AuthProvider({ children }: { children: preact.ComponentChildren }) {
  const { pathname } = useLocation();
  const session = useState<Credentials>({ token: "", expiration: 0n });
  const authenticated = isAuthenticated(session[0]);
  const loginRoute = pathname === "/login" || pathname === "/login/callback";

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
