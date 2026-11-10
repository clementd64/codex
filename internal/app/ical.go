package app

import (
	"context"
	"net/http"
	"time"
	"uuid"
)

type ICalEntry struct {
	Uid     string
	Date    time.Time
	Summary string
}

type ICalGenerator interface {
	ICal(ctx context.Context, user uuid.UUID) ([]ICalEntry, error)
}

func renderIcal(w http.ResponseWriter, entries []ICalEntry) {
	w.Header().Set("Content-Type", "text/calendar")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("BEGIN:VCALENDAR\n"))
	w.Write([]byte("VERSION:2.0\n"))

	for _, entry := range entries {
		w.Write([]byte("BEGIN:VEVENT\n"))
		w.Write([]byte("UID:" + entry.Uid + "\n"))
		w.Write([]byte("DTSTART;VALUE=DATE:" + entry.Date.Format("20060102") + "\n"))
		w.Write([]byte("DTEND;VALUE=DATE:" + entry.Date.Add(24*time.Hour).Format("20060102") + "\n"))
		w.Write([]byte("SUMMARY:" + entry.Summary + "\n"))
		w.Write([]byte("END:VEVENT\n"))
	}

	w.Write([]byte("END:VCALENDAR\n"))
}

func IcalHandler(generator ICalGenerator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := uuid.Parse(r.PathValue("user"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		entries, err := generator.ICal(r.Context(), user)
		if err != nil {
			http.Error(w, "Failed to generate iCal", http.StatusInternalServerError)
			return
		}

		renderIcal(w, entries)
	}
}
