package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/devil-dwj/go-wms/config"
	"github.com/devil-dwj/go-wms/hash"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigJson(t *testing.T) {
	jsonStr := `{
		"a": "a",
		"b": 1
	}`

	tmpFile, err := createTempFile(jsonStr)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	var val struct {
		A string
		B int
	}

	config.MustLoad(tmpFile, &val)

	require.Equal(t, "a", val.A)
	require.Equal(t, 1, val.B)
}

func createTempFile(text string) (string, error) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), hash.Md5Hex([]byte(text))+".json")
	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(tmpfile.Name(), []byte(text), os.ModeTemporary); err != nil {
		return "", err
	}

	filename := tmpfile.Name()
	if err = tmpfile.Close(); err != nil {
		return "", err
	}

	return filename, nil
}
