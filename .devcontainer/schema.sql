create schema if not exists public;
create extension if not exists pg_trgm with schema public;

create schema auth;

create table auth.user (
    id uuid primary key default uuidv7(),
    openid varchar not null unique,
    created_at timestamptz not null default now(),
    last_logged_at timestamptz not null default now()
);

create schema tvshow;

create table tvshow.show (
    id uuid primary key default uuidv7(),
    tmdb_id int not null unique,
    name varchar not null,
    poster varchar,
    firstAirDate date,
    country char(2),
    originalName varchar,
    created_at timestamptz not null default now()
);

create table tvshow.user_show (
    user_id uuid not null references auth.user(id),
    show_id uuid not null references tvshow.show(id),
    created_at timestamptz not null default now(),
    primary key (user_id, show_id)
);

create schema igdb;

create table igdb.platform (
    id bigint primary key,
    name varchar not null,
    abbreviation varchar not null,
    logo_image_id varchar,
    synced_at timestamptz not null default now()
);

create table igdb.franchise (
    id bigint primary key,
    name varchar not null,
    synced_at timestamptz not null default now()
);

create table igdb.company (
    id bigint primary key,
    name varchar not null,
    country integer,
    synced_at timestamptz not null default now()
);

create table igdb.game (
    id bigint primary key,
    name varchar not null,
    summary text,
    first_release_date timestamptz,
    cover_image_id varchar,
    platform_ids bigint[] not null,
    main_franchise_id bigint,
    franchise_ids bigint[] not null,
    type varchar not null,
    status varchar,
    parent_game_id bigint,
    version_parent_id bigint,
    related_game_ids bigint[] not null,
    developer_ids bigint[] not null,
    publisher_ids bigint[] not null,
    localizations jsonb not null,
    search_names text generated always as (
        (json_array(name returning jsonb) || jsonb_path_query_array(localizations, '$.*.name'))::text
    ) stored,
    synced_at timestamptz not null default now()
);

create index game_search_names_trgm_idx on igdb.game using gin (search_names gin_trgm_ops);

create view igdb.main_game as
select * from igdb.game
where type not in (
    'Bundle',
    'DLC',
    'Expansion',
    'Pack / Addon',
    'Season',
    'Standalone Expansion',
    'Update'
) and (status is null or status not in (
    'Rumored',
    'Cancelled',
    'Delisted',
    'Offline'
));
