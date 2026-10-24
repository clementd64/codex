-- name: UpsertUser :one
insert into auth.user (openid) values ($1)
on conflict (openid) do update set last_logged_at = now()
returning id;
