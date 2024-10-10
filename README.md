# YouTube Learning Mode AI Service

This Go-based microservice provides AI-powered responses using OpenAI's GPT-4 API. It processes video context, manages AI conversations, and generates answers based on user questions about YouTube videos. This service is a key component in the YouTube Learning Mode project, enabling dynamic and interactive learning experiences by leveraging video transcripts and snapshots.

## Features

- **AI Session Initialization**: Sets up AI sessions using the video context (title, description, transcript).
- **Question Handling**: Accepts user questions and generates responses using OpenAI's GPT-4.
- **Redis Integration**: Caches AI conversation history to ensure continuity between user interactions.
- **Future Enhancements**: Potential integration with other AI models and switching to OpenAI's assistant for more complex interactions.

## Tech Stack

- **Go**: Core language for the service.
- **OpenAI API**: For generating AI responses.
- **Redis**: For storing conversation history.
- **Gorilla Mux**: Router for handling HTTP requests.

## Prerequisites

Make sure you have the following installed:

- **Go**: Version 1.16 or higher.
- **Docker & Docker Compose**: For containerization.
- **OpenAI API Key**: Required for accessing GPT-4.

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/AnasKhan0607/Youtube-Learning-Mode-Ai-Service.git
cd Youtube-Learning-Mode-Ai-Service
```

### 2. Set Up Environment Variables

Create a `.env` file in the root directory to store your environment variables:

```bash
OPENAI_API_KEY=your_openai_api_key
```

> **Important:** Do not commit the `.env` file to version control as it contains sensitive information.

### 3. Run with Docker

To build and run the service using Docker:

```bash
docker build -t ai-service .
docker run -p 8082:8082 --env-file .env ai-service
```

Or use Docker Compose with other services:

```bash
docker-compose up --build
```

This will start the AI Service along with the other microservices in a shared network.

## Running the Service Without Docker (Optional)

If you want to run the service without Docker, make sure Redis is running locally and your `.env` file is properly set up:

```bash
go run cmd/main.go
```

The server will start on `http://localhost:8082`.

## API Endpoints

### 1. Initialize AI Session

- **Endpoint**: `POST /ai/init-session`
- **Description**: Sets up a new AI session using the video context.
- **Request Body**:

  ```json
  {
    "video_id": "VIDEO_ID",
    "title": "Video Title",
    "channel": "Channel Name",
    "transcript": ["0.00: Transcript line 1", "0.05: Transcript line 2"]
  }
  ```

- **Response**:

  ```json
  {
    "message": "GPT session initialized"
  }
  ```

### 2. Ask AI Question

- **Endpoint**: `POST /ai/ask-question`
- **Description**: Accepts a user question related to the video and returns an AI-generated response.
- **Request Body**:

  ```json
  {
    "video_id": "VIDEO_ID",
    "user_question": "What is the main topic of the video?"
  }
  ```

- **Response**:

  ```json
  {
    "response": "The video discusses..."
  }
  ```

## Project Structure

```
├── cmd/
│   └── main.go                 # Entry point of the AI service
├── pkg/
│   ├── handlers/
│   │   └── handlers.go         # Handles HTTP requests for AI session and question handling
│   ├── services/
│   │   ├── gpt_service.go      # Manages GPT-4 interactions and initializes OpenAI client
│   │   ├── redis_service.go    # Handles Redis connections and operations
│   └── router/
│       └── router.go           # Defines API routes (if applicable)
├── .env                        # Environment variables (DO NOT COMMIT)
├── Dockerfile                  # Docker configuration for containerizing the service
├── go.mod                      # Go module dependencies
├── go.sum                      # Checksums for Go modules
└── README.md                   # Project documentation
```

## Dependencies

- **Go Redis**: For connecting to the Redis database.
- **OpenAI Go SDK**: For communicating with the OpenAI GPT-4 API.
- **Gorilla Mux**: For HTTP request routing.

## Future Enhancements

- **Integration with OpenAI Assistant**: To provide a more dynamic and conversational user experience.
- **Custom Models**: Incorporate additional AI models or fine-tuned versions for more specific responses.
- **Snapshot Analysis**: Leverage video snapshots for deeper context understanding.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.

2. Create a new branch:

   ```bash
   git checkout -b feature/YourFeature
   ```

3. Commit your changes:

   ```bash
   git commit -am 'Add some feature'
   ```

4. Push to the branch:

   ```bash
   git push origin feature/YourFeature
   ```

5. Open a Pull Request.

---

### Important Notes:

- Ensure that Redis is running before starting the AI service to maintain conversation history.
- Rate limits for the OpenAI API may impact response times. Monitor usage through your OpenAI account.
