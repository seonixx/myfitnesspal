package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/seonixx/myfitnesspal"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create a new client
	client, err := myfitnesspal.NewClient(
		os.Getenv("MFP_CLIENT_ID"),
		os.Getenv("MFP_CLIENT_SECRET"),
	)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// Login
	session, err := client.Login(
		os.Getenv("MFP_EMAIL"),
		os.Getenv("MFP_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("Error logging in: %v", err)
	}
	fmt.Printf("Successfully logged in. Session expires at %v\n", session.ExpiresAt)

	// Get user info
	user, err := client.GetUser(session)
	if err != nil {
		log.Fatalf("Error getting user info: %v", err)
	}

	fmt.Printf("User ID: %d\n", user.UserID)
	fmt.Printf("Domain: %s\n", user.Domain)
	fmt.Printf("Region: %s\n", user.Region)
	fmt.Printf("Status: %s\n", user.Status)
	fmt.Printf("Profile:\n")
	fmt.Printf("  Full Name: %s\n", user.Profile.FullName)
	fmt.Printf("  Display Name: %s\n", user.Profile.DisplayName)
	fmt.Printf("  Gender: %s\n", user.Profile.Gender)
	fmt.Printf("  Birthdate: %s\n", user.Profile.Birthdate)
	fmt.Printf("  Weight: %.1f kg\n", user.Profile.Weight*0.453592) // Convert from pounds to kg
	fmt.Printf("  Height: %.1f cm\n", user.Profile.HeightInCM())
	fmt.Printf("  Location: %s, %s\n", user.Profile.Location.PostalCode, user.Profile.Location.Country)
	fmt.Printf("Emails:\n")
	for _, email := range user.ProfileEmails.Emails {
		fmt.Printf("  %s (Primary: %v, Verified: %v)\n", email.Email, email.Primary, email.Verified)
	}

	// Create a new food item from the nutrition label
	food := myfitnesspal.FoodItem{
		UserID:      session.UserID,
		BrandName:   "For Goodness Shakes",
		Description: "Protein Chocolate",
		NutritionalContents: myfitnesspal.NutritionalContents{
			Sugar:              15.0,
			Fiber:              1.3,
			SaturatedFat:       0.3,
			MonounsaturatedFat: 0.1,
			Protein:            20.0,
			Carbohydrates:      16.0,
			Sodium:             0.4, // Salt in g (approximate, as sodium = salt * 0.4, but API expects sodium in g)
			Energy: myfitnesspal.Energy{
				Unit:  "calories",
				Value: 153.0,
			},
			Fat: 0.7,
		},
		ServingSizes: []myfitnesspal.ServingSize{
			{
				Value:               1.0,
				Unit:                "bottle (330ml)",
				NutritionMultiplier: 1.0,
			},
			{
				Value:               100.0,
				Unit:                "ml",
				NutritionMultiplier: 0.303, // 100ml out of 330ml
			},
		},
		Public:      false,
		CountryCode: "GB",
	}

	foodResp, err := client.CreateFood(session, food)
	if err != nil {
		log.Fatalf("Error creating food: %v", err)
	}
	fmt.Printf("\nCreated food item! ID: %s, Description: %s, Brand: %s\n",
		foodResp.Items[0].ID, foodResp.Items[0].Description, foodResp.Items[0].BrandName)

	// Add the created food to the diary
	addReq := myfitnesspal.FoodDiaryAddRequest{
		Type:         "food_entry",
		Date:         "2025-05-26",
		MealPosition: 2,
		Food: struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		}{
			ID:      foodResp.Items[0].ID,
			Version: foodResp.Items[0].ID,
		},
		Servings: 1,
		ServingSize: myfitnesspal.ServingSize{
			Value:               1,
			Unit:                "bottle (330ml)",
			NutritionMultiplier: 1.0,
		},
	}

	addResp, err := client.AddFoodToDiary(session, addReq)
	if err != nil {
		log.Fatalf("Error adding food to diary: %v", err)
	}
	fmt.Printf("Added food to diary! Entry ID: %s\n", addResp.Items[0].ID)

	// Search for my foods
	scope := "user"
	countryCode := "US"
	myFoods, err := client.SearchFood(session, myfitnesspal.SearchFoodRequest{
		Query:       "protein",
		Scope:       &scope,
		CountryCode: &countryCode,
	})
	if err != nil {
		log.Fatalf("Error searching my foods: %v", err)
	}
	fmt.Printf("\nFound %d of my foods:\n", len(myFoods))
	for i, item := range myFoods {
		fmt.Printf("\n%d. %s - %s\n", i+1, item.Item.BrandName, item.Item.Description)
		fmt.Printf("   ID: %s\n", item.Item.ID)
		fmt.Printf("   Calories: %.1f %s\n", item.Item.NutritionalContents.Energy.Value, item.Item.NutritionalContents.Energy.Unit)
		fmt.Printf("   Protein: %.1fg\n", item.Item.NutritionalContents.Protein)
		fmt.Printf("   Carbs: %.1fg\n", item.Item.NutritionalContents.Carbohydrates)
		fmt.Printf("   Fat: %.1fg\n", item.Item.NutritionalContents.Fat)
		fmt.Printf("   Serving Sizes:\n")
		for _, serving := range item.Item.ServingSizes {
			fmt.Printf("     - %.1f %s\n", serving.Value, serving.Unit)
		}
	}
}
