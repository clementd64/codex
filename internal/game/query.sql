-- name: SearchGame :many
with games as (
    select game.id, game.name,
        coalesce(nullif(game.localizations->'EU'->>'cover_image_id', ''), game.cover_image_id, '')::text as cover_image_id,
        game.publisher_ids, game.developer_ids, game.platform_ids, game.first_release_date,
        word_similarity(@query::text, game.search_names) as similarity,
        ceil(count(*) over ()::numeric / @_limit)::bigint as total_pages
    from igdb.main_game as game
    where @query::text <% game.search_names
    order by similarity desc, game.name, game.id
    limit @_limit offset @_offset
)
select game.id, game.name, game.cover_image_id, game.first_release_date,
    coalesce(
        jsonb_agg(jsonb_build_object('id', platform.id, 'name', platform.name) order by platform.name, platform.id)
            filter (where platform.id is not null),
        '[]'::jsonb
    )::jsonb as platforms,
    game.total_pages
from games as game
left join igdb.platform as platform on platform.id = any(game.platform_ids)
group by game.id, game.name, game.cover_image_id, game.first_release_date, game.similarity, game.total_pages
order by game.similarity desc, game.name, game.id;
