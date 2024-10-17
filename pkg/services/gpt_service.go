package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type ThreadManager struct {
	ThreadID string
}

var (
	threadManagers = make(map[string]*ThreadManager)
	mutex          sync.Mutex
)

// Define InitializeRequest in services.go
type InitializeRequest struct {
	SystemInstructions string `json:"system_instructions"`
	VideoID            string `json:"video_id"`
	Title              string `json:"title"`
	Channel            string `json:"channel"`
	Transcript         string `json:"transcript"`
}

// Initialize the OpenAI client and load the API key
func InitOpenAIClient() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// CreateAssistantWithMetadata creates a new assistant based on YouTube video metadata
func CreateAssistantWithMetadata(initReq InitializeRequest) (string, error) {
	url := "https://api.openai.com/v1/assistants"

	requestBody := map[string]interface{}{
		"model":        "gpt-4-turbo",
		"name":         initReq.VideoID,
		"instructions": fmt.Sprintf("You are a helpful assistant for the video titled '%s' by '%s'. Here is the transcript: %s", initReq.Title, initReq.Channel, initReq.Transcript),
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create assistant: %s", string(bodyBytes))
	}

	var createResp struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return createResp.ID, nil
}

// AskAssistantQuestion adds a question to the thread and gets a response
func AskAssistantQuestion(assistantID, question string) (string, error) {
	threadManager, err := GetOrCreateThreadManager(assistantID)
	if err != nil {
		return "", fmt.Errorf("failed to get thread manager: %v", err)
	}

	err = threadManager.AddMessageToThread("user", question)
	if err != nil {
		return "", fmt.Errorf("failed to add message: %v", err)
	}

	return threadManager.RunAssistant(assistantID)
}

func GetOrCreateThreadManager(assistantID string) (*ThreadManager, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if tm, exists := threadManagers[assistantID]; exists {
		return tm, nil
	}

	threadID, err := createThread()
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %v", err)
	}

	tm := &ThreadManager{ThreadID: threadID}
	threadManagers[assistantID] = tm
	return tm, nil
}

func createThread() (string, error) {
	url := "https://api.openai.com/v1/threads"
	requestBody := map[string]interface{}{}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create thread: %s", string(bodyBytes))
	}

	var threadResp struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&threadResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return threadResp.ID, nil
}

func (tm *ThreadManager) AddMessageToThread(role, content string) error {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", tm.ThreadID)

	requestBody := map[string]interface{}{
		"role":    role,
		"content": content,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to add message to thread: %s", string(bodyBytes))
	}

	return nil
}

func (tm *ThreadManager) RunAssistant(assistantID string) (string, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", tm.ThreadID)

	requestBody := map[string]interface{}{
		"assistant_id": assistantID,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to run assistant: %s", string(bodyBytes))
	}

	var runResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&runResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Poll for completion
	for {
		time.Sleep(2 * time.Second)
		status, err := tm.GetRunStatus(runResp.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get run status: %v", err)
		}

		if status == "completed" {
			messages, err := tm.GetThreadMessages()
			if err != nil {
				return "", fmt.Errorf("failed to get thread messages: %v", err)
			}

			// Return the assistant message
			for _, msg := range messages {
				if msg.Role == "assistant" {
					var assistantResponse string
					for _, fragment := range msg.Content {
						if fragment.Type == "text" && fragment.Text != nil {
							assistantResponse += fragment.Text.Value
						}
					}
					return assistantResponse, nil
				}
			}
			return "", fmt.Errorf("no assistant message found")
		}
	}
}

func (tm *ThreadManager) GetRunStatus(runID string) (string, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs/%s", tm.ThreadID, runID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get run status: %s", string(bodyBytes))
	}

	var runStatus struct {
		Status string `json:"status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&runStatus)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return runStatus.Status, nil
}

func (tm *ThreadManager) GetThreadMessages() ([]Message, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", tm.ThreadID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get thread messages: %s", string(bodyBytes))
	}

	var messagesResp struct {
		Data []Message `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&messagesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return messagesResp.Data, nil
}

type TextContent struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations"` // You can adjust this depending on what the annotations are
}

type ContentFragment struct {
	Type string       `json:"type"`
	Text *TextContent `json:"text,omitempty"` // Only include text if it's of type text
	// You can include other content types here like image, video, etc.
}

type Message struct {
	ID      string            `json:"id"`
	Role    string            `json:"role"`
	Content []ContentFragment `json:"content"` // Content is now a list of fragments
}
