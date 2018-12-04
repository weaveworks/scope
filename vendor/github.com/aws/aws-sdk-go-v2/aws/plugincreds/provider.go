// +build go1.8

/*
Package plugincreds implements a credentials provider sourced from a Go
plugin. This package allows you to use a Go plugin to retrieve AWS credentials
for the SDK to use for service API calls.

As of Go 1.8 plugins are only supported on the Linux platform.

Plugin Symbol Name

The "AWSSDKRetrieveCredentials" is the symbol name that will be used to
lookup the credentials provider getter from the plugin. If you want to use a
custom symbol name you should use GetRetrieveFnByName to lookup the
symbol by a custom name.

This symbol is a function that returns two additional functions. One to
retrieve the credentials, and another to determine if the credentials have
expired.

Plugin Symbol Signature

The plugin credential provider requires the symbol to match the
following signature.

  func() (key, secret, token string, exp time.Time, err error)

Plugin Implementation Exmaple

The following is an example implementation of a SDK credential provider using
the plugin provider in this package. See the SDK's example/aws/credential/plugincreds/plugin
folder for a runnable example of this.

  func main() {}

  var myCredProvider provider

  // Build: go build -o plugin.so -buildmode=plugin plugin.go
  func init() {
  	// Initialize a mock credential provider with stubs
  	myCredProvider = provider{"a","b","c"}
  }

  // AWSSDKRetrieveCredentials is the symbol SDK will lookup and use to
  // retrieve AWS credentials with.
  func AWSSDKRetrieveCredentials() (key, secret, token string, exp time.Time, err error) {
    key, secret, token, err = getCreds()
	if err != nil {
		return "", "", "", time.Time{}, err
	}

	return key, secrete, token, time.Now().Add(2 * time.Hour), nil
  }

  func getCreds() (key, secret, token string, err error){
	  // Get credentials
	  return key, secret, token, err
  }

Configuring SDK for Plugin Credentials

To configure the SDK to use a plugin's credential provider you'll need to first
open the plugin file using the plugin standard library package. Once you have a
handle to the plugin you can use the NewCredentials function of this package to
create a new credentials.Credentials value that can be set as the credentials
loader of a Session or Config. See the SDK's example/aws/credential/plugincreds
folder for a runnable example of this.

  // Open plugin, and load it into the process.
  p, err := plugin.Open("somefile.so")
  if err != nil {
  	return nil, err
  }

  // Create a new Credentials value which will source the provider's Retrieve
  // and IsExpired functions from the plugin.
  creds, err := plugincreds.New(p)
  if err != nil {
  	return nil, err
  }

  cfg.Credentails = creds
*/
package plugincreds

import (
	"fmt"
	"plugin"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
)

// ProviderSymbolName the symbol name the SDK will use to retrieve the
// credentials from.
const ProviderSymbolName = `AWSSDKRetrieveCredentials`

// ProviderName is the name this credentials provider will label any returned
// credentials Value with.
const ProviderName = `PluginCredentialsProvider`

const (
	// ErrCodeLookupSymbolError failed to lookup symbol
	ErrCodeLookupSymbolError = "LookupSymbolError"

	// ErrCodeInvalidSymbolError symbol invalid
	ErrCodeInvalidSymbolError = "InvalidSymbolError"

	// ErrCodePluginRetrieveNil Retrieve function was nil
	ErrCodePluginRetrieveNil = "PluginRetrieveNilError"

	// ErrCodePluginIsExpiredNil IsExpired Function was nil
	ErrCodePluginIsExpiredNil = "PluginIsExpiredNilError"

	// ErrCodePluginProviderRetrieve plugin provider's retrieve returned error
	ErrCodePluginProviderRetrieve = "PluginProviderRetrieveError"
)

// Provider is the credentials provider that will use the plugin provided
// Retrieve and IsExpired functions to retrieve credentials.
type Provider struct {
	aws.SafeCredentialsProvider
}

// New returns a new Credentials loader using the plugin provider.
// If the symbol isn't found or is invalid in the plugin an error will be
// returned.
func New(p *plugin.Plugin) (*Provider, error) {
	fn, err := GetRetrieveFn(p)
	if err != nil {
		return nil, err
	}

	provider := &Provider{}
	provider.RetrieveFn = buildRetrieveFn(fn)

	return provider, nil
}

func buildRetrieveFn(fn func() (k, s, t string, ext time.Time, err error)) func() (aws.Credentials, error) {
	return func() (aws.Credentials, error) {
		k, s, t, exp, err := fn()
		if err != nil {
			return aws.Credentials{}, awserr.New(ErrCodePluginProviderRetrieve,
				"failed to retrieve credentials with plugin provider", err)
		}

		creds := aws.Credentials{
			AccessKeyID:     k,
			SecretAccessKey: s,
			SessionToken:    t,
			Source:          ProviderName,

			CanExpire: !exp.IsZero(),
			Expires:   exp,
		}

		return creds, nil
	}
}

// GetRetrieveFn returns the plugin's Retrieve and IsExpired functions
// returned by the plugin's credential provider getter.
//
// Uses ProviderSymbolName as the symbol name when lookup up the symbol. If you
// want to use a different symbol name, use GetRetrieveFnByName.
func GetRetrieveFn(p *plugin.Plugin) (func() (key, secret, token string, exp time.Time, err error), error) {
	return GetRetrieveFnByName(p, ProviderSymbolName)
}

// GetRetrieveFnByName returns the plugin's Retrieve and IsExpired functions
// returned by the plugin's credential provider getter.
//
// Same as GetRetrieveFn, but takes a custom symbolName to lookup with.
func GetRetrieveFnByName(p *plugin.Plugin, symbolName string) (func() (key, secret, token string, exp time.Time, err error), error) {
	sym, err := p.Lookup(symbolName)
	if err != nil {
		return nil, awserr.New(ErrCodeLookupSymbolError,
			fmt.Sprintf("failed to lookup %s plugin provider symbol", symbolName), err)
	}

	fn, ok := sym.(func() (key, secret, token string, exp time.Time, err error))
	if !ok {
		return nil, awserr.New(ErrCodeInvalidSymbolError,
			fmt.Sprintf("symbol %T, does not match the 'func() (key, secret, token string, exp time.Time, err error)'  type", sym), nil)
	}

	return fn, nil
}
