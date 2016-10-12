package structconf

import (
	"testing"

	"github.com/anexia-it/go-structconf/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -package=mocks -destination=mocks/storage.go github.com/anexia-it/go-structconf Storage

func TestOptionStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mocks.NewMockStorage(ctrl)

	c := &SimpleTestConfig{}

	conf, err := NewConfiguration(c, OptionStorage(storage))
	require.NoError(t, err)
	require.NotNil(t, conf)
	require.EqualValues(t, c, conf.config)

	require.EqualValues(t, storage, conf.storage)
}

func TestOptionEncoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := &SimpleTestConfig{}

	encoding := mocks.NewMockEncoding(ctrl)
	conf, err := NewConfiguration(c, OptionEncoding(encoding))
	require.NoError(t, err)
	require.NotNil(t, conf)
	require.EqualValues(t, c, conf.config)
	require.EqualValues(t, encoding, conf.encoding)
}

func TestOptionDefaults(t *testing.T) {
	// TODO: implement me once OptionEncoding is not required anymore SetDefaults
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish()
	//
	//c := &SimpleTestConfig{}
	////defaults := SimpleTestConfig{
	////	Test: "test",
	////}
	//
	//encoding := mocks.NewMockEncoding(ctrl)
	//conf, err := NewConfiguration(c, OptionEncoding(encoding), OptionDefaults(nil))
	//require.NoError(t, err)
	//require.NotNil(t, conf)
	//require.EqualValues(t, c, conf.config)
	//require.EqualValues(t, encoding, conf.encoding)
}
