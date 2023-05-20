package yandex

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	keyID            = "key_id"
	serviceAccountID = "service_account_id"
	keyFile          = "key_file"
)

type IAM struct {
	// mu    sync.Mutex
	token string
}

func NewIAM() (*IAM, error) {
	iam, err := getIAMToken()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create iam token")
	}

	i := &IAM{token: iam}
	// go i.pollIAMToken()

	return i, nil
}

func (i *IAM) Value() string {
	// i.mu.Lock()
	// defer i.mu.Unlock()

	return i.token
}

// func (i *IAM) pollIAMToken() {
// 	t := time.NewTicker(1 * time.Hour)
// 	for range t.C {
// 		iam, err := getIAMToken()
// 		if err != nil {
// 			// do something
// 			continue
// 		}
// 		i.mu.Lock()
// 		i.token = iam
// 		i.mu.Unlock()
// 	}
// }

func loadPrivateKey() (*rsa.PrivateKey, error) {
	file, err := os.Open(viper.GetString(keyFile))
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return rsaPrivateKey, nil
}

func signedToken() (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    viper.GetString(serviceAccountID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Audience:  []string{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	token.Header["kid"] = viper.GetString(keyID)

	privateKey, err := loadPrivateKey()
	if err != nil {
		return "", errors.WithStack(err)
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.New(err.Error())
	}
	return signed, nil
}

func getIAMToken() (string, error) {
	jwt, err := signedToken()
	if err != nil {
		return "", errors.Wrap(err, "unable to sign token: %w")
	}
	resp, err := http.Post(
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, jwt)),
	)
	if err != nil {
		return "", errors.New(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf("%s: %s", resp.Status, body)
	}
	var data struct {
		IAMToken string `json:"iamToken"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", errors.New("decoding error")
	}
	return data.IAMToken, nil
}
