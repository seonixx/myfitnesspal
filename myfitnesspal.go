package myfitnesspal

import (
	"fmt"
	"time"

	"encoding/base64"

	"github.com/go-resty/resty/v2"
)

const (
	identityBaseURL = "https://identity-api.myfitnesspal.com"
	apiBaseURL     = "https://api.myfitnesspal.com"
	userAgent      = "MyFitnessPal/25.19.0 (mfp-mobile-android-google) (Android 11; Pixel 5 / Android Android SDK built for arm64) (preload=false;locale=en_US)"
	apiVersion     = "2.0.50"
)

// Client represents a MyFitnessPal API client
type Client struct {
	identityClient *resty.Client
	apiClient      *resty.Client
	clientID       string
	clientSecret   string
	deviceID       string
	clientToken    *TokenResponse
	signingKey     []byte
	keyID          string
}

// setStandardHeaders sets the standard headers for API requests
func (c *Client) setStandardHeaders(req *resty.Request, session *UserSession) {
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("user-agent", userAgent)
	req.SetHeader("device_id", c.deviceID)
	req.SetHeader("mfp-device-id", c.deviceID)
	req.SetHeader("mfp-client-id", "mfp-mobile-android-google")
	req.SetHeader("api-version", apiVersion)
	req.SetHeader("accept-language", "en-US")
	req.SetHeader("accept-encoding", "gzip")

	if session != nil {
		req.SetHeader("Authorization", "Bearer "+session.AccessToken)
		req.SetHeader("mfp-user-id", session.DomainUserID)
	}
}

// NewClient creates a new MyFitnessPal API client
func NewClient(clientID, clientSecret string) (*Client, error) {
	identityClient := resty.New()
	identityClient.SetBaseURL(identityBaseURL)

	apiClient := resty.New()
	apiClient.SetBaseURL(apiBaseURL)

	// Generate a random device ID
	deviceID := fmt.Sprintf("%x-%x-%x-%x-%x",
		time.Now().UnixNano(),
		time.Now().UnixNano()>>32,
		time.Now().UnixNano()>>16,
		time.Now().UnixNano()>>8,
		time.Now().UnixNano())

	client := &Client{
		identityClient: identityClient,
		apiClient:      apiClient,
		clientID:       clientID,
		clientSecret:   clientSecret,
		deviceID:       deviceID,
	}

	// Get client credentials token
	clientToken, err := client.GetClientCredentialsToken()
	if err != nil {
		return nil, fmt.Errorf("error getting client credentials token: %w", err)
	}
	client.clientToken = clientToken

	// Get client keys to find the signing key
	keys, err := client.GetClientKeys()
	if err != nil {
		return nil, fmt.Errorf("error getting client keys: %w", err)
	}

	// Find the signing key
	for _, key := range keys {
		if key.Key.Use == "sig" && key.Key.Alg == "HS512" {
			// Decode the base64 key
			signingKey, err := base64.RawURLEncoding.DecodeString(key.Key.K)
			if err != nil {
				return nil, fmt.Errorf("error decoding signing key: %w", err)
			}
			client.signingKey = signingKey
			client.keyID = key.Key.Kid
			break
		}
	}

	if client.signingKey == nil {
		return nil, fmt.Errorf("no signing key found")
	}

	return client, nil
}