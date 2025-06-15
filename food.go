package myfitnesspal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type MealNumber int

const (
	Breakfast MealNumber = iota
	Lunch
	Dinner
	Snacks
)

// FoodSearchResult represents a food item from the search results
type FoodSearchResult struct {
	HealthLabels []string `json:"health_labels"`
	Item         struct {
		BrandName           string `json:"brand_name"`
		BrandedWithBarcode  bool   `json:"branded_with_barcode"`
		CountryCode         string `json:"country_code"`
		Deleted             bool   `json:"deleted"`
		Description         string `json:"description"`
		ID                  string `json:"id"`
		NutritionalContents struct {
			AdditionalColumns map[string]interface{} `json:"additional_columns"`
			Calcium           float64                `json:"calcium,omitempty"`
			Carbohydrates     float64                `json:"carbohydrates"`
			Cholesterol       float64                `json:"cholesterol,omitempty"`
			Energy            struct {
				Unit  string  `json:"unit"`
				Value float64 `json:"value"`
			} `json:"energy"`
			Fat                float64 `json:"fat"`
			Fiber              float64 `json:"fiber,omitempty"`
			Grams              float64 `json:"grams"`
			Iron               float64 `json:"iron,omitempty"`
			MonounsaturatedFat float64 `json:"monounsaturated_fat,omitempty"`
			NetCarbs           float64 `json:"net_carbs"`
			PolyunsaturatedFat float64 `json:"polyunsaturated_fat,omitempty"`
			Potassium          float64 `json:"potassium,omitempty"`
			Protein            float64 `json:"protein"`
			SaturatedFat       float64 `json:"saturated_fat,omitempty"`
			Sodium             float64 `json:"sodium,omitempty"`
			Sugar              float64 `json:"sugar,omitempty"`
			TransFat           float64 `json:"trans_fat,omitempty"`
			VitaminA           float64 `json:"vitamin_a,omitempty"`
			VitaminC           float64 `json:"vitamin_c,omitempty"`
		} `json:"nutritional_contents"`
		Public       bool `json:"public"`
		ServingSizes []struct {
			ID                  string  `json:"id"`
			Index               int     `json:"index"`
			NutritionMultiplier float64 `json:"nutrition_multiplier"`
			Unit                string  `json:"unit"`
			Value               float64 `json:"value"`
		} `json:"serving_sizes"`
		Type     string `json:"type"`
		UserID   string `json:"user_id"`
		Verified bool   `json:"verified"`
		Version  string `json:"version"`
	} `json:"item"`
	Tags []string `json:"tags"`
	Type string   `json:"type"`
}

// FoodItem represents a food item to be created
type FoodItem struct {
	UserID              string              `json:"user_id"`
	BrandName           string              `json:"brand_name"`
	Description         string              `json:"description"`
	NutritionalContents NutritionalContents `json:"nutritional_contents"`
	ServingSizes        []ServingSize       `json:"serving_sizes"`
	Public              bool                `json:"public"`
	CountryCode         string              `json:"country_code"`
}

// NutritionalContents represents the nutritional information for a food item
type NutritionalContents struct {
	Calcium            float64 `json:"calcium,omitempty"`
	Carbohydrates      float64 `json:"carbohydrates,omitempty"`
	Cholesterol        float64 `json:"cholesterol,omitempty"`
	Energy             Energy  `json:"energy,omitempty"`
	Fat                float64 `json:"fat,omitempty"`
	Fiber              float64 `json:"fiber,omitempty"`
	Grams              float64 `json:"grams,omitempty"`
	Iron               float64 `json:"iron,omitempty"`
	MonounsaturatedFat float64 `json:"monounsaturated_fat,omitempty"`
	NetCarbs           float64 `json:"net_carbs,omitempty"`
	PolyunsaturatedFat float64 `json:"polyunsaturated_fat,omitempty"`
	Potassium          float64 `json:"potassium,omitempty"`
	Protein            float64 `json:"protein,omitempty"`
	SaturatedFat       float64 `json:"saturated_fat,omitempty"`
	Sodium             float64 `json:"sodium,omitempty"`
	Sugar              float64 `json:"sugar,omitempty"`
	TransFat           float64 `json:"trans_fat,omitempty"`
	VitaminA           float64 `json:"vitamin_a,omitempty"`
	VitaminC           float64 `json:"vitamin_c,omitempty"`
}

// Energy represents the energy content of a food item
type Energy struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

// ServingSize represents a serving size for a food item
type ServingSize struct {
	Value               float64 `json:"value"`
	Unit                string  `json:"unit"`
	NutritionMultiplier float64 `json:"nutrition_multiplier"`
}

// CreateFoodResponse represents the response from creating a food item
type CreateFoodResponse struct {
	Items []struct {
		ID                  string              `json:"id"`
		Description         string              `json:"description"`
		BrandName           string              `json:"brand_name"`
		NutritionalContents NutritionalContents `json:"nutritional_contents"`
		ServingSizes        []ServingSize       `json:"serving_sizes"`
		Public              bool                `json:"public"`
		UserID              string              `json:"user_id"`
		CountryCode         string              `json:"country_code"`
		Version             string              `json:"version"`
		Type                string              `json:"type"`
		Verified            bool                `json:"verified"`
		Deleted             bool                `json:"deleted"`
	} `json:"items"`
}

// FoodDiaryAddRequest represents the request to add a food entry to the diary
type FoodDiaryAddRequest struct {
	Type         string     `json:"type"` // always "food_entry"
	Date         string     `json:"date"`
	MealPosition MealNumber `json:"meal_position"` // 0: Breakfast, 1: Lunch, 2: Dinner, 3: Snacks
	Food         struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"food"`
	Servings    float64     `json:"servings"`
	ServingSize ServingSize `json:"serving_size"`
}

// FoodDiaryAddResponse represents the response from adding a food entry
type FoodDiaryAddResponse struct {
	Items []struct {
		ID                  string              `json:"id"`
		Type                string              `json:"type"`
		ClientID            string              `json:"client_id"`
		Date                string              `json:"date"`
		MealName            string              `json:"meal_name"`
		MealPosition        int                 `json:"meal_position"`
		Food                FoodItem            `json:"food"`
		ServingSize         ServingSize         `json:"serving_size"`
		Servings            float64             `json:"servings"`
		MealFoodID          string              `json:"meal_food_id"`
		NutritionalContents NutritionalContents `json:"nutritional_contents"`
		Geolocation         struct{}            `json:"geolocation"`
		ImageIDs            []string            `json:"image_ids"`
		Tags                []string            `json:"tags"`
		ConsumedAt          *string             `json:"consumed_at"`
		LoggedAt            *string             `json:"logged_at"`
		LoggedAtOffset      *string             `json:"logged_at_offset"`
	} `json:"items"`
}

// CreateFood creates a new food item in the MyFitnessPal database
func (c *Client) CreateFood(session *UserSession, food FoodItem) (*CreateFoodResponse, error) {
	var response CreateFoodResponse

	// Create a new request
	req := c.apiClient.R().
		SetBody(map[string]interface{}{
			"item": food,
		}).
		SetResult(&response)

	// Set standard headers first
	c.setStandardHeaders(req, session)

	resp, err := req.Post("/v2/foods")
	if err != nil {
		return nil, fmt.Errorf("failed to create food: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("create food request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return &response, nil
}

func (c *Client) AddFoodToDiary(session *UserSession, params FoodDiaryAddRequest) (*FoodDiaryAddResponse, error) {
	var respData FoodDiaryAddResponse
	// Wrap the request in an items array
	body := map[string]interface{}{
		"items": []FoodDiaryAddRequest{params},
	}

	// Create a new request
	req := c.apiClient.R().
		SetBody(body)

	// Set standard headers first
	c.setStandardHeaders(req, session)

	resp, err := req.Post("/v2/diary")

	if err != nil {
		return nil, fmt.Errorf("failed to add food to diary: %w", err)
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("add food to diary failed with status %d: %s", resp.StatusCode(), resp.String())
	}
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &respData, nil
}

// SearchFood searches for food items in the MyFitnessPal database
type SearchFoodRequest struct {
	Query       string
	Scope       *string
	MaxItems    *int
	CountryCode *string
}

func (c *Client) SearchFood(session *UserSession, params SearchFoodRequest) ([]FoodSearchResult, error) {
	if params.MaxItems == nil {
		params.MaxItems = new(int)
		*params.MaxItems = 25
	}

	if params.Scope != nil {
		*params.Scope = "all"
	}

	// Validate scope parameter
	if *params.Scope != "all" && *params.Scope != "user" {
		return nil, fmt.Errorf("invalid scope: %s. Must be 'all' or 'user'", *params.Scope)
	}

	// Build the search URL with query parameters
	url := "/v2/search/nutrition"
	queryParams := map[string]string{
		"q":               params.Query,
		"scope":           *params.Scope,
		"max_items":       strconv.Itoa(*params.MaxItems),
		"resource_type[]": "foods",
	}

	if params.CountryCode != nil {
		queryParams["country_code"] = *params.CountryCode
	}

	req := c.apiClient.R().
		SetQueryParams(queryParams).
		SetQueryParam("fields[]", "id").
		SetQueryParam("fields[]", "nutritional_contents").
		SetQueryParam("fields[]", "serving_sizes").
		SetQueryParam("fields[]", "version").
		SetQueryParam("fields[]", "brand_name").
		SetQueryParam("fields[]", "description")

	// Set standard headers first
	c.setStandardHeaders(req, session)

	// Add flow ID header for search
	req.SetHeader("mfp-flow-id", fmt.Sprintf("%x-%x-%x-%x-%x",
		time.Now().UnixNano(),
		time.Now().UnixNano()>>32,
		time.Now().UnixNano()>>16,
		time.Now().UnixNano()>>8,
		time.Now().UnixNano()))

	resp, err := req.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make search request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	var result struct {
		Items []FoodSearchResult `json:"items"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return result.Items, nil
}
