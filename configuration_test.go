package structconf

import (
	"errors"
	"testing"

	"sync"

	"io/ioutil"
	"os"

	"strings"

	"github.com/anexia-it/go-structconf/encoding/json"
	"github.com/anexia-it/go-structconf/storage/file"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
	"gopkg.in/anexia-it/go-structmapper.v1"
)

type TestConfigSimple struct {
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
	c, err = NewConfiguration(TestConfigSimple{})
	require.EqualError(t, err, ErrNotAStructPointer.Error())
	require.Nil(t, c)

	// config needs to be a pointer to a struct
	testValue := "test"
	c, err = NewConfiguration(&testValue)
	require.EqualError(t, err, ErrNotAStructPointer.Error())
	require.Nil(t, c)

	conf := &TestConfigSimple{}

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

func TestNewConfiguration_DefaultOptionsErrorPanic(t *testing.T) {
	// Temporarily override defaultOptions with an option that returns an error
	defOpts := defaultOptions
	defaultOptions = []Option{OptionError(errors.New("test"))}
	// Ensure that defaultOptions is restored to its original value
	defer func() {
		defaultOptions = defOpts
	}()
	conf := &TestConfigSimple{}

	require.Panics(t, func() {
		c, err := NewConfiguration(conf)
		// Code below is unreachable, but gives us a good hint, in case NewConfiguration does not panic
		// as expected
		require.NoError(t, err)
		require.NotNil(t, c)
	})
}

func TestNewConfiguration_EmptyTagName(t *testing.T) {
	conf := &TestConfigSimple{}

	c, err := NewConfiguration(conf, OptionTagName(""))
	require.Error(t, err)
	require.Nil(t, c)

	// Check if the error is a errwrap.Wrapper-type error
	wrapped, ok := err.(errwrap.Wrapper)
	require.EqualValues(t, true, ok, "Returned error is not a errwrap.Wrapper")
	wrappedErrors := wrapped.WrappedErrors()
	require.Len(t, wrappedErrors, 1)
	require.EqualError(t, wrappedErrors[0], structmapper.ErrTagNameEmpty.Error())
}

func TestNewConfiguration_SetDefaultsError(t *testing.T) {
	conf := &TestConfigSimple{}

	c, err := NewConfiguration(conf, OptionDefaults("test"))
	require.EqualError(t, err, ErrConfigStructTypeMismatch.Error())
	require.Nil(t, c)
}

func TestConfiguration_SetDefaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conf := &TestConfigSimple{}
	defaults := TestConfigSimple{
		Test: "test",
	}

	// Create empty configuration
	c, err := NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.EqualValues(t, conf, c.config)

	// Create configuration with mock encoding set
	c, err = NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.EqualValues(t, conf, c.config)

	// Call defaults with nil defaults
	require.EqualError(t, c.SetDefaults(nil), ErrConfigStructIsNil.Error())

	// Call with mismatching type (no pointer)
	require.EqualError(t, c.SetDefaults("test"), ErrConfigStructTypeMismatch.Error())

	testStr := "test"
	// Call with mismatching type (pointer)
	require.EqualError(t, c.SetDefaults(&testStr), ErrConfigStructTypeMismatch.Error())

	//Call with correct value (no pointer)
	require.NoError(t, c.SetDefaults(defaults))
	require.EqualValues(t, defaults.Test, conf.Test)

	// Call with correct value (pointer)
	conf.Test = ""
	require.NoError(t, c.SetDefaults(&defaults))
	require.EqualValues(t, defaults.Test, conf.Test)
}

var _ sync.Locker = (*TestConfigWithLocker)(nil)

type TestConfigWithLocker struct {
	TestConfigSimple

	lockCalled   int
	unlockCalled int
}

func (l *TestConfigWithLocker) Lock() {
	l.lockCalled++
}

func (l *TestConfigWithLocker) Unlock() {
	l.unlockCalled++
}

func TestConfiguration_SetDefaults_WithLocker(t *testing.T) {
	conf := &TestConfigWithLocker{}
	c, err := NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)

	defaults := &TestConfigWithLocker{
		TestConfigSimple: TestConfigSimple{
			Test: "test",
		},
	}

	require.NoError(t, c.SetDefaults(defaults))

	require.EqualValues(t, defaults.Test, conf.Test)
	require.EqualValues(t, 1, conf.lockCalled)
	require.EqualValues(t, 1, conf.unlockCalled)

}

func TestConfiguration_Load_NoEncoding(t *testing.T) {
	conf := &TestConfigSimple{}

	c, err := NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)

	require.EqualError(t, c.Load(), ErrEncodingNotConfigured.Error())
}

func TestConfiguration_Load_NoStorage(t *testing.T) {
	conf := &TestConfigSimple{}

	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	c, err := NewConfiguration(conf, OptionEncoding(enc))
	require.NoError(t, err)
	require.NotNil(t, c)

	require.EqualError(t, c.Load(), ErrStorageNotConfigured.Error())
}

func TestConfiguration_Load_Simple(t *testing.T) {
	conf := &TestConfigSimple{}

	// JSON string representing the configuration
	jsonString := `{"test":"test value"}`

	tempFile, err := ioutil.TempFile("", "go-strutconf-test-")
	require.NoError(t, err)
	require.NotNil(t, tempFile)
	// Clean up temp file
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	fileStorage := file.NewFileStorage(tempFile.Name(), 0640)

	// Write contents to file
	require.NoError(t, ioutil.WriteFile(tempFile.Name(), []byte(jsonString), 0640))

	c, err := NewConfiguration(conf, OptionEncoding(enc), OptionStorage(fileStorage))
	require.NoError(t, err)
	require.NotNil(t, c)

	require.NoError(t, c.Load())
	require.EqualValues(t, "test value", conf.Test)
}

func TestConfiguration_Save_NoEncoding(t *testing.T) {
	conf := &TestConfigSimple{}

	c, err := NewConfiguration(conf)
	require.NoError(t, err)
	require.NotNil(t, c)

	require.EqualError(t, c.Save(), ErrEncodingNotConfigured.Error())
}

func TestConfiguration_Save_NoStorage(t *testing.T) {
	conf := &TestConfigSimple{}

	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	c, err := NewConfiguration(conf, OptionEncoding(enc))
	require.NoError(t, err)
	require.NotNil(t, c)

	require.EqualError(t, c.Save(), ErrStorageNotConfigured.Error())
}

func TestConfiguration_Save_Simple(t *testing.T) {
	conf := &TestConfigSimple{
		Test: "test value",
	}

	// Expected json string
	jsonString := `{"test":"test value"}`

	tempFile, err := ioutil.TempFile("", "go-strutconf-test-")
	require.NoError(t, err)
	require.NotNil(t, tempFile)
	// Clean up temp file
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	fileStorage := file.NewFileStorage(tempFile.Name(), 0640)

	// Write contents to file
	require.NoError(t, ioutil.WriteFile(tempFile.Name(), []byte(jsonString), 0640))

	c, err := NewConfiguration(conf, OptionEncoding(enc), OptionStorage(fileStorage))
	require.NoError(t, err)
	require.NotNil(t, c)

	require.NoError(t, c.Save())

	writtenBytes, err := ioutil.ReadFile(tempFile.Name())
	require.NoError(t, err)
	require.EqualValues(t, jsonString, strings.TrimSuffix(string(writtenBytes), "\n"))
}
