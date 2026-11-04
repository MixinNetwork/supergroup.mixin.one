package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestSDKPrivateKeySupportsLegacyAndSeedFormats(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}
	key := ed25519.NewKeyFromSeed(seed)
	want := hex.EncodeToString(seed)
	inputs := map[string]string{
		"hex seed":           want,
		"uppercase hex seed": strings.ToUpper(want),
		"hex private key":    hex.EncodeToString(key),
	}
	for name, encoding := range map[string]*base64.Encoding{
		"raw URL": base64.RawURLEncoding, "padded URL": base64.URLEncoding,
		"raw standard": base64.RawStdEncoding, "padded standard": base64.StdEncoding,
	} {
		inputs[name+" seed"] = encoding.EncodeToString(seed)
		inputs[name+" private key"] = encoding.EncodeToString(key)
	}
	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			if got := SDKPrivateKey(input); got != want {
				t.Fatal("incorrect converted seed")
			}
			token, err := bot.SignOauthAccessToken("app", "auth", SDKPrivateKey(input), "GET", "/safe/me", "", "PROFILE:READ", "request")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := jwt.Parse(token, func(token *jwt.Token) (any, error) { return key.Public(), nil }); err != nil {
				t.Fatalf("signature does not match original key: %v", err)
			}
		})
	}
}

func TestNormalizePrivateKeyRejectsInvalidKeys(t *testing.T) {
	for _, input := range []string{"", "not-a-key", hex.EncodeToString(make([]byte, 31)), base64.StdEncoding.EncodeToString(make([]byte, 33))} {
		if _, err := normalizePrivateKey(input); err == nil {
			t.Error("accepted invalid key")
		}
	}
}

func TestLegacyKeyEncryptsAndDecryptsMessages(t *testing.T) {
	key := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	legacy := base64.RawURLEncoding.EncodeToString(key)
	public, err := bot.PublicKeyToCurve25519(key.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	sessionID := "0d76734c-25b1-4619-9a15-69b898c9cd6d"
	message := base64.RawURLEncoding.EncodeToString([]byte("hello"))
	encrypted, err := bot.EncryptMessageData(message, []*bot.Session{{SessionID: sessionID, PublicKey: base64.RawURLEncoding.EncodeToString(public[:])}}, SDKPrivateKey(legacy))
	if err != nil {
		t.Fatal(err)
	}
	got, err := bot.DecryptMessageData(encrypted, sessionID, SDKPrivateKey(legacy))
	if err != nil || got != message {
		t.Fatalf("decrypt = %q, %v; want %q", got, err, message)
	}
}
