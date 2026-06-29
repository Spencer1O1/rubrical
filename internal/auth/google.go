package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type GoogleProfile struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func (g GoogleConfig) Enabled() bool {
	return strings.TrimSpace(g.ClientID) != "" && strings.TrimSpace(g.ClientSecret) != ""
}

func (g GoogleConfig) OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (g GoogleConfig) AuthCodeURL(state string) string {
	return g.OAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (g GoogleConfig) ExchangeProfile(ctx context.Context, code string) (GoogleProfile, error) {
	token, err := g.OAuthConfig().Exchange(ctx, code)
	if err != nil {
		return GoogleProfile{}, err
	}
	client := g.OAuthConfig().Client(ctx, token)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return GoogleProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return GoogleProfile{}, fmt.Errorf("google userinfo: %s", strings.TrimSpace(string(body)))
	}
	var profile GoogleProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return GoogleProfile{}, err
	}
	if profile.Sub == "" {
		return GoogleProfile{}, fmt.Errorf("google profile missing subject")
	}
	return profile, nil
}
