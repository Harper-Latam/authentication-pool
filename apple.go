package authentication_pool

import (
	"fmt"
	"strings"

	"github.com/tideland/gorest/jwt"
)

//AppleProvider struct
type AppleProvider struct {
	api appleAPI
}

//AppleUser struct
type AppleUser struct {
	ID        string `json:"sub"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
}

// NewAppleProvider , return AppleProvider struct pointer
func NewAppleProvider() *AppleProvider {
	return &AppleProvider{api: &applePeople{}}
}

type appleAPI interface {
	GetUser(accessToken string) (*AppleUser, error)
}

type applePeople struct{}

// GetClaims decodes the data response and returns the data to identify the user
func GetClaims(idToken string) (string, string, error) {
	j, err := jwt.Decode(idToken)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%v", j.Claims()["sub"]), fmt.Sprintf("%v", j.Claims()["email"]), nil
}

// GetUser get info from User (sub, email)
func (h applePeople) GetUser(accessToken string) (user *AppleUser, err error) {
	ID, email, err := GetClaims(accessToken)
	if err != nil {
		return nil, err
	}
	user = &AppleUser{ID: ID, Email: email}
	return user, nil
}

// Retrieve get info from User (sub, email)
func (f AppleProvider) Retrieve(input *ValidationInput) (*ValidationOutput, error) {
	splitInput := strings.Split(input.Email, "*")
	user, err := f.api.GetUser(input.Secret)
	if err != nil {
		return nil, err
	}
	fmt.Println(user.ID)
	fmt.Println(splitInput[1])
	fmt.Println(splitInput[2])
	fmt.Println(splitInput[0])
	return &ValidationOutput{
		ID:             user.ID,
		FirstName:      splitInput[1],
		LastName:       splitInput[2],
		Email:          splitInput[0],
		PhotoURL:       nil,
		EmailValidated: true,
	}, nil
}

// Name provider auth
func (f AppleProvider) Name() string {
	return "apple"
}
