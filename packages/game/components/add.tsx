import { Pagination } from "design-system/components/pagination";
import { Redirect, useLocation } from "design-system/components/router";
import { useClient } from "design-system/lib/client";
import { useAsync } from "design-system/lib/hook";
import { useState } from "preact/hooks";
import { GameService } from "proto";

export function Add() {
  const client = useClient(GameService);

  const { searchParams, setSearchParams } = useLocation();
  const query = searchParams.get("q");

  const page = Number(searchParams.get("p") ?? 1);
  if (isNaN(page)) {
    return <Redirect to="/game/add" />;
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
        {data?.results.map((game) => (
          <div key={game.id} class="card bg-base-200 shadow-md h-full">
            {game.cover ? (
              <figure>
                <img src={game.cover} alt="poster" />
              </figure>
            ) : null}

            <div class="card-body p-4">
              <h2 class="card-title text-base">{game.name}</h2>
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
