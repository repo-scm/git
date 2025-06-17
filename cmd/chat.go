//go:build linux

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
)

var (
	modelID   string
	quietMode bool
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat with workspace for git repo",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		name := args[0]
		prompt := args[1]
		if name == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Please specify a workspace name")
			os.Exit(1)
		}
		if err := runChat(ctx, config, name, prompt); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.PersistentFlags().StringVarP(&modelID, "model", "m", "litellm/anthropic/claude-opus-4-20250514", "model name")
	chatCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "quiet mode")

	chatCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Usage:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  %s %s <workspace_name> [prompt] [flags]\n\n", cmd.Root().Name(), cmd.Name())
		if cmd.HasLocalFlags() {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Flags:\n")
			cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  -%s, --%s   %s", flag.Shorthand, flag.Name, flag.Usage)
				if flag.Name != "help" && flag.DefValue != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), " (default %s)", flag.DefValue)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n")
			})
		}
		if cmd.HasInheritedFlags() {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nGlobal Flags:\n")
			cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  -%s, --%s   %s", flag.Shorthand, flag.Name, flag.Usage)
				if flag.DefValue != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), " (default %s)", flag.DefValue)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n")
			})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nExample:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git chat your_workspace your_prompt --model provider_name/model_id\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git chat your_workspace your_prompt --model provider_name/model_id --quiet\n")
		return nil
	})
}

func runChat(ctx context.Context, cfg *config.Config, name, prompt string) error {
	models := cfg.Models
	if len(models) == 0 {
		return errors.New("no models found\n")
	}

	model, err := selectModel(ctx, models, modelID)
	if err != nil {
		return errors.Wrap(err, "failed to select model\n")
	}

	fmt.Printf("Model %s selected\n", fmt.Sprintf("%s/%s", model.ProviderName, model.ModelId))
	fmt.Printf("Starting chat with workspace: %s\n", name)
	fmt.Println("Type 'exit' to end the session")
	fmt.Println("Type 'help' for available commands")
	fmt.Println(strings.Repeat("-", 50))

	if quietMode {
		return sendMessage(ctx, model, prompt)
	}

	return startInteractiveChat(ctx, model)
}

func selectModel(_ context.Context, models []config.Model, name string) (config.Model, error) {
	if name != "" {
		for _, model := range models {
			if fmt.Sprintf("%s/%s", model.ProviderName, model.ModelId) == name {
				return model, nil
			}
		}
		return config.Model{}, errors.New("model not found\n")
	}

	if len(models) == 1 {
		return models[0], nil
	}

	fmt.Println("Available models:")

	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, fmt.Sprintf("%s/%s", model.ProviderName, model.ModelId))
	}

	fmt.Print("Select a model (1-", len(models), "): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input >= "1" && input <= fmt.Sprintf("%d", len(models)) {
			var index int
			_, _ = fmt.Sscanf(input, "%d", &index)
			if index > 0 && index <= len(models) {
				return models[index-1], nil
			}
		}
	}

	return config.Model{}, errors.New("invalid selection\n")
}

func startInteractiveChat(ctx context.Context, model config.Model) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		switch strings.ToLower(input) {
		case "help":
			showHelp()
			continue
		case "clear":
			clearScreen()
			continue
		case "models":
			fmt.Printf("TBD\n")
			continue
		case "model":
			fmt.Printf("Current model: %s\n", model)
			continue
		case "exit":
			fmt.Println("Goodbye!")
			return nil
		}
		if err := sendMessage(ctx, model, input); err != nil {
			return err
		}
	}

	return nil
}

func sendMessage(_ context.Context, model config.Model, message string) error {
	apiBase := model.ApiBase
	if apiBase == "" {
		return errors.New("no api base found\n")
	}

	apiKey := model.ApiKey
	if apiKey == "" {
		return errors.New("no api key found\n")
	}

	fmt.Printf("Send: %s\n", message)
	fmt.Printf("Response: %s\n", message)
	fmt.Println()

	return nil
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help   - Show this help message")
	fmt.Println("  clear  - Clear the screen")
	fmt.Println("  models - Show all models")
	fmt.Println("  model  - Show current model")
	fmt.Println("  exit   - Exit the chat session")
	fmt.Println()
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
