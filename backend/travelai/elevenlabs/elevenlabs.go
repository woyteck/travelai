package elevenlabs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type VoiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
	Style           float64 `json:"style"`
	UseSpeakerBoost bool    `json:"use_speaker_boost"`
}

type PronunciationDictionaryLocator struct {
	PronunciationDictionaryId string `json:"pronunciation_dictionary_id"`
	VersionId                 string `json:"version_id"`
}

type TextToSpeechRequest struct {
	Text                            string                           `json:"text"`
	ModelId                         string                           `json:"model_id"`
	VoiceSettings                   VoiceSettings                    `json:"voice_settings"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators"`
}

func TextToSpeech(text string) []byte {
	// voicd := "ErXwobaYiN019PkySvjV" //Antoni
	voice := "iP95p4xoKVk53GoZ742B" //Chris
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%v", voice)

	request := TextToSpeechRequest{
		Text:    text,
		ModelId: "eleven_multilingual_v2",
		VoiceSettings: VoiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0,
			Style:           0,
			UseSpeakerBoost: false,
		},
	}

	postBody, _ := json.Marshal(request)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postBody))
	if err != nil {
		log.Fatalf("Error occured %v", err)
	}
	req.Header.Add("xi-api-key", os.Getenv("ELEVENLABS_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error occured %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Coult not read response")
		}
		fmt.Println(string(body))
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error occured %v", err)
	}

	// os.WriteFile("test.mp3", b, 0644)

	return b
}
