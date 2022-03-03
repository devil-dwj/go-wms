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

	token, err := j.Generate(username)
	require.NoError(t, err)
	require.NotEqual(t, "", token)

	_, err = j.Verify(token)
	require.NoError(t, err)

	time.Sleep(time.Duration(2) * time.Second)

	_, err = j.Verify(token)
	require.NotNil(t, err)
}
