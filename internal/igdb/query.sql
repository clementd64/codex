-- name: UpsertPlatforms :exec
insert into igdb.platform (id, name, abbreviation, logo_image_id)
select id, name, abbreviation, nullif(logo_image_id, '')
from unnest(
    @ids::bigint[],
    @names::text[],
    @abbreviations::text[],
    @logo_image_ids::text[]
) as batch(id, name, abbreviation, logo_image_id)
on conflict (id) do update set
    name = excluded.name,
    abbreviation = excluded.abbreviation,
    logo_image_id = excluded.logo_image_id,
    synced_at = now();

-- name: UpsertCompanies :exec
insert into igdb.company (id, name, country)
select id, name, nullif(country, 0)
from unnest(@ids::bigint[], @names::text[], @countries::int[]) as batch(id, name, country)
on conflict (id) do update set
    name = excluded.name,
    country = excluded.country,
    synced_at = now();

-- name: UpsertFranchises :exec
insert into igdb.franchise (id, name)
select id, name
from unnest(@ids::bigint[], @names::text[]) as batch(id, name)
on conflict (id) do update set
    name = excluded.name,
    synced_at = now();

-- name: UpsertGames :exec
insert into igdb.game (
    id, name, summary, first_release_date, cover_image_id,
    platform_ids, main_franchise_id, franchise_ids, game_type,
    parent_game_id, version_parent_id, related_game_ids, developer_ids, publisher_ids, localizations
)
select id, name, nullif(summary, ''), to_timestamp(nullif(first_release_date, 0)), nullif(cover_image_id, ''),
       coalesce(platform_ids, '{}'::bigint[]), nullif(main_franchise_id, 0),
       coalesce(franchise_ids, '{}'::bigint[]), game_type,
       nullif(parent_game_id, 0), nullif(version_parent_id, 0),
       coalesce(related_game_ids, '{}'::bigint[]),
       coalesce(developer_ids, '{}'::bigint[]), coalesce(publisher_ids, '{}'::bigint[]),
       coalesce(nullif(localizations, 'null'::jsonb), '{}'::jsonb)
from jsonb_to_recordset(@games::jsonb) as batch(
    id bigint, name text, summary text, first_release_date bigint, cover_image_id text,
    platform_ids bigint[], main_franchise_id bigint, franchise_ids bigint[], game_type text,
    parent_game_id bigint, version_parent_id bigint, related_game_ids bigint[],
    developer_ids bigint[], publisher_ids bigint[], localizations jsonb
)
on conflict (id) do update set
    name = excluded.name,
    first_release_date = excluded.first_release_date,
    cover_image_id = excluded.cover_image_id,
    platform_ids = excluded.platform_ids,
    main_franchise_id = excluded.main_franchise_id,
    franchise_ids = excluded.franchise_ids,
    game_type = excluded.game_type,
    parent_game_id = excluded.parent_game_id,
    version_parent_id = excluded.version_parent_id,
    related_game_ids = excluded.related_game_ids,
    developer_ids = excluded.developer_ids,
    publisher_ids = excluded.publisher_ids,
    summary = excluded.summary,
    localizations = excluded.localizations,
    synced_at = now();
