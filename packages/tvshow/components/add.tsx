import { Code, ConnectError } from "@connectrpc/connect";
import { ActionButton } from "design-system/components/actionButton";
import { Pagination } from "design-system/components/pagination";
import { Redirect, useLocation } from "design-system/components/router";
import { useClient } from "design-system/lib/client";
import { countryCodeToFlag } from "design-system/lib/country";
import { useAsync } from "design-system/lib/hook";
import { useState } from "preact/hooks";
import { TvShowService } from "proto";

export function Add() {
  const client = useClient(TvShowService);

  const { searchParams, setSearchParams } = useLocation();
  const query = searchParams.get("q");

  const page = Number(searchParams.get("p") ?? 1);
  if (isNaN(page)) {
    return <Redirect to="/tvshow/add" />;
  }

  const { loading, data } = useAsync(
    async (signal) => {
      if (!query) {
        return null;
      }

      return client.search({ query, page }, { signal });
    },
    [query, page, client],
  );

  const [input, setInput] = useState(query ?? "");
  const [savedIds, setSavedIds] = useState<Set<number>>(() => new Set());

  return (
    <div class="grid gap-4">
      <form
        class="flex items-end"
        onSubmit={(e) => {
          e.preventDefault();
          setSearchParams({ q: input, p: undefined });
        }}
      >
        <fieldset class="fieldset py-0">
          <input
            class="input"
            type="text"
            value={input}
            onInput={(e) => setInput(e.currentTarget.value)}
          />
        </fieldset>
        <div class="flex-1"></div>
        <button class="btn btn-primary" disabled={loading}>
          Search
        </button>
      </form>

      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-6 gap-4">
        {data?.results.map((show) => (
          <div key={show.tmdbId} class="card bg-base-200 shadow-md h-full">
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
                  class="btn-primary"
                  disabled={show.added || savedIds.has(show.tmdbId)}
                  onClick={async () => {
                    try {
                      await client.add({ tmdbId: show.tmdbId });
                    } catch (cause) {
                      if (ConnectError.from(cause).code !== Code.AlreadyExists) {
                        throw cause;
                      }
                    }
                    setSavedIds((ids) => new Set(ids).add(show.tmdbId));
                  }}
                  errorMessage="Could not add this show. Please try again."
                >
                  {show.added || savedIds.has(show.tmdbId) ? "Added" : "Add"}
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
