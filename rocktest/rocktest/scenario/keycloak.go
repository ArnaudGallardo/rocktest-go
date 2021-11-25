package scenario

import (
	"context"
	"errors"

	"github.com/Nerzal/gocloak/v10"
)

func (module *Module) Keycloak(params map[string]interface{}, scenario *Scenario) error {

	paramsEx, err := scenario.ExpandMap(params)
	if err != nil {
		return err
	}

	val, err := scenario.GetString(paramsEx, "value", nil)
	if err != nil {
		return err
	}

	// We have something like
	// - keycloak: VARNAME
	//   params:
	//     url: Keycloak base url
	//     realm: Keycloak realm
	//     clientId: Keycloak client id
	//     clientSecret: Keycloak client secret
	//     username: Keycloak username
	//     password: Keycloak password
	// Which is handy if the value is on multiple lines
	// In this case, we need value AND name fields

	// TODO: check value, should not contain spaces or -

	url, err := scenario.GetString(paramsEx, "url", nil)
	if err != nil {
		return errors.New("Keycloak base url (without /auth/) is required (url)")
	}

	realm, err := scenario.GetString(paramsEx, "realm", nil)
	if err != nil {
		return errors.New("Keycloak realm is required (realm)")
	}

	clientId, err := scenario.GetString(paramsEx, "clientId", nil)
	if err != nil {
		return errors.New("Keycloak client id is required (clientId)")
	}

	clientSecret, err := scenario.GetString(paramsEx, "clientSecret", nil)
	if err != nil {
		return errors.New("Keycloak client secret is required (clientSecret)")
	}

	username, err := scenario.GetString(paramsEx, "username", nil)
	if err != nil {
		return errors.New("Username is required (username)")
	}

	password, err := scenario.GetString(paramsEx, "password", nil)
	if err != nil {
		return errors.New("Password is required (password)")
	}

	client := gocloak.NewClient(url)
	ctx := context.Background()
	token, err := client.Login(ctx, clientId, clientSecret, realm, username, password)
	if err != nil {
		return err
	}

	scenario.PutContext(val, token.AccessToken)

	return nil
}
