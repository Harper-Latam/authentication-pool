package authentication_pool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//AppleProvider struct
type AppleProvider struct {
	api appleAPI
}

// NewAppleProvider , return AppleProvider struct pointer
func NewAppleProvider() *AppleProvider {
	return &AppleProvider{api: &applePeople{}}
}

type appleAPI interface {
	GetUser(accessToken string) (*AppleUser, error)
}

type applePeople struct{}

func (h applePeople) GetUser(accessToken string) (user *AppleUser, err error) {
	//url := fmt.Sprintf("https://appleid.apple.com/auth/authorize?client_id=%s", accessToken)
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=%s", accessToken)
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 400 {
			return nil, fmt.Errorf("the given token is not valid")
		}

		return nil, NewProviderError(err, "invalid response from server. Please try again")
	}

	user = &AppleUser{}
	if err = json.Unmarshal(data, user); err != nil {
		return nil, err
	}

	return user, err
}

type AppleUser struct {
	ID        string `json:"sub"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
}

func (f AppleProvider) Retrieve(input *ValidationInput) (*ValidationOutput, error) {
	user, err := f.api.GetUser(input.Secret)
	if err != nil {
		return nil, err
	}

	return &ValidationOutput{
		ID:             user.ID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		PhotoURL:       &user.Picture,
		EmailValidated: true,
	}, nil
}

func (f AppleProvider) Name() string {
	return "apple"
}
