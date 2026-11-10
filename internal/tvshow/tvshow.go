package tvshow

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"connectrpc.com/connect/v2"
	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/auth"
	tvshowpb "github.com/clementd64/codex/internal/tvshow/proto"
	"github.com/clementd64/codex/internal/tvshow/proto/tvshowconnect"
	tvshowquery "github.com/clementd64/codex/internal/tvshow/query"
	"github.com/clementd64/codex/internal/tvshow/tmdb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TvShowServer struct {
	app   *app.App
	tmdb  tmdb.ClientWithResponsesInterface
	query *tvshowquery.Queries
}

func (s *TvShowServer) List(
	ctx context.Context,
	req *tvshowpb.ListRequest,
) (*tvshowpb.ListResponse, error) {
	if req.GetPage() < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, "page must be positive")
	}

	userID, err := auth.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	list, err := s.query.ListShow(ctx, tvshowquery.ListShowParams{
		UserID: userID,
		Limit:  24,
		Offset: 24 * (req.GetPage() - 1),
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return &tvshowpb.ListResponse{
			Page:       1,
			TotalPages: 1,
		}, nil
	}

	res := &tvshowpb.ListResponse{
		Page:       int32(req.GetPage()),
		TotalPages: int32(list[0].TotalPages),
		Results:    make([]*tvshowpb.TvShow, len(list)),
	}
	for i, show := range list {
		var poster *string
		if show.Poster.Valid {
			p := tmdb.PosterPath(show.Poster.String)
			poster = &p
		}
		var firstAirDate *string
		if show.Firstairdate.Valid {
			date := show.Firstairdate.Time.Format(time.DateOnly)
			firstAirDate = &date
		}

		res.Results[i] = &tvshowpb.TvShow{
			Id:           show.ID.String(),
			Name:         show.Name,
			Poster:       poster,
			FirstAirDate: firstAirDate,
			Country:      &show.Country.String,
			OriginalName: &show.Originalname.String,
			Status:       nullableStringPointer(show.Status),
		}
	}

	return res, nil
}

func (s *TvShowServer) Search(
	ctx context.Context,
	req *tvshowpb.SearchRequest,
) (*tvshowpb.SearchResponse, error) {
	if req.GetQuery() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "query required")
	}

	if req.GetPage() < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, "page must be positive")
	}

	userID, err := auth.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	search, err := tmdb.Unwrap(s.tmdb.SearchTvWithResponse(ctx, &tmdb.SearchTvParams{
		Query: req.GetQuery(),
		Page:  req.GetPage(),
	}))
	if err != nil {
		slog.ErrorContext(ctx, "TMDB search failed", "error", err)
		return nil, err
	}

	tmdbIDs := make([]int32, len(search.Results))
	for i, result := range search.Results {
		tmdbIDs[i] = int32(result.Id)
	}
	addedIDs, err := s.query.ListAddedShowIDs(ctx, tvshowquery.ListAddedShowIDsParams{
		UserID:  userID,
		TmdbIds: tmdbIDs,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Added TV show lookup failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not look up added TV shows")
	}
	added := make(map[int32]bool, len(addedIDs))
	for _, id := range addedIDs {
		added[id] = true
	}

	res := &tvshowpb.SearchResponse{
		Page:       int32(search.Page),
		TotalPages: int32(search.TotalPages),
		Results:    make([]*tvshowpb.SearchTvShow, len(search.Results)),
	}

	for i, result := range search.Results {
		var country *string
		if len(result.OriginCountry) > 0 {
			country = &result.OriginCountry[0]
		}

		var poster *string
		if result.PosterPath != "" {
			p := tmdb.PosterPath(result.PosterPath)
			poster = &p
		}

		res.Results[i] = &tvshowpb.SearchTvShow{
			TmdbId:       int32(result.Id),
			Name:         result.Name,
			Poster:       poster,
			FirstAirDate: &result.FirstAirDate,
			Country:      country,
			OriginalName: &result.OriginalName,
			Added:        added[int32(result.Id)],
		}
	}

	return res, nil
}

func (s *TvShowServer) Add(
	ctx context.Context,
	req *tvshowpb.AddRequest,
) (*emptypb.Empty, error) {
	if req.GetTmdbId() <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, "positive TMDB ID required")
	}

	userID, err := auth.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	showID, err := s.query.GetShowID(ctx, req.GetTmdbId())
	if errors.Is(err, sql.ErrNoRows) {
		showID, err = syncShow(ctx, s.app.DB, s.tmdb, req.GetTmdbId())
		if err != nil {
			slog.ErrorContext(ctx, "Fetching TV show failed", "error", err)
			if connect.CodeOf(err) == connect.CodeUnknown {
				return nil, connect.NewError(connect.CodeInternal, "could not save TV show")
			}
			return nil, err
		}
	} else if err != nil {
		slog.ErrorContext(ctx, "TV show lookup failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not look up TV show")
	}

	if err := s.query.AddUserShow(ctx, tvshowquery.AddUserShowParams{UserID: userID, ShowID: showID}); err != nil {
		slog.ErrorContext(ctx, "Adding TV show failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not add TV show")
	}
	return &emptypb.Empty{}, nil
}

func (s *TvShowServer) Delete(
	ctx context.Context,
	req *tvshowpb.DeleteRequest,
) (*emptypb.Empty, error) {
	userID, err := auth.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	showID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, "invalid id")
	}

	if err := s.query.DeleteUserShow(ctx, tvshowquery.DeleteUserShowParams{
		UserID: userID,
		ShowID: showID,
	}); err != nil {
		slog.ErrorContext(ctx, "Delete TV show failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not delete TV show")
	}

	return &emptypb.Empty{}, nil
}

func (s *TvShowServer) ICal(ctx context.Context, user uuid.UUID) ([]app.ICalEntry, error) {
	rows, err := s.query.GetIcal(ctx, user)
	if err != nil {
		return nil, err
	}
	entries := make([]app.ICalEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, app.ICalEntry(row))
	}
	return entries, nil
}

func Register(server *connect.Server, application *app.App) *TvShowServer {
	tmdbClient, _ := tmdb.NewTMDBClient(application.TMDBApikey)
	s := &TvShowServer{
		app:   application,
		tmdb:  tmdbClient,
		query: tvshowquery.New(application.DB),
	}
	tvshowconnect.RegisterTvShowServiceHandler(server, s)
	return s
}
