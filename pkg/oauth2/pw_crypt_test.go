// Developer: Saif Hamdan
// Date: 18/7/2023

package oauth2

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestCrypto(t *testing.T) {

	t.Run("OK", func(t *testing.T) {
		password, err := EncryptPassword("saif")
		require.NoError(t, err)
		require.NotEmpty(t, password)

		require.True(t, ComparePassword(password, "saif"))
	})
	t.Run("tooLongPassword", func(t *testing.T) {
		password, err := EncryptPassword("W7Px4TcHrQnljd1212KGJ9skL1gdIBvuXZftqY2aU0mo3SiDhfewfefd12C6z5wEpRbFOVeM8NyA")
		require.EqualError(t, err, bcrypt.ErrPasswordTooLong.Error())
		require.Empty(t, password)
	})
}
