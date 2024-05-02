package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"woyteck.pl/travelai/db"
	"woyteck.pl/travelai/elevenlabs"
	"woyteck.pl/travelai/openai"
)

type TalkRequest struct {
	ConversationId string `json:"conversationId"`
	Text           string `json:"text"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := db.Connect()

	r := gin.Default()

	r.GET("/conversation", func(c *gin.Context) {
		conv := openai.New(db, GetContext())

		c.JSON(200, gin.H{
			"conversation": conv,
		})
	})

	r.GET("conversation/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")

		conv := openai.GetConversation(db, uuid)

		c.JSON(200, gin.H{
			"conversation": conv,
		})
	})
	r.POST("/talk", func(c *gin.Context) {
		request := TalkRequest{}
		err := c.BindJSON(&request)
		if err != nil {
			c.AbortWithStatus(500)
		}

		if request.ConversationId == "" {
			fmt.Println("ConversationId not found in request")
			c.AbortWithStatus(400)
			return
		}

		conv := openai.GetConversation(db, request.ConversationId)
		openai.AddMessaage(db, conv.Id, "user", request.Text)
		conv.Messages = append(conv.Messages, openai.Message{Role: "user", Content: request.Text})

		completions := openai.GetCompletionShort(conv.Messages, "gpt-4-turbo")
		answer := completions.Choices[0].Message.Content
		openai.AddMessaage(db, conv.Id, "assistant", answer)

		b := elevenlabs.TextToSpeech(answer)
		c.Data(http.StatusOK, "audio/mpeg", b)
	})
	r.Run()
}

func GetContext() string {
	context := "Jestem kumplem do pogawędek, nazywam się Janusz."
	// context := "Jestem doradcą osobistym, mężczyzną. Specjalizuję się w pomocy turystom w znajdowaniu i polecaniu ciekawych miejsc w okolicy dopasowanych do preferencji użytkownika.\n"
	context += "Odpowiadam krótko i na temat, jeśli czegoś nie wiem, to po prostu odpowiadam, że nie wiem, nic innego.\n"
	context += "Pozycja GPS użytkownika: 53.12731033501957, 20.350897448728542\n"
	context += "Odpowiadam tekstem, nie formatuję odpowiedzi w żaden sposób, nie używam markdown. Jeśli moja odpowiedź ma postać punktów zwracam je zdanie po zdaniu, bez numeracji. Nie używam skrótów. \n"
	context += "wszystkie liczby i cyfry zwracam słownie."
	context += "Moje odpowiedzi są sarkastyczne i przepełnione czarnym humorem, czasem zawierają przekleństwa, np: dupa, zajebiście, kurła."

	return context
}
