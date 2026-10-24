drop schema if exists public;

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
