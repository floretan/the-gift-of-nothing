package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/ayush6624/go-chatgpt"
	_ "github.com/joho/godotenv/autoload"
)

// Read the API key from the environment variable OPENAI_KEY or from a .env file.
var openAIKey = os.Getenv("OPENAI_KEY")

func sendRESTCallAndCreateMarkdown(target string, gift string) {

	client, err := chatgpt.NewClient(openAIKey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create the directory if it doesn't exist
	if _, err := os.Stat(fmt.Sprintf("content/posts/%s", target)); os.IsNotExist(err) {
		os.Mkdir(fmt.Sprintf("content/posts/%s", target), 0755)
	}

	filename := fmt.Sprintf("content/posts/%s/%s.md", target, gift)

	// Check if the file already exists.
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		fmt.Printf("%s already exists\n", filename)
		return
	}

	fmt.Printf("Generating: %s for %s\n", gift, target)

	res, err := client.Send(ctx, &chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT35Turbo,

		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: "You are a satirical but fact-based copywriter for a websites that wants to prevent people from buying useless stuff that ends up in landfills. You are writing a gift guide for a " + target + ". You are writing about " + gift + ".",
			},
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: "Format the output as a markdown file with a yaml front matter including a catchy 'title' field and a matching 'path' field for the URL. Don't include headers in the paragraphs.",
			},
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: fmt.Sprintf("Write a satirical article in five paragraphs why a %s is a terrible present for a %s. End with a note suggesting that not giving anything might be the best option.", gift, target),
			},
		},
	})
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	markdownText := string(res.Choices[0].Message.Content)

	// Insert the category into the front matter.
	markdownText = strings.Replace(markdownText, "---\n", fmt.Sprintf("---\ntags: [\"Gifts for %s\", \"%s\"]\n", target, gift), 1)

	err = os.WriteFile(filename, []byte(markdownText), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Printf("Created %s\n", filename)
}

func main() {
	file, err := os.Open("gifts.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var wg sync.WaitGroup

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading record:", err)
			continue
		}

		target := record[0]
		gift := record[1] // Assuming category is in the first column

		wg.Add(1)

		go func(target string, gift string) {
			defer wg.Done()

			// Limit the number of concurrent requests to 5.
			sem := make(chan struct{}, 5)
			sem <- struct{}{}
			defer func() {
				<-sem
				sendRESTCallAndCreateMarkdown(target, gift)
			}()

		}(target, gift)
	}

	wg.Wait()

	fmt.Println("All markdown files created successfully.")
}
