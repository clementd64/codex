package igdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/clementd64/codex/internal/app"
	igdbquery "github.com/clementd64/codex/internal/igdb/query"
)

func Fetch(ctx context.Context, app *app.App) error {
	client := NewIGDBClient(ctx, app.IGDBClientId, app.IGDBClientSecret)
	queries := igdbquery.New(app.DB)

	count, err := client.FetchPlatforms(ctx, queries)
	if err != nil {
		return err
	}
	slog.Info("Stored platforms", "count", count, "table", "igdb.platform")

	count, err = client.FetchCompanies(ctx, queries)
	if err != nil {
		return fmt.Errorf("dump companies (%d stored): %w", count, err)
	}
	slog.Info("Stored companies", "count", count, "table", "igdb.company")

	count, err = client.FetchFranchises(ctx, queries)
	if err != nil {
		return fmt.Errorf("dump franchises (%d stored): %w", count, err)
	}
	slog.Info("Stored franchises", "count", count, "table", "igdb.franchise")

	count, err = client.FetchGames(ctx, queries)
	if err != nil {
		return fmt.Errorf("dump games (%d stored): %w", count, err)
	}
	slog.Info("Stored games", "count", count, "table", "igdb.game")

	return nil
}

func (c *Client) FetchPlatforms(ctx context.Context, queries *igdbquery.Queries) (int, error) {
	return c.FetchAll(ctx, "platforms", "name,abbreviation,platform_logo.image_id,versions.name,versions.platform_logo.image_id", "", func(page []Platform) (int, error) {
		batch := igdbquery.UpsertPlatformsParams{
			Ids: make([]int64, len(page)), Names: make([]string, len(page)),
			Abbreviations: make([]string, len(page)), LogoImageIds: make([]string, len(page)),
		}
		for i, platform := range page {
			batch.Ids[i], batch.Names[i], batch.Abbreviations[i] = platform.ID, platform.Name, platform.Abbreviation
			if logo := platform.preferredLogo(); logo != nil {
				batch.LogoImageIds[i] = logo.ImageID
			}
		}
		return len(page), queries.UpsertPlatforms(ctx, batch)
	})
}

func (c *Client) FetchFranchises(ctx context.Context, queries *igdbquery.Queries) (int, error) {
	return c.FetchAll(ctx, "franchises", "name", "", func(page []Franchise) (int, error) {
		batch := igdbquery.UpsertFranchisesParams{Ids: make([]int64, len(page)), Names: make([]string, len(page))}
		for i, franchise := range page {
			batch.Ids[i], batch.Names[i] = franchise.ID, franchise.Name
		}
		return len(page), queries.UpsertFranchises(ctx, batch)
	})
}

func (c *Client) FetchCompanies(ctx context.Context, queries *igdbquery.Queries) (int, error) {
	return c.FetchAll(ctx, "companies", "name,country", "", func(page []Company) (int, error) {
		batch := igdbquery.UpsertCompaniesParams{
			Ids: make([]int64, len(page)), Names: make([]string, len(page)), Countries: make([]int32, len(page)),
		}
		for i, company := range page {
			batch.Ids[i], batch.Names[i], batch.Countries[i] = company.ID, company.Name, company.Country
		}
		return len(page), queries.UpsertCompanies(ctx, batch)
	})
}

type gameRow struct {
	Record
	Summary          string                         `json:"summary"`
	Name             string                         `json:"name"`
	FirstReleaseDate int64                          `json:"first_release_date"`
	CoverImageID     string                         `json:"cover_image_id"`
	PlatformIDs      []int64                        `json:"platform_ids"`
	MainFranchiseID  int64                          `json:"main_franchise_id"`
	FranchiseIDs     []int64                        `json:"franchise_ids"`
	Status           string                         `json:"status"`
	Type             string                         `json:"type"`
	ParentGameID     int64                          `json:"parent_game_id"`
	VersionParentID  int64                          `json:"version_parent_id"`
	RelatedGameIDs   []int64                        `json:"related_game_ids"`
	DeveloperIDs     []int64                        `json:"developer_ids"`
	PublisherIDs     []int64                        `json:"publisher_ids"`
	Localizations    map[string]gameLocalizationRow `json:"localizations"`
}

type gameLocalizationRow struct {
	Record
	Name         string `json:"name"`
	CoverImageID string `json:"cover_image_id"`
}

func (c *Client) FetchGames(ctx context.Context, queries *igdbquery.Queries) (int, error) {
	const fields = "name,summary,first_release_date,cover.image_id,game_localizations.name,game_localizations.region.identifier,game_localizations.cover.image_id,keywords.name,platforms,franchise,franchises,game_type.type,game_status.status,parent_game,version_parent,dlcs,expansions,standalone_expansions,remakes,remasters,ports,expanded_games,involved_companies.company,involved_companies.developer,involved_companies.publisher"
	return c.FetchAll(ctx, "games", fields, "game_type = (0,1,2,3,4,6,7,8,9,10,11,13,14)", func(page []Game) (int, error) {
		rows := make([]gameRow, 0, len(page))
		for _, game := range page {
			if game.Type == nil || !officialGameType(game.Type.ID) || slices.ContainsFunc(game.Keywords, func(keyword Keyword) bool {
				return strings.Contains(strings.ToLower(keyword.Name), "unofficial")
			}) {
				continue
			}
			row := gameRow{
				Record: game.Record, Name: game.Name, Summary: game.Summary, FirstReleaseDate: game.FirstReleaseDate,
				Localizations: make(map[string]gameLocalizationRow),
				PlatformIDs:   game.Platforms, MainFranchiseID: game.Franchise, FranchiseIDs: game.Franchises,
				Type: game.Type.Type, ParentGameID: game.ParentGame, VersionParentID: game.VersionParent,
			}
			if game.Status != nil {
				row.Status = game.Status.Status
			}
			for _, localization := range game.Localizations {
				if localization.Region == nil || localization.Region.Identifier == "" || localization.Cover == nil || localization.Cover.ImageID == "" {
					continue
				}
				row.Localizations[localization.Region.Identifier] = gameLocalizationRow{
					Record: localization.Record, Name: localization.Name,
					CoverImageID: localization.Cover.ImageID,
				}
			}
			for _, ids := range [][]int64{game.DLCs, game.Expansions, game.StandaloneExpansions, game.Remakes, game.Remasters, game.Ports, game.ExpandedGames} {
				row.RelatedGameIDs = appendUniqueIDs(row.RelatedGameIDs, ids...)
			}
			for _, involved := range game.InvolvedCompanies {
				if !involved.Developer && !involved.Publisher {
					continue
				}
				if involved.Company <= 0 {
					return 0, fmt.Errorf("igdb game %d: invalid company %d", game.ID, involved.Company)
				}
				if involved.Developer {
					row.DeveloperIDs = appendUniqueIDs(row.DeveloperIDs, involved.Company)
				}
				if involved.Publisher {
					row.PublisherIDs = appendUniqueIDs(row.PublisherIDs, involved.Company)
				}
			}
			if game.Cover != nil {
				row.CoverImageID = game.Cover.ImageID
			}
			rows = append(rows, row)
		}
		if len(rows) == 0 {
			return 0, nil
		}
		batch, err := json.Marshal(rows)
		if err != nil {
			return 0, err
		}
		return len(rows), queries.UpsertGames(ctx, batch)
	})
}

// Prefer the original hardware image over regional revisions, such as iQue for N64.
func (p Platform) preferredLogo() *PlatformLogo {
	for _, version := range p.Versions {
		if strings.EqualFold(version.Name, "Initial version") && version.Logo != nil && version.Logo.ImageID != "" {
			return version.Logo
		}
	}
	return p.Logo
}

func officialGameType(id int64) bool {
	// Exclude Mod (5), Fork (12), and unknown future types.
	return id >= 0 && id <= 14 && id != 5 && id != 12
}

func appendUniqueIDs(ids []int64, additions ...int64) []int64 {
	for _, id := range additions {
		if id > 0 && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}
