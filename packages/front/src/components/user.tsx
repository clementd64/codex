import { Auth } from "design-system/components/authprovider";
import { goto } from "design-system/components/router";
import { clearRefreshToken } from "design-system/lib/auth";
import { useCallback, useContext } from "preact/hooks";

export function User() {
  const [credentials, setCredentials] = useContext(Auth);

  const handleLogout = useCallback(async () => {
    clearRefreshToken();
    setCredentials({ token: "", expiration: 0n, userId: "" });
    goto("/login", true);
  }, []);

  const tvShowIcalUrl = new URL(
    `/ical/tvshow/${credentials.userId}`,
    window.location.origin,
  ).toString();

  return (
    <>
      <fieldset class="fieldset border-base-content/40 rounded-box border p-4">
        <legend class="fieldset-legend">TV Show</legend>
        <div class="join w-full">
          <input class="input input-bordered join-item" value={tvShowIcalUrl} readonly />
          <button
            class="btn join-item"
            onClick={() => navigator.clipboard.writeText(tvShowIcalUrl)}
          >
            Copy
          </button>
        </div>
      </fieldset>

      <div class="h-4"></div>

      <button class="btn btn-error" onClick={handleLogout}>
        Logout
      </button>
    </>
  );
}
