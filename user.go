package myfitnesspal

import (
	"fmt"
	"time"
)

// User represents a user's information from the API
type User struct {
	UserID      int64        `json:"userId"`
	Domain      string       `json:"domain"`
	Region      string       `json:"region"`
	Status      string       `json:"status"`
	Profile     UserProfile  `json:"profile"`
	ProfileEmails struct {
		IsEmailVerified bool `json:"isEmailVerified"`
		Emails         []struct {
			Email     string `json:"email"`
			Verified  bool   `json:"verified"`
			Primary   bool   `json:"primary"`
		} `json:"emails"`
	} `json:"profileEmails"`
	AccountLinks []struct {
		UserID        int64  `json:"userId"`
		Domain        string `json:"domain"`
		DomainUserID  string `json:"domainUserId"`
		Timestamps    struct {
			Created string `json:"CREATED"`
			Updated string `json:"UPDATED"`
		} `json:"timestamps"`
		ConsentMatrixVersion string `json:"consent_matrix_version"`
		Consents            struct {
			TransferOutsideLocation bool `json:"transfer_outside_location"`
			SensitiveDataProcessing bool `json:"sensitive_data_processing"`
		} `json:"consents"`
		AdConsentsLastSeen *time.Time `json:"ad_consents_last_seen,omitempty"`
		AdvertisingIDs     struct{}    `json:"advertising_ids"`
		IsAutoGenerated    *bool `json:"isAutoGenerated"`
	} `json:"accountLinks"`
}

// UserProfile represents a user's profile information
type UserProfile struct {
	FullName         string  `json:"fullName"`
	DisplayName      string  `json:"displayName"`
	FirstName        string  `json:"firstName"`
	LastName         *string `json:"lastName"`
	ProfilePictureURI string `json:"profilePictureUri"`
	Birthdate        string  `json:"birthdate"`
	Gender           string  `json:"gender"`
	Weight           float64 `json:"weight"`
	Height           float64 `json:"height"` // Height in inches
	Locale           string  `json:"locale"`
	Location         struct {
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	} `json:"location"`
}

// HeightInCM returns the height in centimeters
func (p *UserProfile) HeightInCM() float64 {
	return p.Height * 2.54 // Convert inches to centimeters
}

// GetUser fetches the user's information from the API
func (c *Client) GetUser(session *UserSession) (*User, error) {
	var user User

	// Create a new request
	req := c.identityClient.R().
		SetResult(&user)

	// Set standard headers first
	c.setStandardHeaders(req, session)

	resp, err := req.Get("/users/" + session.UserID + "?fetch_profile=true&fetch_emails=true")
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}

	return &user, nil
}