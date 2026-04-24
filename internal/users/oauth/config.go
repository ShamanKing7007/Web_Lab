package oauth

import "os"

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	Provider     string // "yandex" или "vk"
}

func LoadOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		CallbackURL:  os.Getenv("CALLBACK_URL"),
		Provider:     os.Getenv("OAUTH_PROVIDER"),
	}
}
