import { Code, ConnectError } from "@connectrpc/connect";
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

  return (
    <div class="grid gap-4">
      <form
        class="flex items-end"
        onSubmit={(e) => {
          e.preventDefault();
          setSearchParams({ q: input });
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
                <AddShowButton tmdbId={show.tmdbId} added={show.added} />
              </div>
            </div>
          </div>
        ))}
      </div>

      <div class="m-auto">
        <Pagination
          page={page}
          total={data?.totalPages}
          onNavigate={(page) => setSearchParams({ p: page })}
        />
      </div>
    </div>
  );
}

function AddShowButton({ tmdbId, added }: { tmdbId: number; added: boolean }) {
  const client = useClient(TvShowService);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isAdded = added || saved;

  async function handleAdd() {
    if (saving || isAdded) return;

    setSaving(true);
    setError(null);
    try {
      await client.add({ tmdbId });
      setSaved(true);
    } catch (cause) {
      if (ConnectError.from(cause).code === Code.AlreadyExists) {
        setSaved(true);
      } else {
        setError("Could not add this show. Please try again.");
      }
    } finally {
      setSaving(false);
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
        class={isAdded ? "btn" : "btn btn-primary"}
        disabled={saving || isAdded}
        onClick={handleAdd}
      >
        {saving ? (
          <span class="loading loading-spinner loading-sm" aria-hidden="true"></span>
        ) : null}
        {isAdded ? "Added" : saving ? "Adding…" : "Add"}
      </button>
    </>
  );
}
