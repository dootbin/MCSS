package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

//Message Message struct
type Message struct {
	//	username string
	Content string `json:"content"`
}

//DiscordMessage Send webhook Message to discord
func DiscordMessage(message string, url string) {

	discordMessage := Message{message}

	messageJSON, _ := json.Marshal(discordMessage)

	fmt.Println("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(messageJSON))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}
