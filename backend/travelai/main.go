package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"woyteck.pl/travelai/elevenlabs"
	"woyteck.pl/travelai/openai"
)

type TalkRequest struct {
	Text string `json:"text"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/talk", func(c *gin.Context) {
		request := TalkRequest{}
		err := c.BindJSON(&request)
		if err != nil {
			c.AbortWithStatus(500)
		}

		messages := []openai.Message{
			{
				Role:    "system",
				Content: GetContext(),
			},
			{
				Role:    "user",
				Content: request.Text,
			},
		}
		messages = append(messages, openai.Message{
			Role:    "user",
			Content: request.Text,
		})
		completions := openai.GetCompletionShort(messages, "gpt-3.5-turbo")

		b := elevenlabs.TextToSpeech(completions.Choices[0].Message.Content)

		c.Data(http.StatusOK, "audio/mpeg", b)
	})
	r.Run()
}

func GetContext() string {
	context := "Jestem doradcą osobistym, mężczyzną. Specjalizuję się w pomocy turystom w znajdowaniu i polecaniu ciekawych miejsc w okolicy dopasowanych do preferencji użytkownika.\n"
	context += "Odpowiadam krótko i na temat, jeśli czegoś nie wiem, to po prostu odpowiadam, że nie wiem, nic innego.\n"
	context += "Pozycja GPS użytkownika: 40.85321191865616, 14.268199042559345\n"
	context += "Użytkownik lubi placki w każdej formie, burbon whiskey (szczególnie Jack Daniel's z Pepsi).\n"

	return context
}
