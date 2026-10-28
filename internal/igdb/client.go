package igdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Client struct {
	httpClient *http.Client
	clientID   string
}

func NewIGDBClient(ctx context.Context, clientID, clientSecret string) *Client {
	if ctx.Value(oauth2.HTTPClient) == nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: 30 * time.Second})
	}
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://id.twitch.tv/oauth2/token",
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	httpClient := config.Client(ctx)
	httpClient.Timeout = 30 * time.Second
	return &Client{httpClient: httpClient, clientID: clientID}
}

func (c *Client) Fetch[T any](ctx context.Context, endpoint, query string) ([]T, error) {
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.igdb.com/v4/"+endpoint, strings.NewReader(query))
		if err != nil {
			return nil, fmt.Errorf("igdb %s request: %w", endpoint, err)
		}
		req.Header.Set("Client-ID", c.clientID)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "text/plain")

		res, err := c.httpClient.Do(req)
		delay := time.Second << attempt
		if err != nil {
			var networkError net.Error
			if ctx.Err() != nil || attempt >= 3 || !errors.As(err, &networkError) || !networkError.Timeout() {
				return nil, fmt.Errorf("igdb %s request: %w", endpoint, err)
			}
		} else {
			if res.StatusCode == http.StatusOK {
				defer res.Body.Close()
				var results []T
				if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
					return nil, fmt.Errorf("igdb %s response: %w", endpoint, err)
				}
				return results, nil
			}
			res.Body.Close()
			err = fmt.Errorf("igdb %s: unexpected HTTP status %d", endpoint, res.StatusCode)
			if attempt >= 3 || (res.StatusCode != http.StatusTooManyRequests && res.StatusCode < 500) {
				return nil, err
			}
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = max(delay, time.Duration(seconds)*time.Second)
			} else if until, parseErr := http.ParseTime(res.Header.Get("Retry-After")); parseErr == nil {
				delay = max(delay, time.Until(until))
			}
		}
		slog.WarnContext(ctx, "Retrying IGDB request", "endpoint", endpoint, "attempt", attempt+1, "delay", delay, "error", err)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (c *Client) FetchAll[T Model](ctx context.Context, endpoint, fields, filter string, insert func([]T) (int, error)) (int, error) {
	const pageSize = 500
	stored := 0
	var lastID int64
	for {
		where := fmt.Sprintf("id > %d", lastID)
		if filter != "" {
			where += " & " + filter
		}
		query := fmt.Sprintf("fields %s; sort id asc; where %s; limit %d;", fields, where, pageSize)
		page, err := c.Fetch[T](ctx, endpoint, query)
		if err != nil {
			return stored, err
		}
		if len(page) == 0 {
			return stored, nil
		}
		for _, item := range page {
			id := item.GetID()
			if id <= lastID {
				return stored, fmt.Errorf("igdb %s %d: expected ID greater than %d", endpoint, id, lastID)
			}
			lastID = id
		}
		n, err := insert(page)
		if err != nil {
			return stored, fmt.Errorf("store %s batch ending at %d: %w", endpoint, lastID, err)
		}
		stored += n
		if n > 0 && stored%10000 == 0 {
			slog.InfoContext(ctx, "Importing IGDB records", "endpoint", endpoint, "stored", stored, "last_id", lastID)
		}
		if len(page) < pageSize {
			return stored, nil
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return stored, ctx.Err()
		case <-timer.C:
		}
	}
}
