package convert

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func RecognizeWithGemini(imageData []byte) (string, string) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		return "", "Failed to create client"
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash-exp")

	mimeType := detectImageType(imageData)

	resp, err := model.GenerateContent(ctx,
		genai.Text("Extract the mathematical formula or equation from this image. Convert it to proper LaTeX format. Return ONLY the LaTeX code without any explanation, markdown formatting, or additional text."),
		genai.ImageData(mimeType, imageData))
	if err != nil {
		return "", "Generation failed"
	}

	if len(resp.Candidates) == 0 {
		return "", "No candidates"
	}

	if resp.Candidates[0].Content == nil {
		return "", "No content"
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "No parts"
	}

	text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	text = strings.ReplaceAll(text, "```latex", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.TrimSpace(text)

	if text == "" {
		return "", "Empty response"
	}

	return text, ""
}

func detectImageType(data []byte) string {
	contentType := http.DetectContentType(data)
	if strings.HasPrefix(contentType, "image/") {
		return contentType
	}
	return "image/jpeg"
}
