import { ActionButton } from "design-system/components/actionButton";
import { useMenuAction } from "design-system/components/menuAction";
import { Pagination } from "design-system/components/pagination";
import { goto, Redirect, useLocation } from "design-system/components/router";
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

  const { searchParams, setSearchParams } = useLocation();
  const page = Number(searchParams.get("p") ?? 1);
  if (isNaN(page)) {
    return <Redirect to="/tvshow" />;
  }

  const [reloadCount, setReloadCount] = useState(0);

  const { loading, data } = useAsync(
    async (signal) => {
      return client.list({ page }, { signal });
    },
    [client, page, reloadCount],
  );

  if (loading) {
    return <></>;
  }

  return (
    <div class="grid gap-4">
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
                <ActionButton
                  class="btn-error"
                  onClick={async () => {
                    await client.delete({ id: show.id });
                    setReloadCount((count) => count + 1);
                  }}
                  errorMessage="Could not delete this show. Please try again."
                >
                  <span class="icon-[mdi--trash]"></span>
                </ActionButton>
              </div>
            </div>
          </div>
        ))}
      </div>

      <Pagination
        class="m-auto"
        page={page}
        total={data?.totalPages}
        onNavigate={(page) => setSearchParams({ p: page })}
      />
    </div>
  );
}
