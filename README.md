# MyFitnessPal API client for Go

MyFitnessPal API client for Go that actually works.

Leverages private APIs and OAuth authentication to give persistent, reliable connectivity. No need for browser emulation or cookie jar shenanigans.

## Installation

```bash
go get github.com/seonixx/myfitnesspal@latest
```

## Quick Start

1. Create a .env file:

```bash
cp .env.sample .env
```

2. Create a new file `main.go` with the following code:

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/seonixx/myfitnesspal"
)

func main() {
    // Create a new client (credentials in .env.sample)
    client := myfitnesspal.NewClient(
        os.Getenv("MFP_CLIENT_ID"),
        os.Getenv("MFP_CLIENT_SECRET"),
    )

    // Authenticate a user
    session, err := client.AuthenticateUser(context.Background(), "username", "password")
    if err != nil {
        log.Fatal(err)
    }

    // Search for foods
    results, err := client.SearchFood(context.Background(), session, myfitnesspal.SearchFoodRequest{
        Query: "chicken breast",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Work with the search results
    for _, food := range results {
        log.Printf("Found food: %s (ID: %s)", food.Name, food.ID)
    }
}
```

3. Run the example:

```bash
go run main.go
```

## Features

- OAuth authentication
- Search MFP food database
- Create foods
- Add foods to diary
- More coming soon...

Need an endpoint I haven't done yet? Create an issue and I'll add it.

## API Documentation

### Authentication

```go
// Create a new client
client := myfitnesspal.NewClient(clientID, clientSecret)

// Authenticate a user
session, err := client.Login(ctx, username, password)
```

### Food

```go
// Add a food to the database
foodResp, err := client.CreateFood(session, food)
```

### Diary

```go
// Add a food to your diary
addResp, err := client.AddFoodToDiary(session, req)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
