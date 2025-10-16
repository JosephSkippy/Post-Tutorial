package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type UpdatePostPayload struct {
	Title   *string `json:"title" `
	Content *string `json:"content"`
}

func updatePost(postID int, p UpdatePostPayload, wg *sync.WaitGroup) {
	defer wg.Done()

	// Construct the URL for the update endpoint
	url := fmt.Sprintf("http://localhost:3000/v1/posts/%d", postID)

	// Create the JSON payload
	b, _ := json.Marshal(p)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers as needed, for example:
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Update response status:", resp.Status)
	fmt.Printf("Response body: %s\n", string(body))
}

func main() {
	var wg sync.WaitGroup

	// Assuming the post ID to update is 1
	postID := 3

	numRequests := 10 // More requests = higher chance of collision

	wg.Add(numRequests * 2)

	for i := 0; i < numRequests; i++ {
		content := fmt.Sprintf("NEW content FROM USER A TEST %d", i)
		title := fmt.Sprintf("NEW title FROM USER A TEST %d", i)
		go updatePost(postID, UpdatePostPayload{Title: &title}, &wg)
		go updatePost(postID, UpdatePostPayload{Content: &content}, &wg)
	}
	wg.Wait()
}
