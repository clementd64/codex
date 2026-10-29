-- name: SearchGame :many
select id, name, cover_image_id,
    ceil(count(*) over ()::numeric / @_limit)::bigint as total_pages
from igdb.main_game
where @query::text <% search_names
order by word_similarity(@query::text, search_names) desc, name, id
limit @_limit offset @_offset;
