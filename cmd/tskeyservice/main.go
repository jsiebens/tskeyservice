package main

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hashicorp/go-bexpr"
	"github.com/labstack/echo/v4"
	"github.com/tailscale/tailscale-client-go/tailscale"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const BEARER_SCHEMA = "Bearer "

type KeyResponse struct {
	Key string `json:"key"`
}

func start() error {
	ctx := context.Background()

	apiKey := os.Getenv("TS_API_KEY")
	tailnet := os.Getenv("TS_TAILNET")
	issuer := os.Getenv("TS_KEYS_ISSUER")
	tags := os.Getenv("TS_KEYS_TAGS")
	filter := os.Getenv("TS_KEYS_BEXPR")

	client, err := tailscale.NewClient(apiKey, tailnet)
	if err != nil {
		return err
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return err
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	evaluator, err := bexpr.CreateEvaluator(filter)
	if err != nil {
		return err
	}

	expiry := 5 * time.Minute
	capabilities := tailscale.KeyCapabilities{}
	capabilities.Devices.Create.Reusable = false
	capabilities.Devices.Create.Ephemeral = true
	if len(tags) != 0 {
		capabilities.Devices.Create.Tags = strings.Split(tags, ",")
	}

	e := echo.New()
	e.HideBanner = true
	e.GET("/key", func(c echo.Context) error {
		ctx := c.Request().Context()

		authHeader := c.Request().Header.Get("Authorization")

		if len(authHeader) == 0 || !strings.HasPrefix(authHeader, BEARER_SCHEMA) {
			return echo.ErrUnauthorized
		}

		idToken, err := verifier.Verify(ctx, authHeader[len(BEARER_SCHEMA):])
		if err != nil {
			return echo.ErrBadRequest
		}

		var claims = make(map[string]interface{})
		if err := idToken.Claims(&claims); err != nil {
			return echo.ErrBadRequest
		}

		if ok, _ := evaluator.Evaluate(claims); ok {
			key, err := client.CreateKey(ctx, capabilities, tailscale.WithKeyExpiry(expiry))
			if err != nil {
				return echo.ErrInternalServerError
			}

			return c.JSON(http.StatusOK, &KeyResponse{Key: key.Key})
		}

		return echo.ErrForbidden
	})

	return e.Start(":8080")
}

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}
