package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alexhokl/helper/git"
	"github.com/alexhokl/helper/iohelper"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

type generateOptions struct {
	modelName string
}

var generateOpts generateOptions

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate git commit comment using models from Ollama",
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	flags := generateCmd.Flags()
	flags.StringVarP(&generateOpts.modelName, "model", "m", "llama3.1:8b", "Name of the model used to generate comment")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	hasStagedFiles, err := git.HasStagedFiles()
	if err != nil {
		return fmt.Errorf("unable to check if there are staged files: %v", err)
	}
	if !hasStagedFiles {
		return fmt.Errorf("no staged files")
	}

	// create a stream to temp file
	diffContentFile, err := os.CreateTemp("", "git-comment")
	if err != nil {
		return fmt.Errorf("unable to create temp file: %v", err)
	}
	defer diffContentFile.Close()

	git.DiffToStream(true, diffContentFile)

	// fmt.Println("Diff content is saved to", diffContentFile.Name())

	diffContent, err := iohelper.ReadStringFromFile(diffContentFile.Name())
	if err != nil {
		return fmt.Errorf("unable to read diff content: %v", err)
	}

	ollamaClient, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("unable to create Ollama client: %v", err)
	}
	request := &api.GenerateRequest{
		Model:   generateOpts.modelName,
		Stream:  new(bool), // false
		System: "You are an expert in programming and expert user in using git",
		Prompt: fmt.Sprintf(
					"Given the following output of `git diff`. Generate git a commit comment. Please include only the comment without extra explanation.\n\n%s",
					diffContent,
				),
		Options: map[string]interface{}{
			"temperature": 0.0,   // avoid randomness
		},
	}
	responseFunc := func(response api.GenerateResponse) error {
		cleanedResponse := cleanUpResponse(response.Response)
		fmt.Println(cleanedResponse)
		return nil
	}

	if err := ollamaClient.Generate(ctx, request, responseFunc); err != nil {
		return fmt.Errorf("unable to generate comment: %v", err)
	}

	return nil
}

func cleanUpResponse(response string) string {
	return strings.Trim(response, "\"")
}
