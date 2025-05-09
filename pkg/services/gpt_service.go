package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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

// CreateAssistantWithMetadata creates a new assistant based on YouTube video metadata
func CreateAssistantWithMetadata(initReq InitializeRequest) (string, error) {
	url := "https://api.openai.com/v1/assistants"

	requestBody := map[string]interface{}{
		"model":        "gpt-4o-mini",
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
func AskAssistantQuestion(videoID, assistantID, question string, timestamp int) (string, error) {
	threadManager, err := GetOrCreateThreadManager(assistantID)
	if err != nil {
		return "", fmt.Errorf("failed to get thread manager: %v", err)
	}

	// Pass the timestamp to AddMessageToThread
	err = threadManager.AddMessageToThread("user", question, assistantID, timestamp)
	if err != nil {
		return "", fmt.Errorf("failed to add message: %v", err)
	}

	// Run the assistant as usual
	return threadManager.RunAssistant(assistantID)
}

// GetOrCreateThreadManager retrieves the thread from Redis or creates a new one if it doesn't exist
func GetOrCreateThreadManager(assistantID string) (*ThreadManager, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// Generate Redis key using assistantID
	redisKey := fmt.Sprintf("thread_id:%s", assistantID)

	log.Printf("🔎 Checking Redis for thread ID: %s", redisKey)

	// Check if a thread ID already exists in Redis
	threadID, err := RedisClient.Get(Ctx, redisKey).Result()
	if err != nil {
		log.Println("❌ No thread found for Assistant:", assistantID)
		log.Println("🔵 Attempting to create a new thread...")

		// 🔹 Create a new thread if none exists
		threadID, err = createThread()
		if err != nil {
			return nil, fmt.Errorf("failed to create thread: %v", err)
		}

		// 🔹 Store the new thread ID in Redis
		err = RedisClient.Set(Ctx, redisKey, threadID, 168*time.Hour).Err()
		if err != nil {
			log.Printf("⚠️ Failed to store thread ID in Redis for Assistant: %s, Error: %v", assistantID, err)
			return nil, fmt.Errorf("failed to store thread ID in Redis: %v", err)
		}

		log.Printf("✅ Successfully created and stored thread ID: %s for Assistant: %s", threadID, assistantID)
	} else {
		log.Printf("✅ Found existing thread ID: %s for Assistant: %s", threadID, assistantID)
	}

	// Create a ThreadManager instance
	tm := &ThreadManager{ThreadID: threadID}
	threadManagers[assistantID] = tm
	return tm, nil
}

func createThread() (string, error) {
	url := "https://api.openai.com/v1/threads"
	requestBody := map[string]interface{}{}

	log.Println("🔵 Creating new thread...") // Debugging log

	body, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ Failed to marshal thread creation request: %v", err)
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ Failed to create HTTP request for thread creation: %v", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send thread creation request: %v", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("❌ Thread creation failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("failed to create thread: %s", string(bodyBytes))
	}

	var threadResp struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&threadResp)
	if err != nil {
		log.Printf("❌ Failed to decode thread creation response: %v", err)
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Thread created with ID %s", threadResp.ID)
	return threadResp.ID, nil
}

// Storing each interaction message in Redis
func (tm *ThreadManager) AddMessageToThread(role, content, assistantID string, timestamp int) error {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", tm.ThreadID)

	prompt := createPrompt(content, timestamp)
	log.Printf("📝 Adding message to thread. Role: %s, Assistant: %s", role, assistantID)

	requestBody := map[string]interface{}{
		"role":    role,
		"content": prompt,
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
		log.Printf("⚠️ Failed to add message to thread. StatusCode: %d, Response: %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("failed to add message to thread: %s", string(bodyBytes))
	}

	// ✅ Store both user and AI interactions under `assistant_id`
	interactionKey := fmt.Sprintf("interactions:%s", assistantID)
	prefix := "User: "
	if role == "assistant" {
		prefix = "Assistant: "
	}
	
	err = RedisClient.RPush(Ctx, interactionKey, prefix+prompt).Err()
	if err == nil {
		err = RedisClient.Expire(Ctx, interactionKey, 168*time.Hour).Err()
	}
	
	if err != nil {
		log.Printf("⚠️ Failed to store interaction in Redis for Assistant: %s, Error: %v", assistantID, err)
		return fmt.Errorf("failed to store interaction in Redis: %v", err)
	}

	log.Printf("✅ Interaction message stored in Redis for Assistant: %s", assistantID)
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

					// ✅ Store assistant's response in Redis under assistant-specific key
					err = RedisClient.RPush(Ctx, fmt.Sprintf("interactions:%s", assistantID), "Assistant: "+assistantResponse).Err()
					if err != nil {
						log.Printf("Failed to store assistant response in Redis for Assistant %s: %v", assistantID, err)
						return "", fmt.Errorf("failed to store assistant response in Redis: %v", err)
					}

					log.Printf("✅ Assistant response stored in Redis for Assistant: %s", assistantID)
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

	// Log the retrieval request
	log.Printf("Fetching messages from thread with ID: %s", tm.ThreadID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for thread message retrieval: %v", err)
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request to get thread messages: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Failed to fetch thread messages. StatusCode: %d, Response: %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("failed to get thread messages: %s", string(bodyBytes))
	}

	// Log the raw response body from OpenAI for debugging purposes
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("Raw thread messages response: %s", string(bodyBytes))

	var messagesResp struct {
		Data []Message `json:"data"`
	}
	err = json.Unmarshal(bodyBytes, &messagesResp)
	if err != nil {
		log.Printf("Failed to decode thread messages response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Log successful message retrieval
	log.Printf("Successfully fetched %d messages from thread with ID: %s", len(messagesResp.Data), tm.ThreadID)
	return messagesResp.Data, nil
}

func createPrompt(question string, timestamp int) string {
	// Format the timestamp as mm:ss
	// Create the prompt by appending the timestamp to the question
	return fmt.Sprintf("At the timestamp <%d>, user asks: %s, Give a response based on the context of the video around the timestamp. Don't include the timestamp in your response. Sound natural, and human", timestamp, question)
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

// GenerateSummary takes a video ID, retrieves the transcript from Redis, and returns a concise summary.
func GenerateSummary(transcript string) (string, error) {
	if transcript == "" {
		return "", fmt.Errorf("transcript is empty")
	}

	systemPrompt := "You are a professional assistant specializing in summarizing video content. Your summaries should be structured, concise, and focused on the key ideas, themes, and takeaways from the video. Exclude unnecessary details or repetitive information. Present the summary in a clear and organized format with headings if applicable."
	prompt := fmt.Sprintf("Please summarize the following video transcript. Focus on the key topics, main arguments, and actionable takeaways. Exclude irrelevant details, filler, or repetitive information, title of the video. Organize the summary into the following sections:\n\n1. Overview: Briefly introduce the video and its main purpose.\n2. Key Points: Outline the major ideas, concepts, or arguments presented.\n3. Conclusion: Summarize the overall message or conclusions drawn in the video.\n\nTranscript:\n%s", transcript)

	temperature := 0.8
	maxTokens := 16000
	response, err := CallGPT(prompt, systemPrompt, temperature, maxTokens)
	if err != nil {
		return "", fmt.Errorf("GPT call failed: %v", err)
	}
	return response, nil
}

func GenerateQuiz(transcript string) (map[string]interface{}, error) {
	if transcript == "" {
		return nil, fmt.Errorf("transcript is empty")
	}
	systemPrompt := "You are a helpful assistant that generates multiple choice questions given the full video transcript."
	prompt := fmt.Sprintf("Generate 10 multiple-choice questions in structured JSON format based on the following transcript. Each question must have exactly one correct answer. If multiple valid answers are mentioned in the transcript, only include one of them as part of the options. The questions should be based on the transcript and should not be outside the transcript. Ensure that the answer field exactly matches one of the provided options. Do NOT add any letters like 'A, B, C, D' before the options. Just provide the options. Transcript:\n\n%s", transcript)
	response, err := CallGPT2(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("GPT call failed: %v", err)
	}
	return response, nil
}

func CallGPT(prompt string, systemPrompt string, temperature float64, maxTokens int) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini", // or gpt-3.5-turbo for lower cost
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},

		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GPT API call failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GPT API error: %s", string(body))
	}

	var gptResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gptResponse); err != nil {
		return "", fmt.Errorf("failed to decode GPT response: %v", err)
	}

	// Extract the summary from the assistant's message
	return gptResponse["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string), nil
}

func CallGPT2(prompt string, systemPrompt string) (map[string]interface{}, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	// Define the request body
	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini", // Replace with another model if required
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name": "quiz_generation",
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"questions": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"text": map[string]interface{}{
										"type": "string",
									},
									"timestamp": map[string]interface{}{
										"type": "string", // Removed "format" field
									},
									"options": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"option": map[string]interface{}{
													"type": "string",
												},
												"explanation": map[string]interface{}{
													"type": "string",
												},
											},
											"required":             []string{"option", "explanation"},
											"additionalProperties": false,
										},
									},
									"answer": map[string]interface{}{
										"type": "string",
									},
								},
								"required":             []string{"text", "timestamp", "options", "answer"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"questions"},
					"additionalProperties": false,
				},
				"strict": true,
			},
		},
		"temperature": 0.7,
		"max_tokens":  10000, // Adjust based on expected response size
	}

	// Marshal the request body into JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	// Send the request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GPT API call failed: %v", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GPT API error: %s", string(body))
	}

	// Parse the response
	var gptResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gptResponse); err != nil {
		return nil, fmt.Errorf("failed to decode GPT response: %v", err)
	}

	return gptResponse, nil
}
