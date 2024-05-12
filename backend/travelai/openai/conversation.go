package openai

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Conversation struct {
	db       *sql.DB
	Id       uuid.UUID `json:"id" db:"uuid"`
	Messages []Message `json:"messages"`
}

func NewConversation(db *sql.DB, initialContext string) *Conversation {
	conv := Conversation{
		db: db,
		Id: uuid.New(),
		Messages: []Message{
			{Role: "system", Content: initialContext},
		},
	}

	_, err := db.Exec("INSERT INTO conversations (uuid) VALUES ($1)", conv.Id)
	if err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec("INSERT INTO messages (role, content, conversation_id) VALUES ($1, $2, $3)", conv.Messages[0].Role, conv.Messages[0].Content, conv.Id)
	if err != nil {
		fmt.Println(err)
	}

	return &conv
}

func GetConversation(db *sql.DB, id string) Conversation {
	query := "SELECT role, content FROM messages WHERE conversation_id = $1"
	rows, err := db.Query(query, id)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	conv := Conversation{
		Id:       uuid.MustParse(id),
		Messages: []Message{},
	}

	for rows.Next() {
		message := Message{}
		rows.Scan(&message.Role, &message.Content)
		conv.Messages = append(conv.Messages, message)
	}

	return conv
}

func AddMessaage(db *sql.DB, conversationId uuid.UUID, role string, content string) {
	_, err := db.Exec("INSERT INTO messages (role, content, conversation_id) VALUES ($1, $2, $3)", role, content, conversationId)
	if err != nil {
		fmt.Println(err)
	}
}
