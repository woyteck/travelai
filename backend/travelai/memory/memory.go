package memory

import (
	"database/sql"
	"os"
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

type MemoryFragment struct {
	Id              int       `db:"id"`
	CreatedAt       time.Time `db:"created_at"`
	ContentOriginal string    `db:"content_original"`
	ContentRefined  string    `db:"content_refined"`
	IsRefined       bool      `db:"is_refined"`
	IsEmbedded      bool      `db:"is_embedded"`
	MemoryId        int       `db:"memory_id"`
}

func Remember(text string, fragmentId int) {
	client := openai.New(openai.Config{
		ApiKey: os.Getenv("OPENAI_API_KEY"),
	})
	vector := client.GetEmbedding(text, "text-embedding-ada-002")

	qdrant := qdrant_client.NewClient()

	payload := map[string]any{}
	payload["text"] = text
	payload["fragment_id"] = fragmentId

	qdrant.UpsertPoints("memory", vector, uuid.New(), payload)
}

func Recall(text string) string {
	client := openai.New(openai.Config{
		ApiKey: os.Getenv("OPENAI_API_KEY"),
	})
	vector := client.GetEmbedding(text, "text-embedding-ada-002")

	qdrant := qdrant_client.NewClient()
	response := qdrant.Search("memory", vector, 1)
	if len(response.Result) == 0 {
		return ""
	}
	payload := response.Result[0].Payload

	if response.Result[0].Score < 0.8 {
		return ""
	}

	return payload["text"].(string)
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

func GetNotRefinedFragments(db *sql.DB) ([]MemoryFragment, error) {
	rows, err := db.Query("SELECT id, created_at, content_original, content_refined, is_refined, is_embedded, memory_id FROM memory_fragments WHERE is_refined=false")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	fragments := []MemoryFragment{}
	for rows.Next() {
		fragment := MemoryFragment{}
		rows.Scan(&fragment.Id, &fragment.CreatedAt, &fragment.ContentOriginal, &fragment.ContentRefined, &fragment.IsRefined, &fragment.IsEmbedded, &fragment.MemoryId)
		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

func GetNotEmbeddedFragments(db *sql.DB) ([]MemoryFragment, error) {
	rows, err := db.Query("SELECT id, created_at, content_original, content_refined, is_refined, is_embedded, memory_id FROM memory_fragments WHERE is_refined=true AND is_embedded=false AND content_refined != ''")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	fragments := []MemoryFragment{}
	for rows.Next() {
		fragment := MemoryFragment{}
		rows.Scan(&fragment.Id, &fragment.CreatedAt, &fragment.ContentOriginal, &fragment.ContentRefined, &fragment.IsRefined, &fragment.IsEmbedded, &fragment.MemoryId)
		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

func EmbedMemories(db *sql.DB, fragments []MemoryFragment) error {
	for _, fragment := range fragments {
		Remember(fragment.ContentRefined, fragment.Id)
		fragment.IsEmbedded = true
		err := UpdateFragment(db, fragment)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateFragment(db *sql.DB, fragment MemoryFragment) error {
	_, err := db.Exec("UPDATE memory_fragments SET content_refined=$1, is_refined=$2, is_embedded=$3 WHERE id=$4", fragment.ContentRefined, fragment.IsRefined, fragment.IsEmbedded, fragment.Id)
	return err
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
