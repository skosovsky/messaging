package user

import "time"

type Message struct {
	UserID      int       `json:"user_id"`
	RecipientID int       `json:"recipient_id"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

const MessageSchema = `{
						  "$schema": "http://json-schema.org/draft-07/schema#",
						  "type": "object",
						  "properties": {
							"user_id": { "type": "integer", "format": "int64" },
							"recipient_id": { "type": "integer", "format": "int64" },
							"message": { "type": "string" },
							"created_at": { "type": "string", "format": "date-time" }
						  },
						  "required": ["user_id", "recipient_id", "message", "created_at"]
						}`
