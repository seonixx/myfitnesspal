package myfitnesspal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ClientKey represents a client key from the API
type ClientKey struct {
	Key struct {
		Kty string `json:"kty"`
		Use string `json:"use"`
		Kid string `json:"kid"`
		K   string `json:"k"`
		Alg string `json:"alg"`
	} `json:"key"`
	ClientID   string `json:"clientId"`
	KeyID      string `json:"keyId"`
	Timestamps struct {
		Created string `json:"CREATED"`
		Updated string `json:"UPDATED"`
	} `json:"timestamps"`
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Data         string `json:"data,omitempty"`
}

// UserSession represents a user's authentication session
type UserSession struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	DomainUserID string    `json:"domain_user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	Data         string    `json:"data"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GetClientKeys retrieves the client keys from the API
func (c *Client) GetClientKeys() ([]ClientKey, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))

	var result struct {
		Embedded struct {
			ClientKeys []ClientKey `json:"clientKeys"`
		} `json:"_embedded"`
	}

	req := c.identityClient.R().
		SetResult(&result)

	// Set standard headers first
	c.setStandardHeaders(req, nil)

	// Override specific headers
	req.SetHeader("Authorization", "Basic "+auth)

	resp, err := req.Get("/clientKeys")
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return result.Embedded.ClientKeys, nil
}

// GetClientCredentialsToken gets an OAuth token using client credentials
func (c *Client) GetClientCredentialsToken() (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "client_credentials")

	var token TokenResponse

	req := c.identityClient.R().
		SetBody(data.Encode()).
		SetResult(&token)

	// Set standard headers first
	c.setStandardHeaders(req, nil)

	// Override Content-Type for form data
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	resp, err := req.Post("/oauth/token")
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return &token, nil
}

// createSessionFromTokenResponse creates a UserSession from a TokenResponse and fetches additional user info
func (c *Client) createSessionFromTokenResponse(mfpUserID string, tokenResp *TokenResponse) (*UserSession, error) {
	// Parse the ID token to get the user ID
	parts := strings.Split(tokenResp.IDToken, ".")
	if mfpUserID == "" && len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	var userID string
	if mfpUserID == "" {
		// Decode the payload
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("error decoding token payload: %w", err)
		}
		var claims struct {
			Sub string `json:"sub"`
		}
		if err := json.Unmarshal(payload, &claims); err != nil {
			return nil, fmt.Errorf("error parsing token claims: %w", err)
		}
		userID = claims.Sub
	} else {
		userID = mfpUserID
	}

	// Create initial session with identity user ID
	session := &UserSession{
		UserID:       userID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		Data:         tokenResp.Data,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	// Get user info to extract the DomainUserID
	user, err := c.GetUser(session)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	// Find the DomainUserID from account links
	for _, link := range user.AccountLinks {
		if link.Domain == "MFP" {
			session.DomainUserID = link.DomainUserID
			break
		}
	}
	session.Email = user.ProfileEmails.Emails[0].Email
	session.FirstName = user.Profile.FirstName
	if user.Profile.LastName != nil {
		session.LastName = *user.Profile.LastName
	}

	if session.DomainUserID == "" {
		return nil, fmt.Errorf("no domain user ID found in account links")
	}

	return session, nil
}

// Login authenticates with username and password and returns the user's tokens
func (c *Client) Login(username, password string) (*UserSession, error) {
	// Create the JWT claims
	claims := jwt.MapClaims{
		"password": password,
		"username": username,
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token.Header["kid"] = c.keyID

	// Sign the token
	credentials, err := token.SignedString(c.signingKey)
	if err != nil {
		return nil, fmt.Errorf("error signing token: %w", err)
	}

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("credentials", credentials)
	data.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	data.Set("redirect_uri", "mfp://identity/callback")
	data.Set("response_type", "code")
	data.Set("scope", "openid")

	// Create a client that doesn't follow redirects
	noRedirectClient := resty.New()
	noRedirectClient.SetBaseURL(identityBaseURL)
	noRedirectClient.SetRedirectPolicy(resty.NoRedirectPolicy())

	req := noRedirectClient.R().
		SetBody(data.Encode())

	// Set standard headers first
	c.setStandardHeaders(req, nil)

	// Override specific headers
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	req.SetHeader("Authorization", "Bearer "+c.clientToken.AccessToken)

	resp, err := req.Post("/oauth/authorize")
	if err != nil {
		// Check if this is a redirect error
		if strings.Contains(err.Error(), "auto redirect is disabled") {
			// Get the Location header from the response
			location := resp.Header().Get("Location")
			if location == "" {
				return nil, fmt.Errorf("no location header in response")
			}

			// Parse the code from the redirect URL
			redirectURL, err := url.Parse(location)
			if err != nil {
				return nil, fmt.Errorf("error parsing redirect URL: %w", err)
			}

			code := redirectURL.Query().Get("code")
			if code == "" {
				return nil, fmt.Errorf("no code in redirect URL")
			}

			// Exchange the code for a token
			tokenResp, err := c.ExchangeCodeForToken(code)
			if err != nil {
				return nil, fmt.Errorf("error exchanging code for token: %w", err)
			}

			return c.createSessionFromTokenResponse("", tokenResp)
		}
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode())
}

// RefreshUserToken refreshes a user's access token using their refresh token
func (c *Client) RefreshUserToken(mfpUserID string, refreshToken string) (*UserSession, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token provided")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	var tokenResp TokenResponse

	req := c.identityClient.R().
		SetBody(data.Encode()).
		SetResult(&tokenResp)

	// Set standard headers first
	c.setStandardHeaders(req, nil)

	// Override Content-Type for form data
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	resp, err := req.Post("/oauth/token")
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return c.createSessionFromTokenResponse(mfpUserID, &tokenResp)
}

// IsTokenExpired checks if a token is expired or about to expire
func (c *Client) IsTokenExpired(expiresIn int, createdAt time.Time) bool {
	expiry := createdAt.Add(time.Duration(expiresIn) * time.Second)
	// Consider token expired if it's within 30 seconds of expiry
	return time.Until(expiry) < 30*time.Second
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (c *Client) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("redirect_uri", "mfp://identity/callback")

	var tokenResp TokenResponse

	req := c.identityClient.R().
		SetBody(data.Encode()).
		SetResult(&tokenResp)

	// Set standard headers first
	c.setStandardHeaders(req, nil)

	// Override Content-Type for form data
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	resp, err := req.Post("/oauth/token")
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return &tokenResp, nil
}
