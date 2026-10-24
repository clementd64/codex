import { useMenuAction } from "design-system/components/menuAction";
import { goto } from "design-system/components/router";
import { useClient } from "design-system/lib/client";
import { countryCodeToFlag } from "design-system/lib/country";
import { useAsync } from "design-system/lib/hook";
import { useState } from "preact/hooks";
import { TvShowService } from "proto";

export function Home() {
  useMenuAction(
    [
      {
        name: "Add",
        icon: "icon-[mdi--add]",
        class: "btn-primary",
        onClick() {
          goto("/tvshow/add");
        },
      },
    ],
    [],
  );

  const client = useClient(TvShowService);
  const [reloadCount, setReloadCount] = useState(0);

  const { loading, data } = useAsync(
    async (signal) => {
      return client.list({}, { signal });
    },
    [client, reloadCount],
  );

  if (loading) {
    return <></>;
  }

  return (
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-6 gap-4">
      {data?.results.map((show) => (
        <div key={show.id} class="card bg-base-200 shadow-md h-full">
          {show.poster ? (
            <figure>
              <img src={show.poster} alt="poster" />
            </figure>
          ) : null}

          <div class="card-body p-4">
            <h2 class="card-title text-base">
              {show.country ? countryCodeToFlag(show.country) : ""} {show.name}
            </h2>

            {show.originalName && show.originalName !== show.name ? (
              <p class="text-xs">{show.originalName}</p>
            ) : null}

            {show.firstAirDate ? (
              <p class="text-base-content/50">
                {new Date(show.firstAirDate).toLocaleDateString(navigator.languages, {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </p>
            ) : null}

            <div class="card-actions mt-auto justify-end">
              <DeleteShowButton
                id={show.id}
                onSuccess={() => setReloadCount((count) => count + 1)}
              />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function DeleteShowButton({ id, onSuccess }: { id: string; onSuccess: () => void }) {
  const client = useClient(TvShowService);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleAdd() {
    if (loading) return;

    setLoading(true);
    setError(null);
    try {
      await client.delete({ id });
      onSuccess();
    } catch {
      setError("Could not delete this show. Please try again.");
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
      <button type="button" class="btn btn-error" disabled={loading} onClick={handleAdd}>
        {loading ? (
          <span class="loading loading-spinner loading-sm" aria-hidden="true"></span>
        ) : (
          <span class="icon-[mdi--trash]"></span>
        )}
      </button>
    </>
  );
}
