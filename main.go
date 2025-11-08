package main

import (
	"bytes" // New import for handling JSON request body
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// --- CONFIGURATION ---
// !! REPLACE THIS WITH YOUR ACTUAL RESUME TEXT !!
var resumeText = `
---------------------
Cheikh Diagne Seck
Backend Developer
Dakar, Senegal
+221763199588

About
I'm an experienced Individual with a robust background in software engineering and consulting. I've held
roles such as Senior Software Engineer and Consultant, cultivating expertise in Docker, Go, JavaScript,
Git, and more. I've Developed and maintained software solutions, integrated with legacy systems, and
built services for different types of industries.

Technical Skills
Top 3 skills: Golang・9 yrs, JavaScript・10 yrs, MongoDB・10 yrs

Work experience skills:
Simsol Technologies: MongoDB, PostgreSQL, CouchDB, TypeScript, React, JavaScript,
Flutter, Golang, Vue.js
Ardan Labs: Git, Markdown, Golang
dHub: Java, Node.js, MongoDB, Angular, TypeScript, JavaScript, Swift, Apache cordova,
Express.js
Frantzi: PHP, WooCommerce, SAP, Microsoft SQL Server
Orkiv: PHP, Java, MySQL, MongoDB, Angular, Mongoose, TypeScript, JavaScript, Objective-
C, Golang, Apache tomcat, Unity

Work Experience
Senior Software Engineer/Product Manager - Simsol Technologies (Aug 2022 - Aug 2025)
Developed a library for rapid web app development and deployment (under 2 weeks).
Built integrations with Twilio, Outlook, and WhatsApp for enhanced user interaction.
Implemented real-time dashboards using WebSockets.
Created a parcel tracking and management platform.
Wrote code for automated report, spreadsheet, and PDF generation.
Copywriter - Ardan Labs (Dec 2022 - Aug 2023)
Write copies for social media content.
Contribute to technical blogs.
Contribute to digital marketing campaigns.
Software Developer - dHub (Sep 2017 - Sep 2021)
Design and build backend infrastructure for warehouse management system.
Build web interface and mobile application.
Published mobile application to Apple store and the playstore.
Software developer - Frantzi (Oct 2017 - Apr 2018)
Page 1 of 2Build Woocommerce plugin that sync'd SAP stock information with Woocommerce
Perform various tweaks to their website.
Founder - Orkiv (May 2014 - Aug 2017)
Develop SaaS that specialized in inventory management, website building and blog
managment.
Built a social media platform that was leveraged, by our clients, to reach their customers.
Contribute to marketing campaigns.
Wrote web scrapers that gathered information about prospective leads.
Build custom web applications for clients.

Projects
Go Blog (2022) - https://cheikhhseck.medium.com/
I use this as an outlet to share my thoughts about Go. The blog consists of various tutorials
that aim to teach readers how to build different types of programs with the Go programming
language.

Education
Bryant University - Bachelor's degree・Information technology and computer science (Aug 2012 -
May 2014)
---------------------
`

// Model configuration constants
var localModelName = "gemma3:270m"

const geminiModelName = "gemini-2.5-flash"
const ollamaAPIUrl = "http://localhost:11434/api/generate" // Ollama's default API endpoint

// Struct for incoming JSON request from the browser extension
type AnalyzeRequest struct {
	HTML string `json:"html"`
}

// Struct for outgoing JSON response to the browser extension
type AnalyzeResponse struct {
	JSCode string `json:"jsCode"`
}

// Struct for making POST request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Struct for parsing the JSON response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	// Other fields like 'model', 'created_at', 'done' are ignored
}

type AppServer struct {
	geminiModel   *genai.GenerativeModel // Only initialized if !useLocalModel
	useLocalModel bool
}

// aiapplyHandler handles the core logic
func (s *AppServer) aiapplyHandler(w http.ResponseWriter, r *http.Request) {
	// --- CORS Handling (CRITICAL for ngrok) ---
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS pre-flight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Decode the incoming request (HTML from the extension)
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.HTML == "" {
		http.Error(w, "HTML content is empty", http.StatusBadRequest)
		return
	}

	// 2. Build the prompt for the AI
	prompt := buildPrompt(req.HTML)

	// 3. Call AI
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	var rawResponse string
	var err error

	if s.useLocalModel {
		log.Printf("Sending prompt to local Ollama API (%s) at %s...", localModelName, ollamaAPIUrl)

		// --- DEBUGGING: LOG CURL COMMAND ---
		ollamaReqBody := OllamaRequest{
			Model:  localModelName,
			Prompt: "---",
			Stream: false,
		}
		jsonData, _ := json.Marshal(ollamaReqBody)

		// Escape quotes in the JSON string for safe inclusion in the shell command
		jsonString := strings.ReplaceAll(string(jsonData), "\"", "\\\"")

		// Log the reusable curl command
		curlCommand := fmt.Sprintf(
			"curl -X POST %s -d \"%s\"",
			ollamaAPIUrl,
			jsonString,
		)
		log.Println("\n--- DEBUG CURL COMMAND ---")
		log.Println(curlCommand)
		log.Println("--------------------------")
		// ------------------------------------

		rawResponse, err = s.generateContentLocal(ctx, prompt)
	} else {
		log.Printf("Sending prompt to Gemini API (%s)...", geminiModelName)
		resp, geminiErr := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
		if geminiErr != nil {
			err = geminiErr
		} else {
			rawResponse = extractRawTextFromGemini(resp)
		}
	}

	if err != nil {
		log.Printf("Error generating content: %v", err)
		http.Error(w, "Failed to get response from AI", http.StatusInternalServerError)
		return
	}

	// 4. Extract and clean the JavaScript code from the raw response
	jsCode := extractJSFromRawText(rawResponse)
	if jsCode == "" {
		log.Println("AI did not return any usable content.")
		http.Error(w, "AI did not return usable code", http.StatusInternalServerError)
		return
	}

	log.Println("Received JS code from AI. Sending to client.", jsCode)

	// 5. Send the JavaScript code back to the extension
	res := AnalyzeResponse{JSCode: jsCode}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// generateContentLocal executes an HTTP POST request to the local Ollama API endpoint.
func (s *AppServer) generateContentLocal(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  localModelName,
		Prompt: prompt,
		Stream: false, // Wait for the entire response
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Create the HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", ollamaAPIUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request to Ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	// Set a client timeout slightly shorter than the context timeout for robustness
	client := &http.Client{ /*Timeout: 90 * time.Second*/ }
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama API at %s. Is Ollama running?: %w", ollamaAPIUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// The raw response text is inside the 'response' field of the Ollama response JSON
	return ollamaResp.Response, nil
}

func buildPrompt(modalHTML string) string {
	// This prompt is highly specific to the task.
	return fmt.Sprintf(`
You are an expert automation AI. Your task is to analyze a user's resume and the HTML of a job application modal step and generate *only* JavaScript code to fill in the form and proceed to the next step.

--- USER RESUME ---
%s
--- END RESUME ---

--- MODAL HTML CONTENT ---
%s
--- END HTML ---

INSTRUCTIONS:
1.  Analyze the resume and the MODAL HTML (which is only a single step/slide of the application).
2.  Identify and fill all form fields (input, textarea, select) using the resume information.
3.  Match fields robustly using "[name=\"...\"]", "[aria-label=\"...\"]", or "label[for=\"...\"]". Example: document.querySelector('[name="firstName"]').value = 'John';
4.  **Action Logic (Priority Order):**
    a. **If a "Submit application" button is found:**
        i. Find ALL checkboxes in the modal and set their "checked" property to "false" to ensure no companies are followed.
        ii. Generate code to click the "Submit application" button (using "[data-live-test-easy-apply-submit-button]").
    b. **If a "Review your application" button is found:**
        i. Generate code to click the "Review your application" button (using "[data-live-test-easy-apply-review-button]").
    c. **If a "Next" or "Continue to next step" button is found:**
        i. Generate code to click the "Next" button (using "[data-easy-apply-next-button]").
5.  Do NOT include javascript or any other text, just the executable JavaScript code.
6.  If you cannot find a match for a field, omit code for it.
7. Some fields appear to require the user clicking it for the input to be recognized, take this into account

JAVASCRIPT CODE:
`, resumeText, modalHTML)
}

// Extracts the raw text string from the Gemini response.
func extractRawTextFromGemini(resp *genai.GenerateContentResponse) string {
	if resp == nil || resp.Candidates == nil || len(resp.Candidates) == 0 {
		return ""
	}
	part := resp.Candidates[0].Content.Parts[0]
	if txt, ok := part.(genai.Text); ok {
		return string(txt)
	}
	return ""
}

// Cleans the raw text response by removing markdown wrappers (used for both models).
func extractJSFromRawText(rawText string) string {
	// Clean the response
	jsCode := rawText
	// Remove markdown backticks if the AI includes them
	jsCode = strings.TrimSpace(jsCode)

	// Handle common markdown wrappers (```javascript, ```)
	if strings.HasPrefix(jsCode, "```javascript") {
		jsCode = strings.TrimPrefix(jsCode, "```javascript")
		jsCode = strings.TrimSuffix(jsCode, "```")
	} else if strings.HasPrefix(jsCode, "```") {
		jsCode = strings.TrimPrefix(jsCode, "```")
		jsCode = strings.TrimSuffix(jsCode, "```")
	}

	// Final trim
	return strings.TrimSpace(jsCode)
}

func main() {
	// Command-line flag definition
	useLocalModelPtr := flag.Bool("local-model", false, "Use local Ollama model (deepseek-r1) instead of Gemini.")
	flag.Parse()

	localModel := os.Getenv("LLAMA_MODEL")

	if localModel != "" {
		localModelName = localModel
	}

	// 1. Conditional Configuration
	if !*useLocalModelPtr {
		// --- GEMINI SETUP ---
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable not set. Please set it or run with '-local-model'.")
		}
	} else {
		// --- OLLAMA SETUP ---
		log.Println("--- WARNING: Using local Ollama model 'deepseek-r1' via API endpoint ---")
		log.Printf("Ensure Ollama is running and accessible at %s.", ollamaAPIUrl)
	}

	// Load Resume File Path if set
	resumeFile := os.Getenv("RESUME_FILE_PATH")
	if resumeFile != "" {
		resData, err := os.ReadFile(resumeFile)
		if err != nil {
			log.Fatalf("Error finding resume data at %s: %s", resumeFile, err)
		}
		resumeText = string(resData)
	}

	// 2. Initialize Client Conditionally
	var geminiModel *genai.GenerativeModel
	if !*useLocalModelPtr {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
		if err != nil {
			log.Fatalf("Failed to create Gemini client: %v", err)
		}
		defer client.Close()
		geminiModel = client.GenerativeModel(geminiModelName)
	}

	// Set up the server state
	s := &AppServer{
		geminiModel:   geminiModel,
		useLocalModel: *useLocalModelPtr,
	}

	// 3. Set up HTTP server
	http.HandleFunc("/api/analyze", s.aiapplyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Printf("Make sure to run 'ngrok http %s' and update your content script", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
