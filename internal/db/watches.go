package db

import (
	"context"
	"log"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type WatchData struct {
	MeetingID string
	GuildID   string
	ChannelID string
	Options   types.FeatureFlags
}

func (db DatabasePool) GetAllWatches() []WatchData {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()
	conn, err := db.pool.Take(ctx)
	if err != nil {
		log.Println("error: could not get new connection from database: %w", err)
	}
	defer db.pool.Put(conn)

	var watches []WatchData
	err = sqlitex.Execute(conn, `
		SELECT
			meeting_id,
			server_id,
			channel_id,
			silent,
			summary,
			history_type,
			command,
			link
		FROM watches;`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				watchData := WatchData{
					MeetingID: stmt.ColumnText(0),
					GuildID:   stmt.ColumnText(1),
					ChannelID: stmt.ColumnText(2),
					Options: types.FeatureFlags{
						Silent:         stmt.ColumnBool(3),
						Summaries:      stmt.ColumnBool(4),
						HistoryLevel:   stmt.ColumnText(5),
						RestartCommand: stmt.ColumnText(6),
						JoinLink:       stmt.ColumnText(7),
					},
				}
				watches = append(watches, watchData)
				return nil
			},
		})
	if err != nil {
		log.Println("error: could not get all watches from database: %w", err)
	}

	return watches
}

func (db DatabasePool) SaveWatch(watch WatchData) {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()
	conn, err := db.pool.Take(ctx)
	if err != nil {
		log.Println("error: could not get new connection from database: %w", err)
	}
	defer db.pool.Put(conn)

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
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()
	conn, err := db.pool.Take(ctx)
	if err != nil {
		log.Println("error: could not get new connection from database: %w", err)
	}
	defer db.pool.Put(conn)

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
