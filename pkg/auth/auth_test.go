package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Implementing the Source interface
type testAuthSource struct {
	user string
	pass string
}

func (tst testAuthSource) AdminUsername() string { return tst.user }
func (tst testAuthSource) AdminPassword() string { return tst.pass }

/*
Table testing the authorize function
*/
func TestAuthorize(t *testing.T) {
	type authTestCase struct {
		desc          string
		inputUsername string
		inputPassword string
		realUsername  string
		realPassword  string
		expectError   error
	}
	cache := NewCache()
	for _, tc := range []authTestCase{
		{
			desc:          "Passing test case where auth works",
			inputUsername: "admin",
			inputPassword: "abc123",
			realUsername:  "admin",
			realPassword:  "abc123",
			expectError:   nil,
		},
		{
			desc:          "Auth fails because username is empty",
			inputUsername: "",
			inputPassword: "abc123",
			realUsername:  "admin",
			realPassword:  "abc123",
			expectError:   &InvalidCredentials{},
		},
		{
			desc:          "Auth fails because password is empty",
			inputUsername: "admin",
			inputPassword: "",
			realUsername:  "admin",
			realPassword:  "abc123",
			expectError:   &InvalidCredentials{},
		},
		{
			desc:          "Auth fails because password is wrong",
			inputUsername: "admin",
			inputPassword: "xyz987",
			realUsername:  "admin",
			realPassword:  "abc123",
			expectError:   &InvalidCredentials{},
		},
		{
			desc:          "Auth fails because username is wrong",
			inputUsername: "admin",
			inputPassword: "",
			realUsername:  "admin",
			realPassword:  "abc123",
			expectError:   &InvalidCredentials{},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {

			_, err := Authorize(&Credentials{Username: tc.inputUsername,
				Password: tc.inputPassword},
				cache,
				testAuthSource{user: tc.realUsername, pass: tc.realPassword})
			assert.Equal(t, tc.expectError, err)

		})
	}
}