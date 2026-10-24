-- name: ListShow :many
select show.id, show.name, show.poster, show.firstAirDate::varchar, show.country, show.originalName
from tvshow.show
join tvshow.user_show on show.id = user_show.show_id
where user_show.user_id = @user_id;

-- name: AddUserShow :exec
insert into tvshow.user_show (user_id, show_id) values (@user_id, @show_id);

-- name: DeleteUserShow :exec
delete from tvshow.user_show where user_id = @user_id and show_id = @show_id;

-- name: UpsertShow :one
insert into tvshow.show (tmdb_id, name, poster, firstAirDate, country, originalName)
values ($1, $2, $3, $4, $5, $6)
on conflict (tmdb_id) do update set
    name = excluded.name,
    poster = excluded.poster,
    firstAirDate = excluded.firstAirDate,
    country = excluded.country,
    originalName = excluded.originalName
returning id;

-- name: ListAddedShowIDs :many
select show.tmdb_id
from tvshow.show
join tvshow.user_show on user_show.show_id = show.id
where user_show.user_id = @user_id
  and show.tmdb_id = ANY(@tmdb_ids::int[]);

-- name: GetShowID :one
select id from tvshow.show where tmdb_id = @tmdb_id;
