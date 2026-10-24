import { useCallback, useContext, useState } from "preact/hooks";
import { authorize, login } from "design-system/lib/auth";
import { useAsync } from "design-system/lib/hook";
import { Redirect, useLocation } from "design-system/components/router";
import { Auth } from "design-system/components/authprovider";

export function Login() {
  const url = useLocation();
  const [loading, setLoading] = useState(false);

  const handleLogin = useCallback(async () => {
    setLoading(true);

    try {
      await authorize();
    } finally {
      setLoading(false);
    }
  }, []);

  return (
    <div class="hero bg-base-200 min-h-screen">
      <div class="hero-content">
        <fieldset class="fieldset bg-base-100 rounded-box w-xs p-4 shadow-2xl">
          <legend class="fieldset-legend">Login</legend>
          {url.searchParams.has("error") ? (
            <div role="alert" class="alert alert-error">
              <span class="icon-[mdi--cross-circle-outline]"></span>
              <span>{url.searchParams.get("error")}</span>
            </div>
          ) : null}
          <button class="btn btn-neutral" type="button" onClick={handleLogin} disabled={loading}>
            {loading ? <span class="loading loading-spinner"></span> : null}
            Login with {import.meta.env.VITE_OIDC_PROVIDER_NAME}
          </button>
        </fieldset>
      </div>
    </div>
  );
}

export function Callback() {
  const url = useLocation();
  const [_, setCredentials] = useContext(Auth);

  const { loading, error } = useAsync(async (signal) => {
    if (!url.searchParams.has("code") || !url.searchParams.has("state")) {
      throw new Error("Missing required query parameters");
    }

    const credentials = await login(url.searchParams.get("code")!, url.searchParams.get("state")!);
    if (!signal.aborted) {
      setCredentials(credentials);
    }
  }, []);

  if (error) {
    return <Redirect to={`/login?${new URLSearchParams({ error: error.message })}`} />;
  }

  if (loading) {
    return (
      <div class="hero min-h-screen">
        <div class="hero-content">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      </div>
    );
  }

  return null;
}
