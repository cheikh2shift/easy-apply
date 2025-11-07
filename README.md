# **🚀 AI Job Apply Helper**

This project automates the first steps of the job application process on LinkedIn's "Easy Apply" jobs by using a Tampermonkey userscript to extract job application HTML and sending it to a Go backend powered by the Gemini API. The AI analyzes the form and returns JavaScript code to fill out the form fields automatically.

## **🛠️ Prerequisites**

To run this project, you need two main components running simultaneously:

1. **Frontend (Browser):** A modern web browser with the Tampermonkey extension installed.  
2. **Backend (Server):** A Go environment running the backend server, exposed publicly via **ngrok**.

## **1\. Frontend Setup (Tampermonkey)**

The userscript, content\_script.js, runs in your browser to interact with the LinkedIn page.

### **1.1 Tampermonkey Requirement**

This script **must** be run using the **Tampermonkey** browser extension (available for Chrome, Firefox, Edge, etc.). Tampermonkey provides the necessary security grants (GM\_xmlhttpRequest, GM\_addElement) to communicate with your local API server without being blocked by LinkedIn's Content Security Policy (CSP).

### **1.2 Installation**

1. **Install Tampermonkey:** Install the Tampermonkey extension for your browser.  
2. **Create a New Script:**  
   * Click the Tampermonkey icon in your browser toolbar.  
   * Select Create a new script....  
3. **Paste Code:** Replace the default content with the complete code from your content\_script.js file (provided in previous steps).  
4. **Save:** Save the script (usually by clicking File \-\> Save or pressing Ctrl+S/Cmd+S).  
5. **Verification:** Ensure the script is enabled in the Tampermonkey Dashboard.

## **2\. Backend Setup (Go Server and Ngrok)**

The Go backend handles the connection to the Gemini API and contains the core logic for generating form-filling scripts.

### **2.1 Environment Variables**

The Go program requires the following environment variables to be set before running the server:

| Variable | Description |
| :---- | :---- |
| **GEMINI\_API\_KEY** | Your API key for accessing the Google Gemini API. |

**Example (.env file content):**

GEMINI\_API\_KEY="YOUR\_SECRET\_API\_KEY\_HERE"

### **2.2 Ngrok Forwarding**

Because the Tampermonkey script runs inside your browser and needs to communicate with your local Go server, you must use **ngrok** to create a secure, publicly accessible tunnel.

1. **Start your Go server.** (Assuming it runs on port 8080).  
2. **Run ngrok:** Execute the following command in your terminal:  
   ngrok http 8080

   * If your Go server runs on a different port, replace 8080 with the correct port number.  
3. **Update Script:** Ngrok will provide a public URL (e.g., https://xxxx.ngrok-free.app). **You MUST update this URL** in your Tampermonkey script (content\_script.js) at the following line:  
   // content\_script.js  
   const API\_ROOT \= "YOUR\_NGROK\_URL\_HERE";   
   // Example: const API\_ROOT \= "\[https://58d35aa75449.ngrok-free.app\](https://58d35aa75449.ngrok-free.app)";

## **3\. Usage**

1. **Start the Go server** locally.  
2. **Start ngrok** and confirm the public URL matches the API\_ROOT in your Tampermonkey script.  
3. **Navigate to LinkedIn Job Search:** Go to a LinkedIn jobs search results page (e.g., https://www.linkedin.com/jobs/search/?keywords=software%20developer).  
4. **Watch the Magic:** The script will automatically:  
   * Extract all job listings.  
   * Click the first job's detail panel.  
   * Click **"Easy Apply"** (if available).  
   * Send the modal content to the AI via your ngrok tunnel.  
   * Execute the returned script to fill out the form.  
   * If the job list is exhausted, it clicks the **"Next"** pagination button and restarts the process on the new page.