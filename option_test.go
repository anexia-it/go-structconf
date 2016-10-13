package structconf

import (
	"testing"

	"github.com/anexia-it/go-structconf/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -package=mocks -destination=mocks/storage.go github.com/anexia-it/go-structconf/storage Storage

func TestOptionStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mocks.NewMockStorage(ctrl)

	c := &TestConfigSimple{}

	conf, err := NewConfiguration(c, OptionStorage(storage))
	require.NoError(t, err)
	require.NotNil(t, conf)
	require.EqualValues(t, c, conf.config)

	require.EqualValues(t, storage, conf.storage)
}

func TestOptionEncoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := &TestConfigSimple{}

	encoding := mocks.NewMockEncoding(ctrl)
	conf, err := NewConfiguration(c, OptionEncoding(encoding))
	require.NoError(t, err)
	require.NotNil(t, conf)
	require.EqualValues(t, c, conf.config)
	require.EqualValues(t, encoding, conf.encoding)
}

func TestOptionDefaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := &TestConfigSimple{}
	defaults := &TestConfigSimple{
		Test: "test",
	}

	conf, err := NewConfiguration(c, OptionDefaults(defaults))
	require.NoError(t, err)
	require.NotNil(t, conf)
	// Check if the defaults were correctly applied to the config
	require.EqualValues(t, defaults, conf.config)
}
