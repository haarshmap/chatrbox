package tasks

import (
	"database/sql"
	"log"

	"github.com/haarshmap/chatrbox/internal/server"
)

func MessageWorker(db *sql.DB, msgChan <-chan server.Message) {
	log.Printf("worker received channel: %p", msgChan)
	stmt, err := db.Prepare(`
        INSERT INTO message_logs
        (message_id, roomcode, username, message, time_stamp, reply_to_id)
        VALUES (?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		log.Printf("failed to prepare statement: %v", err)
		return
	}
	defer stmt.Close()

	log.Println("message worker started")

	for msg := range msgChan {
		_, err := stmt.Exec(
			msg.MessageID,
			msg.RoomCode,
			msg.Username,
			msg.Message,
			msg.Time_Stamp,
			msg.ReplyToID,
		)
		if err != nil {
			log.Printf("failed to insert message: %v", err)
		}
	}
}
