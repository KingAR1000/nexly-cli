package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Message struct {
	Role    string
	Content string
}

type StreamCallback func(string)

type Provider interface {
	Name() string
	SendMessage(ctx context.Context, messages []Message, streamCallback StreamCallback) error
	GetModels() []string
}

type SimpleProvider struct {
	name   string
	apiKey string
	model  string
	apiURL string
}

func NewSimpleProvider(provider, apiKey, model string) *SimpleProvider {
	var apiURL string
	switch provider {
	case "openai":
		apiURL = "https://api.openai.com/v1/chat/completions"
	case "anthropic":
		apiURL = "https://api.anthropic.com/v1/messages"
	case "google":
		apiURL = "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":streamGenerateContent?alt=sse"
	case "openrouter":
		apiURL = "https://openrouter.ai/api/v1/chat/completions"
	case "nvidia":
		apiURL = "https://integrate.api.nvidia.com/v1/chat/completions"
	default:
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	return &SimpleProvider{
		name:   provider,
		apiKey: apiKey,
		model:  model,
		apiURL: apiURL,
	}
}

func (p *SimpleProvider) Name() string {
	return p.name
}

func (p *SimpleProvider) GetModels() []string {
	switch p.name {
	case "openai":
		return []string{"gpt-4", "gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"}
	case "anthropic":
		return []string{"claude-3-5-sonnet-20241022", "claude-3-opus-20240229"}
	case "google":
		return []string{"gemini-1.5-pro", "gemini-1.5-flash"}
	case "openrouter":
		return []string{"openai/gpt-4", "anthropic/claude-3.5-sonnet"}
	case "nvidia":
		return []string{"nvidia/llama-3.1-nemotron-70b-instruct"}
	default:
		return []string{"gpt-4"}
	}
}

func (p *SimpleProvider) SendMessage(ctx context.Context, messages []Message, streamCallback StreamCallback) error {
	if p.apiKey == "" {
		return fmt.Errorf("API key not configured for provider: %s", p.name)
	}

	reqBody := p.buildRequestBody(messages)
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	p.setHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return p.handleResponse(resp, streamCallback)
}

func (p *SimpleProvider) buildRequestBody(messages []Message) map[string]interface{} {
	switch p.name {
	case "google":
		return map[string]interface{}{
			"contents": formatGoogleMessages(messages),
			"generationConfig": map[string]interface{}{
				"temperature":     0.7,
				"maxOutputTokens": 4096,
			},
		}
	case "anthropic":
		var systemMsg string
		var userMsgs []Message
		for _, m := range messages {
			if m.Role == "system" {
				systemMsg = m.Content
			} else {
				userMsgs = append(userMsgs, m)
			}
		}
		body := map[string]interface{}{
			"model":      p.model,
			"messages":   userMsgs,
			"stream":     true,
			"max_tokens": 4096,
		}
		if systemMsg != "" {
			body["system"] = systemMsg
		}
		return body
	default:
		return map[string]interface{}{
			"model":       p.model,
			"messages":    messages,
			"stream":      true,
			"temperature": 0.7,
		}
	}
}

func (p *SimpleProvider) setHeaders(req *http.Request) {
	switch p.name {
	case "anthropic":
		req.Header.Set("x-api-key", p.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("Content-Type", "application/json")
	case "google":
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")
	case "openrouter":
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HTTP-Referer", "https://nexlycode.vercel.app")
		req.Header.Set("X-Title", "Nexly")
	case "nvidia":
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")
	default:
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")
	}
}

func (p *SimpleProvider) handleResponse(resp *http.Response, streamCallback StreamCallback) error {
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)

	switch p.name {
	case "anthropic":
		return p.handleAnthropicStream(reader, streamCallback)
	case "google":
		return p.handleGoogleStream(reader, streamCallback)
	default:
		return p.handleOpenAIStream(reader, streamCallback)
	}
}

func (p *SimpleProvider) handleOpenAIStream(reader *bufio.Reader, streamCallback StreamCallback) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var response struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &response); err != nil {
			continue
		}

		if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
			streamCallback(response.Choices[0].Delta.Content)
		}
	}

	return nil
}

func (p *SimpleProvider) handleAnthropicStream(reader *bufio.Reader, streamCallback StreamCallback) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var response struct {
			Delta struct {
				Text string `json:"text"`
			} `json:"delta"`
		}

		if err := json.Unmarshal([]byte(data), &response); err != nil {
			continue
		}

		if response.Delta.Text != "" {
			streamCallback(response.Delta.Text)
		}
	}

	return nil
}

func (p *SimpleProvider) handleGoogleStream(reader *bufio.Reader, streamCallback StreamCallback) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var response struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}

		if err := json.Unmarshal([]byte(data), &response); err != nil {
			continue
		}

		if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
			streamCallback(response.Candidates[0].Content.Parts[0].Text)
		}
	}

	return nil
}

func formatGoogleMessages(messages []Message) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, m := range messages {
		role := m.Role
		if role == "system" {
			role = "model"
		}
		result = append(result, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": m.Content},
			},
		})
	}
	return result
}
