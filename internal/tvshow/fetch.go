package tvshow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"uuid"

	"connectrpc.com/connect/v2"
	"github.com/clementd64/codex/internal/app"
	tvshowquery "github.com/clementd64/codex/internal/tvshow/query"
	"github.com/clementd64/codex/internal/tvshow/tmdb"
)

type FetchOptions struct {
	// Limit caps the number of shows refreshed; zero means unlimited.
	Limit int32
	// Force refreshes shows regardless of their last sync time.
	Force bool
}

func Fetch(ctx context.Context, application *app.App, options FetchOptions) error {
	if options.Limit < 0 {
		return errors.New("refresh limit must be non-negative")
	}
	client, err := tmdb.NewTMDBClient(application.TMDBApikey)
	if err != nil {
		return err
	}
	query := tvshowquery.New(application.DB)

	ids, err := query.FetchShowBatch(ctx, tvshowquery.FetchShowBatchParams{
		ForceRefresh: options.Force,
		Limit:        options.Limit,
	})
	if err != nil {
		return err
	}

	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := syncShow(ctx, application.DB, client, id); err != nil {
			return err
		}
		slog.InfoContext(ctx, "Refreshed TV show", "tmdb_id", id)
	}
	return nil
}

func syncShow(ctx context.Context, db *sql.DB, client tmdb.ClientWithResponsesInterface, id int32) (uuid.UUID, error) {
	show, episodes, err := fetchShow(ctx, client, id)
	if err != nil {
		return uuid.UUID{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer tx.Rollback()
	query := tvshowquery.New(tx)
	showID, err := query.UpsertShow(ctx, show)
	if err != nil {
		return uuid.UUID{}, err
	}
	for _, episode := range episodes {
		episode.ShowID = showID
		if err := query.UpsertEpisode(ctx, episode); err != nil {
			return uuid.UUID{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return uuid.UUID{}, err
	}
	return showID, nil
}

func fetchShow(ctx context.Context, client tmdb.ClientWithResponsesInterface, id int32) (tvshowquery.UpsertShowParams, []tvshowquery.UpsertEpisodeParams, error) {
	show, err := tmdb.Unwrap(client.TvSeriesDetailsWithResponse(ctx, id, nil))
	if err != nil {
		return tvshowquery.UpsertShowParams{}, nil, err
	}
	params := tvshowquery.UpsertShowParams{
		TmdbID:       id,
		Name:         show.Name,
		Originalname: nullableString(show.OriginalName),
		Poster:       nullableString(show.PosterPath),
		Status:       nullableString(show.Status),
		Backdrop:     nullableString(show.BackdropPath),
	}
	if len(show.OriginCountry) > 0 {
		params.Country = nullableString(show.OriginCountry[0])
	}
	params.Firstairdate, err = nullableDate(show.FirstAirDate)
	if err != nil {
		return params, nil, errors.New("invalid first air date")
	}

	var episodes []tvshowquery.UpsertEpisodeParams
	for _, season := range show.Seasons {
		if season.EpisodeCount == 0 {
			continue
		}
		details, err := tmdb.Unwrap(client.TvSeasonDetailsWithResponse(ctx, id, int32(season.SeasonNumber), nil))
		if err != nil {
			return params, nil, fmt.Errorf("season %d: %w", season.SeasonNumber, err)
		}
		for _, episode := range details.Episodes {
			date, err := nullableDate(episode.AirDate)
			if err != nil {
				return params, nil, connect.NewError(connect.CodeUnavailable, fmt.Sprintf("season %d episode %d has an invalid air date", season.SeasonNumber, episode.EpisodeNumber))
			}
			episodes = append(episodes, tvshowquery.UpsertEpisodeParams{
				Season:  int32(season.SeasonNumber),
				Episode: int32(episode.EpisodeNumber),
				Name:    episode.Name,
				AirDate: date,
				Preview: nullableString(episode.StillPath),
			})
		}
	}
	return params, episodes, nil
}

func nullableString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableDate(value string) (sql.NullTime, error) {
	if value == "" {
		return sql.NullTime{}, nil
	}
	date, err := time.Parse(time.DateOnly, value)
	return sql.NullTime{Time: date, Valid: err == nil}, err
}
