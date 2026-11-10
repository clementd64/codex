-- name: ListShow :many
select show.id, show.name, show.poster, show.firstAirDate, show.country, show.originalName, show.status,
       ceil(count(*) over ()::numeric / @_limit)::bigint as total_pages
from tvshow.show
join tvshow.user_show on show.id = user_show.show_id
where user_show.user_id = @user_id
order by show.name
limit @_limit offset @_offset;

-- name: AddUserShow :exec
insert into tvshow.user_show (user_id, show_id) values (@user_id, @show_id);

-- name: DeleteUserShow :exec
delete from tvshow.user_show where user_id = @user_id and show_id = @show_id;

-- name: UpsertShow :one
insert into tvshow.show (tmdb_id, name, poster, firstAirDate, country, originalName, status, backdrop)
values ($1, $2, $3, $4, $5, $6, $7, $8)
on conflict (tmdb_id) do update set
    name = excluded.name,
    poster = excluded.poster,
    firstAirDate = excluded.firstAirDate,
    country = excluded.country,
    originalName = excluded.originalName,
    status = excluded.status,
    backdrop = excluded.backdrop,
    synced_at = now()
returning id;

-- name: ListAddedShowIDs :many
select show.tmdb_id
from tvshow.show
join tvshow.user_show on user_show.show_id = show.id
where user_show.user_id = @user_id
  and show.tmdb_id = ANY(@tmdb_ids::int[]);

-- name: GetShowID :one
select id from tvshow.show where tmdb_id = @tmdb_id;

-- name: FetchShowBatch :many
select tmdb_id from tvshow.show
where tmdb_id > @last_id
order by tmdb_id
limit @_limit;

-- name: UpsertEpisode :exec
insert into tvshow.episode (show_id, season, episode, name, air_date, preview)
values (@show_id, @season, @episode, @name, @air_date, @preview)
on conflict (show_id, season, episode) do update set
    name = excluded.name,
    air_date = excluded.air_date,
    preview = excluded.preview;
