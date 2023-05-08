package main

import (
    "encoding/json"
    "fmt"
    "path/filepath"
	"io/ioutil"
	"os"
    "log"
	"strings"
	"gopkg.in/yaml.v2"
    "github.com/chzyer/readline"
	"github.com/parnurzeal/gorequest"
)

type Config struct {
	OpenAIKey string `yaml:"openai_key"`
}

func loadConfig() (*Config, error) {
	configFilePath := filepath.Join(os.Getenv("HOME"), ".openai.yaml")
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

    return &config, nil
}

func main() {
    config, err := loadConfig()

    if err != nil {
        fmt.Print("Error loading config: %s\n", err)
        os.Exit(1)
    }

    rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

    var (
		line   string
		history []string
		pos    int
	)

	for {

        rl.Config.AutoComplete = readline.NewPrefixCompleter(
			readline.PcItem("/quit"),
		)
		input, err := rl.Readline()
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "/quit" {
			break
		}

		if input == "/history" {
			for _, h := range history {
				fmt.Println(h)
			}
			continue
		}

		if input == "up" {
			if pos > 0 {
				pos--
				line = history[pos]
				rl.SetPrompt(fmt.Sprintf("> %s", line))
			}
			continue
		}

		if input == "down" {
			if pos < len(history)-1 {
				pos++
				line = history[pos]
				rl.SetPrompt(fmt.Sprintf("> %s", line))
			} else {
				pos = len(history)
				rl.SetPrompt("> ")
				line = ""
			}
			continue
		}

		history = append(history, input)
		pos = len(history)


		request := gorequest.New()
		resp, body, errs := request.Post("https://api.openai.com/v1/completions").
			Set("Content-Type", "application/json").
			Set("Authorization", fmt.Sprintf("Bearer %s", config.OpenAIKey)).
            Send(fmt.Sprintf(`{"prompt": "%s", "model": "text-davinci-003", "temperature": 1.0, "max_tokens": 1000}`, input)).
			End()

		if errs != nil {
			fmt.Println(errs)
			continue
		}

		if resp.StatusCode != 200 {
			fmt.Printf("Received non-200 status code: %d\n", resp.StatusCode)
			continue
		}

		var response map[string]interface{}
		err = json.Unmarshal([]byte(body), &response)
		if err != nil {
			fmt.Println(err)
			continue
		}

		choices, ok := response["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			fmt.Println("Unable to retrieve completion from response")
			continue
		}

		completion, ok := choices[0].(map[string]interface{})["text"].(string)
		if !ok {
			fmt.Println("Unable to retrieve completion from response")
			continue
		}

        // Clean line and print in bold
		fmt.Print("\033[2K\r\n\033[1m" + strings.TrimSpace(completion) + "\033[0m\n\n")
	}
}

