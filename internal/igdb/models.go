package igdb

import "context"

type Model interface {
	GetID() int64
}

type Record struct {
	ID int64 `json:"id"`
}

func (r Record) GetID() int64 {
	return r.ID
}

type Platform struct {
	Record
	Name         string            `json:"name"`
	Abbreviation string            `json:"abbreviation"`
	Logo         *PlatformLogo     `json:"platform_logo,omitempty"`
	Versions     []PlatformVersion `json:"versions,omitempty"`
}

type PlatformLogo struct {
	ImageID string `json:"image_id"`
}

type PlatformVersion struct {
	Name string        `json:"name"`
	Logo *PlatformLogo `json:"platform_logo,omitempty"`
}

func (c *Client) Platforms(ctx context.Context, query string) ([]Platform, error) {
	return c.Fetch[Platform](ctx, "platforms", query)
}

type Company struct {
	Record
	Name    string `json:"name"`
	Country int32  `json:"country"`
}

func (c *Client) Companies(ctx context.Context, query string) ([]Company, error) {
	return c.Fetch[Company](ctx, "companies", query)
}

type Franchise struct {
	Record
	Name string `json:"name"`
}

func (c *Client) Franchises(ctx context.Context, query string) ([]Franchise, error) {
	return c.Fetch[Franchise](ctx, "franchises", query)
}

type Game struct {
	Record
	Name                 string            `json:"name"`
	FirstReleaseDate     int64             `json:"first_release_date"`
	Cover                *GameCover        `json:"cover,omitempty"`
	Platforms            []int64           `json:"platforms"`
	Franchise            int64             `json:"franchise"`
	Franchises           []int64           `json:"franchises"`
	Type                 *GameType         `json:"game_type"`
	ParentGame           int64             `json:"parent_game"`
	VersionParent        int64             `json:"version_parent"`
	DLCs                 []int64           `json:"dlcs"`
	Expansions           []int64           `json:"expansions"`
	StandaloneExpansions []int64           `json:"standalone_expansions"`
	Remakes              []int64           `json:"remakes"`
	Remasters            []int64           `json:"remasters"`
	Ports                []int64           `json:"ports"`
	ExpandedGames        []int64           `json:"expanded_games"`
	InvolvedCompanies    []InvolvedCompany `json:"involved_companies"`
}

type GameType struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type InvolvedCompany struct {
	Company   int64 `json:"company"`
	Developer bool  `json:"developer"`
	Publisher bool  `json:"publisher"`
}

type GameCover struct {
	ImageID string `json:"image_id"`
}

func (c *Client) Games(ctx context.Context, query string) ([]Game, error) {
	return c.Fetch[Game](ctx, "games", query)
}
