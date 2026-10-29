package game

import (
	"context"
	"log/slog"
	"strings"

	"connectrpc.com/connect/v2"
	"github.com/clementd64/codex/internal/app"
	gamepb "github.com/clementd64/codex/internal/game/proto"
	"github.com/clementd64/codex/internal/game/proto/gameconnect"
	gamequery "github.com/clementd64/codex/internal/game/query"
)

type GameServer struct {
	app   *app.App
	query *gamequery.Queries
}

func (s *GameServer) Search(
	ctx context.Context,
	req *gamepb.SearchRequest,
) (*gamepb.SearchResponse, error) {
	if strings.TrimSpace(req.GetQuery()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "query required")
	}

	if req.GetPage() < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, "page must be positive")
	}

	games, err := s.query.SearchGame(ctx, gamequery.SearchGameParams{
		Query:  strings.TrimSpace(req.GetQuery()),
		Limit:  24,
		Offset: 24 * (req.GetPage() - 1),
	})
	if err != nil {
		slog.ErrorContext(ctx, "search failed", "error", err)
		return nil, err
	}
	if len(games) == 0 {
		return &gamepb.SearchResponse{
			Page:       1,
			TotalPages: 1,
		}, nil
	}

	res := &gamepb.SearchResponse{
		Page:       req.GetPage(),
		TotalPages: int32(games[0].TotalPages),
		Results:    make([]*gamepb.SearchGame, len(games)),
	}
	for i, game := range games {
		var cover *string
		if game.CoverImageID.Valid {
			c := "https://images.igdb.com/igdb/image/upload/t_cover_big/" +
				game.CoverImageID.String + ".jpg"
			cover = &c
		}

		res.Results[i] = &gamepb.SearchGame{
			Id:    int32(game.ID),
			Name:  game.Name,
			Cover: cover,
			Added: false,
		}
	}

	return res, nil
}

func Register(server *connect.Server, application *app.App) {
	gameconnect.RegisterGameServiceHandler(server, &GameServer{
		app:   application,
		query: gamequery.New(application.DB),
	})
}
