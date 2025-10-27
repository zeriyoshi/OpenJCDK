package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"golang.org/x/net/html"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// ----------------------------------------------------------------
// Message Manager
// ----------------------------------------------------------------
type MessageData struct {
	Tweet string
	Alt   string
}

func parseMessage(description string) *MessageData {
	if description == "" {
		return nil
	}

	data := &MessageData{}
	doc, err := html.Parse(strings.NewReader(description))
	if err != nil {
		return nil
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "tweet" && n.FirstChild != nil {
				data.Tweet = n.FirstChild.Data
			} else if n.Data == "alt" && n.FirstChild != nil {
				data.Alt = n.FirstChild.Data
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	if data.Tweet != "" {
		return data
	}
	return nil
}

func generateMessage(description string) (string, string) {
	data := parseMessage(description)
	if data != nil {
		return data.Tweet, data.Alt
	}

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst)
	hour := now.Hour()

	footer := getEnvDefault("MESSAGE_FOOTER", "#邪神ちゃん今日の１枚 をどうぞ。")
	var header string

	if hour >= 4 && hour < 11 {
		header = getEnvDefault("MESSAGE_HEADER_MORNING", "フォロワーの皆さま、おはようございます！ ")
	} else if hour >= 11 && hour < 15 {
		header = getEnvDefault("MESSAGE_HEADER_NOON", "フォロワーの皆さま、ランチタイムです！ ")
	} else {
		header = getEnvDefault("MESSAGE_HEADER_NIGHT", "フォロワーの皆さま、今日も１日おつかれさまでした。お休み前に ")
	}

	return header + footer, ""
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ----------------------------------------------------------------
// Google Drive
// ----------------------------------------------------------------
func getRandomImageFromDrive(ctx context.Context) ([]byte, *drive.File, error) {
	serviceAccountKey := os.Getenv("GOOGLE_DRIVE_SERVICE_ACCOUNT_KEY")
	directoryID := os.Getenv("GOOGLE_DRIVE_DIRECTORY_ID")

	if serviceAccountKey == "" {
		return nil, nil, fmt.Errorf("GOOGLE_DRIVE_SERVICE_ACCOUNT_KEY environment variable is not set")
	}
	if directoryID == "" {
		return nil, nil, fmt.Errorf("GOOGLE_DRIVE_DIRECTORY_ID environment variable is not set")
	}

	driveService, err := drive.NewService(ctx, option.WithCredentialsJSON([]byte(serviceAccountKey)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	var fileIDs []string
	pageToken := ""

	for {
		query := fmt.Sprintf("'%s' in parents and trashed = false", directoryID)
		call := driveService.Files.List().Q(query).Spaces("drive").Fields("nextPageToken, files(id)")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		response, err := call.Do()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range response.Files {
			fileIDs = append(fileIDs, file.Id)
		}

		pageToken = response.NextPageToken
		if pageToken == "" {
			break
		}
	}

	if len(fileIDs) == 0 {
		return nil, nil, fmt.Errorf("no files found in directory")
	}

	rand.Seed(time.Now().UnixNano())
	selectedID := fileIDs[rand.Intn(len(fileIDs))]

	resp, err := driveService.Files.Get(selectedID).Download()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	metadata, err := driveService.Files.Get(selectedID).Fields("name, mimeType, description").Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return fileData, metadata, nil
}

// ----------------------------------------------------------------
// X API v2
// ----------------------------------------------------------------
func createOAuth1Client() *http.Client {
	config := oauth1.NewConfig(
		os.Getenv("TWITTER_CONSUMER_KEY"),
		os.Getenv("TWITTER_CONSUMER_SECRET"),
	)
	token := oauth1.NewToken(
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
	)
	return config.Client(oauth1.NoContext, token)
}

func uploadMedia(client *http.Client, mediaData []byte, mimeType string, altText string) (string, error) {
	if strings.HasPrefix(mimeType, "image/") {
		return simpleUpload(client, mediaData, mimeType, altText)
	}
	return chunkedUpload(client, mediaData, mimeType, altText)
}

func simpleUpload(client *http.Client, mediaData []byte, mimeType string, altText string) (string, error) {
	url := "https://api.x.com/2/media/upload"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("media_type", mimeType)
	writer.WriteField("media_category", "tweet_image")

	if altText != "" {
		writer.WriteField("alt_text", altText)
	}

	part, err := writer.CreateFormFile("media", "upload")
	if err != nil {
		return "", err
	}

	part.Write(mediaData)

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyText, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("media upload failed: %s - %s", resp.Status, string(bodyText))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			ID       string `json:"id"`
			MediaKey string `json:"media_key"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Data.ID == "" {
		return "", fmt.Errorf("media ID not found in response: %s", string(bodyBytes))
	}

	return result.Data.ID, nil
}

func chunkedUpload(client *http.Client, mediaData []byte, mimeType string, altText string) (string, error) {
	initBody := &bytes.Buffer{}
	initWriter := multipart.NewWriter(initBody)
	initWriter.WriteField("total_bytes", fmt.Sprintf("%d", len(mediaData)))
	initWriter.WriteField("media_type", mimeType)
	initWriter.WriteField("media_category", "tweet_video")
	if altText != "" {
		initWriter.WriteField("alt_text", altText)
	}
	initWriter.Close()

	req, _ := http.NewRequest("POST", "https://api.x.com/2/media/upload/init", initBody)
	req.Header.Set("Content-Type", initWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("media upload init failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var initResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &initResult); err != nil {
		return "", err
	}

	if initResult.Data.ID == "" {
		return "", fmt.Errorf("media ID not found in init response")
	}

	chunkSize := 5 * 1024 * 1024
	segmentIndex := 0

	for i := 0; i < len(mediaData); i += chunkSize {
		end := i + chunkSize
		if end > len(mediaData) {
			end = len(mediaData)
		}
		chunk := mediaData[i:end]

		appendBody := &bytes.Buffer{}
		appendWriter := multipart.NewWriter(appendBody)
		appendWriter.WriteField("media_id", initResult.Data.ID)
		appendWriter.WriteField("segment_index", fmt.Sprintf("%d", segmentIndex))

		part, _ := appendWriter.CreateFormField("media")
		part.Write(chunk)
		appendWriter.Close()

		req, _ := http.NewRequest("POST", "https://api.x.com/2/media/upload/append", appendBody)
		req.Header.Set("Content-Type", appendWriter.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
			bodyText, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("media upload append failed: %s - %s", resp.Status, string(bodyText))
		}

		segmentIndex++
	}

	finalizeBody := &bytes.Buffer{}
	finalizeWriter := multipart.NewWriter(finalizeBody)
	finalizeWriter.WriteField("media_id", initResult.Data.ID)
	finalizeWriter.Close()

	req, _ = http.NewRequest("POST", "https://api.x.com/2/media/upload/finalize", finalizeBody)
	req.Header.Set("Content-Type", finalizeWriter.FormDataContentType())

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	finalizeBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("media upload finalize failed: %s - %s", resp.Status, string(finalizeBytes))
	}

	var finalizeResult struct {
		Data struct {
			ProcessingInfo *struct {
				State          string `json:"state"`
				CheckAfterSecs int    `json:"check_after_secs"`
			} `json:"processing_info"`
		} `json:"data"`
	}
	if err := json.Unmarshal(finalizeBytes, &finalizeResult); err != nil {
		return "", err
	}

	if finalizeResult.Data.ProcessingInfo != nil {
		if err := waitForProcessing(client, initResult.Data.ID, finalizeResult.Data.ProcessingInfo.CheckAfterSecs); err != nil {
			return "", err
		}
	}

	return initResult.Data.ID, nil
}

func waitForProcessing(client *http.Client, mediaID string, checkAfterSecs int) error {
	if checkAfterSecs > 0 {
		time.Sleep(time.Duration(checkAfterSecs) * time.Second)
	}

	for i := 0; i < 60; i++ {
		statusBody := &bytes.Buffer{}
		statusWriter := multipart.NewWriter(statusBody)
		statusWriter.WriteField("media_id", mediaID)
		statusWriter.Close()

		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.x.com/2/media/upload/status?media_id=%s", mediaID), nil)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var statusResult struct {
			ProcessingInfo struct {
				State          string `json:"state"`
				CheckAfterSecs int    `json:"check_after_secs"`
			} `json:"processing_info"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&statusResult); err != nil {
			return err
		}

		switch statusResult.ProcessingInfo.State {
		case "succeeded":
			return nil
		case "failed":
			return fmt.Errorf("media processing failed")
		case "in_progress":
			if statusResult.ProcessingInfo.CheckAfterSecs > 0 {
				time.Sleep(time.Duration(statusResult.ProcessingInfo.CheckAfterSecs) * time.Second)
			} else {
				time.Sleep(2 * time.Second)
			}
		}
	}

	return fmt.Errorf("media processing timeout")
}

func postTweet(client *http.Client, text string, mediaID string) error {
	url := "https://api.x.com/2/tweets"

	tweet := map[string]interface{}{
		"text": text,
	}

	if mediaID != "" {
		tweet["media"] = map[string]interface{}{
			"media_ids": []string{mediaID},
		}
	}

	jsonData, _ := json.Marshal(tweet)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyText, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tweet failed: %s - %s", resp.Status, string(bodyText))
	}

	return nil
}

// ----------------------------------------------------------------
// Main
// ----------------------------------------------------------------
func main() {
	ctx := context.Background()

	fmt.Println("Fetching random image from Google Drive...")
	fileData, metadata, err := getRandomImageFromDrive(ctx)
	if err != nil {
		fmt.Printf("Error fetching image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Selected file: %s (%s)\n", metadata.Name, metadata.MimeType)

	message, altText := generateMessage(metadata.Description)
	fmt.Printf("Message: %s\n", message)
	if altText != "" {
		fmt.Printf("Alt text: %s\n", altText)
	}

	client := createOAuth1Client()

	fmt.Println("Uploading media to X...")
	mediaID, err := uploadMedia(client, fileData, metadata.MimeType, altText)
	if err != nil {
		fmt.Printf("Error uploading media: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Media uploaded successfully (ID: %s)\n", mediaID)

	fmt.Println("Posting tweet...")
	if err := postTweet(client, message, mediaID); err != nil {
		fmt.Printf("Error posting tweet: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Tweet posted successfully!")
}
