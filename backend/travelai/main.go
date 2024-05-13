package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/mkideal/cli"
	"woyteck.pl/travelai/cache"
	"woyteck.pl/travelai/db"
	"woyteck.pl/travelai/elevenlabs"
	"woyteck.pl/travelai/memory"
	"woyteck.pl/travelai/openai"
	prompter "woyteck.pl/travelai/prompter"
	"woyteck.pl/travelai/scraper"
)

type argT struct {
	cli.Helper
	Mode          string `cli:"*mode" usage:"select mode: api | scrap"`
	UrlToScrap    string `cli:"url" usage:"url to scrap"`
	ScrapSelector string `cli:"selector" usage:"css selector to extract text from the page, for example: \".article-content p\""`
}

type Coords struct {
	IsValid   bool    `json:"isValid"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TalkRequest struct {
	ConversationId string `json:"conversationId"`
	Text           string `json:"text"`
	Coords         Coords `json:"coords,omitempty"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// db := db.Connect()
	// cache := cache.New(db)
	// cache.ColelctGarbage()
	// result := cache.Get("test")
	// fmt.Println(result)
	// if result == "" {
	// 	cache.Set("test", "Lorem ipsum", time.Hour)
	// }

	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		if argv.Mode == "api" {
			db := db.Connect()
			startApi(db)
		}
		if argv.Mode == "scrap" {
			if argv.UrlToScrap == "" {
				fmt.Println("url is required")
			}
			if argv.ScrapSelector == "" {
				fmt.Println("scrap css selector is required")
			}
			err := scrapWebsite(argv.UrlToScrap, argv.ScrapSelector) // .article-content p
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	}))
}

func scrapWebsite(url string, selector string) error {
	paragraphs, _ := scraper.ScrapWebPage(url, selector)
	db := db.Connect()
	err := memory.RememberArticle(db, paragraphs, url)
	if err != nil {
		return err
	}

	return nil
}

func startApi(db *sql.DB) {
	r := gin.Default()

	r.GET("/conversation", func(c *gin.Context) {
		conv := openai.NewConversation(db, GetContext())

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

		fmt.Println(request)

		if request.ConversationId == "" {
			fmt.Println("ConversationId not found in request")
			c.AbortWithStatus(400)
			return
		}

		conv := openai.GetConversation(db, request.ConversationId)
		openai.AddMessaage(db, conv.Id, "user", request.Text)
		conv.Messages = append(conv.Messages, openai.Message{Role: "user", Content: request.Text})

		cache := cache.New(db)
		prompter := prompter.New(&cache)
		category := prompter.ClassifyQuestion(request.Text)
		fmt.Println(category)
		if category == "info" {
			memory.Remember(request.Text)
			fmt.Println("remembered: ", request.Text)
		}

		if category == "question" {
			additionalContext := memory.Recall(request.Text)
			if additionalContext != "" {
				additionalContext += "\n"
			}
			if request.Coords.IsValid {
				additionalContext += fmt.Sprintf("Aktualna lokalizacja GPS użytkownika: %v, %v.\n", request.Coords.Latitude, request.Coords.Longitude)
			}
			if additionalContext != "" {
				fmt.Println("Additional context: " + additionalContext)
				openai.AddMessaage(db, conv.Id, "system", additionalContext)
				conv.Messages = append(conv.Messages, openai.Message{Role: "system", Content: additionalContext})
			}
		}
		client := openai.New(openai.Config{
			ApiKey: os.Getenv("OPENAI_API_KEY"),
		})
		completions := client.GetCompletionShort(conv.Messages, "gpt-4-turbo")
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
	// context += "Pozycja GPS użytkownika: 53.12731033501957, 20.350897448728542\n"
	context += "Lokalizacja GPS będzie podawana w komunikatach systemowych"
	context += "Odpowiadam tekstem, nie formatuję odpowiedzi w żaden sposób, nie używam markdown. Jeśli moja odpowiedź ma postać punktów zwracam je zdanie po zdaniu, bez numeracji. Nie używam skrótów. \n"
	context += "wszystkie liczby i cyfry zwracam słownie."
	context += "Moje odpowiedzi są sarkastyczne i przepełnione czarnym humorem, czasem zawierają przekleństwa, np: dupa, zajebiście, kurła."
	context += "Jeśli mam nowe informacje o lokalizacji GPS użytkownika szukam gdzie to jest, żeby umieć nazwać to miejsce. Spytany o lokalizację podaję nazwę najbliższej miejscowości, a nie współrzędne geograficzne."

	return context
}
