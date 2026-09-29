package test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aethergrok/pkg/grokrunner"
)

func createTestWav(t *testing.T) string {
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "test_audio.wav")

	sampleRate := uint32(8000)
	durationSeconds := uint32(1)
	numSamples := sampleRate * durationSeconds
	dataSize := numSamples * 2
	totalSize := uint32(36) + dataSize

	file, err := os.Create(wavPath)
	if err != nil {
		t.Fatalf("failed to create wav: %v", err)
	}
	defer file.Close()

	_, _ = file.Write([]byte("RIFF"))
	_ = binary.Write(file, binary.LittleEndian, totalSize)
	_, _ = file.Write([]byte("WAVE"))
	_, _ = file.Write([]byte("fmt "))
	_ = binary.Write(file, binary.LittleEndian, uint32(16))
	_ = binary.Write(file, binary.LittleEndian, uint16(1))
	_ = binary.Write(file, binary.LittleEndian, uint16(1))
	_ = binary.Write(file, binary.LittleEndian, sampleRate)
	_ = binary.Write(file, binary.LittleEndian, sampleRate*2)
	_ = binary.Write(file, binary.LittleEndian, uint16(2))
	_ = binary.Write(file, binary.LittleEndian, uint16(16))
	_, _ = file.Write([]byte("data"))
	_ = binary.Write(file, binary.LittleEndian, dataSize)

	samples := make([]byte, dataSize)
	_, _ = file.Write(samples)

	return wavPath
}

func TestTranscriptionWithResolvedConfig(t *testing.T) {
	baseURL, apiKey, candidateModels := grokrunner.ResolveTranscriptionConfig("")

	if apiKey == "" {
		t.Fatalf("expected non-empty apiKey from ~/.grok/config.toml")
	}
	if baseURL == "" {
		t.Fatalf("expected non-empty baseURL")
	}
	if len(candidateModels) == 0 {
		t.Fatalf("expected candidateModels")
	}

	wavPath := createTestWav(t)
	audioBytes, err := os.ReadFile(wavPath)
	if err != nil {
		t.Fatalf("failed to read wav: %v", err)
	}
	audioB64 := base64.StdEncoding.EncodeToString(audioBytes)

	targetURL := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	payload := map[string]interface{}{
		"model":  candidateModels[0],
		"stream": false,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "input_audio",
						"input_audio": map[string]string{
							"data":   audioB64,
							"format": "wav",
						},
					},
					{
						"type": "text",
						"text": "Transcribe this audio. Output only transcription.",
					},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("Endpoint unreachable in test environment: %v", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var resObj struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &resObj); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resObj.Choices) == 0 {
		t.Fatalf("expected choices in response")
	}

	t.Logf("Success! Transcribed result: %s", resObj.Choices[0].Message.Content)
}
