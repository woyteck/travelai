package memory

import (
	"github.com/google/uuid"
	"woyteck.pl/travelai/openai"
	"woyteck.pl/travelai/qdrant_client"
)

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
