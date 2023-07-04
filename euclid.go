package main

import (
    "encoding/json"
    "fmt"
    "path/filepath"
	"io/ioutil"
	"os"
    "log"
	"strings"
    "flag"
	"gopkg.in/yaml.v2"
    "github.com/chzyer/readline"
	"github.com/parnurzeal/gorequest"
	"github.com/mitchellh/go-wordwrap"
	"github.com/buger/goterm"
)

type Config struct {
	OpenAIKey string `yaml:"openai_key"`
}

type Message struct {
    Role        string `json:"role"`
    Content     string `json:"content"`
}

type Conversation struct {
	Messages []Message `json:"messages"`
}

func execSingle(config *Config) {
    rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

    input, err := rl.Readline()
	if err != nil {
		return
	}

    var prompt string
    prompt = "\"prompt\": \"" + input + "\""
    request := gorequest.New()
    resp, body, errs := request.Post("https://api.openai.com/v1/completions").
        Set("Content-Type", "application/json").
        Set("Authorization", fmt.Sprintf("Bearer %s", config.OpenAIKey)).
        Send(fmt.Sprintf(`{ %s, "model": "%s", "temperature": 0.5, "max_tokens": 1000}`, prompt, "text-davinci-003")).
        End()

    if errs != nil {
        fmt.Println(errs)
        return
    }

    if resp.StatusCode != 200 {
        fmt.Printf("Received non-200 status code: %d\n", resp.StatusCode)
        return
    }

    var response map[string]interface{}
    err = json.Unmarshal([]byte(body), &response)
    if err != nil {
        fmt.Println(err)
        return
    }

    completion := ""

    choices, ok := response["choices"].([]interface{})
    if !ok || len(choices) == 0 {
        fmt.Println("Unable to retrieve completion from response")
        return
    }
    reply, ok := choices[0].(map[string]interface{})["text"].(string)
    if !ok {
        fmt.Println("Unable to retrieve completion from response")
        return
    }
    completion = reply

    // Clean line and print reply in bold
    fmt.Print("\033[2K\r\n\033[1m" + strings.TrimSpace(completion) + "\033[0m\n\n")
}

func execInteractive(config *Config) {
    rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

    // Prompt input_history
    var (
		line string
		input_history []string
		index_history int
		conversation Conversation
        prompt string
	)

	for {
        rl.Config.AutoComplete = readline.NewPrefixCompleter(
			readline.PcItem("/quit"),
			readline.PcItem("/input_history"),
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

		if input == "/input_history" {
			for _, h := range input_history {
				fmt.Println(h)
			}
			continue
		}

		if input == "up" {
			if index_history > 0 {
				index_history--
				line = input_history[index_history]
				rl.SetPrompt(fmt.Sprintf("> %s", line))
			}
			continue
		}

		if input == "down" {
			if index_history < len(input_history)-1 {
				index_history++
				line = input_history[index_history]
				rl.SetPrompt(fmt.Sprintf("> %s", line))
			} else {
				index_history = len(input_history)
				rl.SetPrompt("> ")
				line = ""
			}
			continue
		}

		input_history = append(input_history, input)
		index_history = len(input_history)

        message_user := Message{
            Role: "user",
            Content: input,
        }

        conversation.Messages = append(conversation.Messages, message_user)

        data, err := json.Marshal(conversation.Messages)
        prompt = "\"messages\": "  + string(data)

        if err != nil {
            fmt.Println(err)
            continue
        }

        request := gorequest.New()
		resp, body, errs := request.Post("https://api.openai.com/v1/chat/completions").
			Set("Content-Type", "application/json").
			Set("Authorization", fmt.Sprintf("Bearer %s", config.OpenAIKey)).
            Send(fmt.Sprintf(`{ %s, "model": "%s", "temperature": 0.5, "max_tokens": 1000}`, prompt, "gpt-3.5-turbo")).
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

        completion := ""

        content, ok := response["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

        if !ok || len(content) == 0 {
            fmt.Println("Unable to retrieve completion from response")
            continue
        }

        completion = content

        // Get the terminal line length
        lineLength := uint(goterm.Width())

        // Clean line and print reply in bold
        hyphenated := wordwrap.WrapString(strings.TrimSpace(completion), lineLength)

		fmt.Print("\033[2K\r\n\033[1m" + hyphenated + "\033[0m\n\n")

        message_assistant := Message {
            Role: "assistant",
            Content: strings.TrimSpace(completion),
        }

        conversation.Messages = append(conversation.Messages, message_assistant)
	}
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
    var mode_interactive bool
    mode_interactive = false

    // Define command line switches
    interactive := flag.Bool("interactive", false, "Enable interactive mode")

    // Parse command line arguments
    flag.Parse()

    // Check if a specific argument was provided
    if flag.Parsed() {
        if *interactive {
            mode_interactive = true;
        }
    }

    config, err := loadConfig()

    if err != nil {
        fmt.Print("Failed to load configuration")
        os.Exit(1)
    }

    if mode_interactive == true {
        execInteractive(config)
    } else {
        execSingle(config)
    }
}

