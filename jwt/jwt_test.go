package jwt_test

import (
	"testing"
	"time"

	"github.com/devil-dwj/go-wms/jwt"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	j := jwt.NewJWT("secret", time.Duration(1)*time.Hour)

	username := "dwj"
	role := "admin"

	token, err := j.Generate(username, role)
	require.NoError(t, err)
	require.NotEqual(t, "", token)

	err = j.Verify(token)
	require.NoError(t, err)

	time.Sleep(time.Duration(2) * time.Second)

	err = j.Verify(token)
	require.NotNil(t, err)
}
