package structconf

import (
	"errors"
	"testing"

	"github.com/anexia-it/go-structconf/mocks"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -package=mocks -destination=mocks/encoding.go github.com/anexia-it/go-structconf Encoding

type SimpleTestConfig struct {
	Test string `config:"test"`
}

func OptionError(err error) Option {
	return func(*Configuration) error {
		return err
	}
}

func TestNewConfiguration(t *testing.T) {
	// nil value is not allowed
	c, err := NewConfiguration(nil)
	require.EqualError(t, err, ErrConfigStructIsNil.Error())
	require.Nil(t, c)

	// config needs to be a pointer
	c, err = NewConfiguration(SimpleTestConfig{})
	require.EqualError(t, err, ErrNotAStructPointer.Error())
	require.Nil(t, c)

	// config needs to be a pointer to a struct
	testValue := "test"
	c, err = NewConfiguration(&testValue)
	require.EqualError(t, err, ErrNotAStructPointer.Error())
	require.Nil(t, c)

	conf := &SimpleTestConfig{}

	// Check if valid call without options works
	c, err = NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.EqualValues(t, conf, c.config)

	testErr := errors.New("test error")

	// Supply an option which returns an error and check if the error is passed back
	c, err = NewConfiguration(conf, OptionError(testErr))
	require.Error(t, err)
	require.Nil(t, c)

	// Now check if we retrieved a multierror
	multiErr, ok := err.(*multierror.Error)
	require.EqualValues(t, ok, true, "Returned error is not a multierror.Error")
	require.Len(t, multiErr.WrappedErrors(), 1)
	require.EqualError(t, multiErr.WrappedErrors()[0], testErr.Error())

	// Check if two supplied errors are both returned
	testErr2 := errors.New("test error2")
	c, err = NewConfiguration(conf, OptionError(testErr), OptionError(testErr2))
	require.Error(t, err)
	require.Nil(t, c)

	// Now check if we retrieved a multierror
	multiErr, ok = err.(*multierror.Error)
	require.EqualValues(t, ok, true, "Returned error is not a multierror.Error")
	require.Len(t, multiErr.WrappedErrors(), 2)
	require.EqualError(t, multiErr.WrappedErrors()[0], testErr.Error())
	require.EqualError(t, multiErr.WrappedErrors()[1], testErr2.Error())
}

func TestConfiguration_SetDefaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conf := &SimpleTestConfig{}
	//defaults := SimpleTestConfig{
	//	Test: "test",
	//}

	// Create empty configuration
	c, err := NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.EqualValues(t, conf, c.config)

	// Call SetDefaults without an encoding set-up
	require.EqualError(t, c.SetDefaults(nil), ErrEncodingNotConfigured.Error())

	mockEncoding := mocks.NewMockEncoding(ctrl)

	// Create configuration with mock encoding set
	c, err = NewConfiguration(conf, OptionEncoding(mockEncoding))
	require.NoError(t, err)
	require.NotNil(t, c)
	require.EqualValues(t, conf, c.config)
	require.EqualValues(t, mockEncoding, c.encoding)

	// Call defaults with nil defaults
	require.EqualError(t, c.SetDefaults(nil), ErrConfigStructIsNil.Error())

	// Call with mismatching type (no pointer)
	require.EqualError(t, c.SetDefaults("test"), ErrConfigStructTypeMismatch.Error())

	testStr := "test"
	// Call with mismatching type (pointer)
	require.EqualError(t, c.SetDefaults(&testStr), ErrConfigStructTypeMismatch.Error())

	// TODO: re-enable tests once SetDefaults actually works
	// Call with correct value (no pointer)
	//require.NoError(t, c.SetDefaults(defaults))
	//require.EqualValues(t, defaults.Test, conf.Test)
	//
	//// Call with correct value (pointer)
	//conf.Test = ""
	//require.NoError(t, c.SetDefaults(&defaults))
	//require.EqualValues(t, defaults.Test, conf.Test)
}
