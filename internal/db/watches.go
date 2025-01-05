package db

import (
	"context"
	"log"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"zombiezen.com/go/sqlite/sqlitex"
)

type WatchData struct {
	MeetingID string
	GuildID   string
	ChannelID string
	Options   types.FeatureFlags
}

func (db DatabasePool) SaveWatch(watch WatchData) {
	conn, err := db.pool.Take(context.TODO())
	if err != nil {
		log.Println("error: could not get new connection from database: %w", err)
	}
	defer conn.Close()

	err = sqlitex.Execute(conn, `
		INSERT INTO watches (
			meeting_id,
			server_id,
			channel_id,
			silent,
			summary,
			history_type,
			command,
			link
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		);`,
		&sqlitex.ExecOptions{
			Args: []any{
				watch.MeetingID,
				watch.GuildID,
				watch.ChannelID,
				watch.Options.Silent,
				watch.Options.Summaries,
				watch.Options.HistoryLevel,
				watch.Options.RestartCommand,
				watch.Options.JoinLink,
			},
		})
	if err != nil {
		log.Println("error: could not save watch to database: %w", err)
	}
}

func (db DatabasePool) DeleteWatch(guildID string, meetingID string) {
	conn, err := db.pool.Take(context.TODO())
	if err != nil {
		log.Println("error: could not get new connection from database: %w", err)
	}
	defer conn.Close()

	err = sqlitex.Execute(conn, `
		DELETE FROM watches
		WHERE meeting_id = ?
			AND server_id = ?;`,
		&sqlitex.ExecOptions{
			Args: []any{meetingID, guildID},
		})
	if err != nil {
		log.Println("error: could not delete watch from database: %w", err)
	}
}
