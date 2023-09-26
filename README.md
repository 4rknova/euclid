Euclid is a simple command line interface for ChatGPT.

# Configuration

To use the tool, create a configuration file at your
user home directory called ".euclid.yaml" with the
following contents:

    openai_key: <YOUR OPENAI API KEY>
    model_single: gpt-3.5-turbo-instruct
    model_chat: gpt-3.5-turbo

For more information on available models, see:
https://platform.openai.com/docs/models

# Usage

Simply type your prompts and press return to interact
with ChatGPT. You can use the up / down arrow keys to
navigate previous prompts.

If euclid is run without any arguments, it will do nothing.

The following command line switches are available:


| Switch              | Description                      |
|:--------------------|:---------------------------------|
| --interactive       | Launches euclid in chat mode     |
| --prompt "<string>" | Used for one-off interactions    |

The following commands are supported in interactive mode:

| Command             | Description                      |
|:--------------------|:---------------------------------|
| __/quit__           | Quits the program                |
|__/input_history__   | Shows the history of user inputs |
