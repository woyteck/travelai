package memory

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"woyteck.pl/travelai/openai"
	"woyteck.pl/travelai/qdrant_client"
)

type Memory struct {
	Id        int
	Uuid      uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Type      string
	Source    string
	Content   string
}

func Remember(info string) {
	vector := openai.GetEmbedding(info, "text-embedding-ada-002")

	qdrant := qdrant_client.NewClient()

	payload := map[string]any{}
	payload["info"] = info

	qdrant.UpsertPoints("memory", vector, uuid.New(), payload)
}

func Recall(text string) string {
	vector := openai.GetEmbedding(text, "text-embedding-ada-002")

	qdrant := qdrant_client.NewClient()
	response := qdrant.Search("memory", vector, 1)
	if len(response.Result) == 0 {
		return ""
	}
	payload := response.Result[0].Payload

	if response.Result[0].Score < 0.8 {
		return ""
	}

	return payload["info"].(string)
}

func RememberArticle(db *sql.DB, paragraphs []string, url string) error {
	if articleExists(db, url) {
		return nil
	}

	memory := Memory{
		Uuid:      uuid.New(),
		CreatedAt: time.Now(),
		Type:      "web_article",
		Source:    url,
		Content:   strings.Join(paragraphs, "\n"),
	}

	lastInsertId := 0
	err := db.QueryRow("INSERT INTO memories (uuid, created_at, memory_type, source, content) VALUES ($1, $2, $3, $4, $5) RETURNING id", memory.Uuid, memory.CreatedAt, memory.Type, memory.Source, memory.Content).Scan(&lastInsertId)
	if err != nil {
		return err
	}

	memory.Id = lastInsertId

	for _, paragraph := range paragraphs {
		_, err = db.Exec("INSERT INTO memory_fragments (created_at, content_original, memory_id) VALUES ($1, $2, $3)", memory.CreatedAt, paragraph, memory.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func articleExists(db *sql.DB, source string) bool {
	id := 0
	err := db.QueryRow("SELECT id FROM memories WHERE source=$1", source).Scan(&id)
	if err != nil {
		return false
	}

	if err == sql.ErrNoRows {
		return false
	}

	return true
}
