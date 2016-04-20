/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubectl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/runtime"
)

// SecretGeneratorV1 supports stable generation of an opaque secret
type SecretGeneratorV1 struct {
	// Name of secret (required)
	Name string
	// Type of secret (optional)
	Type string
	// FileSources to derive the secret from (optional)
	FileSources []string
	// LiteralSources to derive the secret from (optional)
	LiteralSources []string
}

// Ensure it supports the generator pattern that uses parameter injection
var _ Generator = &SecretGeneratorV1{}

// Ensure it supports the generator pattern that uses parameters specified during construction
var _ StructuredGenerator = &SecretGeneratorV1{}

// Generate returns a secret using the specified parameters
func (s SecretGeneratorV1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	err := ValidateParams(s.ParamNames(), genericParams)
	if err != nil {
		return nil, err
	}
	delegate := &SecretGeneratorV1{}
	fromFileStrings, found := genericParams["from-file"]
	if found {
		fromFileArray, isArray := fromFileStrings.([]string)
		if !isArray {
			return nil, fmt.Errorf("expected []string, found :%v", fromFileStrings)
		}
		delegate.FileSources = fromFileArray
		delete(genericParams, "from-file")
	}
	fromLiteralStrings, found := genericParams["from-literal"]
	if found {
		fromLiteralArray, isArray := fromLiteralStrings.([]string)
		if !isArray {
			return nil, fmt.Errorf("expected []string, found :%v", fromFileStrings)
		}
		delegate.LiteralSources = fromLiteralArray
		delete(genericParams, "from-literal")
	}
	params := map[string]string{}
	for key, value := range genericParams {
		strVal, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("expected string, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}
	delegate.Name = params["name"]
	delegate.Type = params["type"]
	return delegate.StructuredGenerate()
}

// ParamNames returns the set of supported input parameters when using the parameter injection generator pattern
func (s SecretGeneratorV1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"name", true},
		{"type", false},
		{"from-file", false},
		{"from-literal", false},
		{"force", false},
	}
}

// StructuredGenerate outputs a secret object using the configured fields
func (s SecretGeneratorV1) StructuredGenerate() (runtime.Object, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	secret := &api.Secret{}
	secret.Name = s.Name
	secret.Data = map[string][]byte{}
	if len(s.Type) > 0 {
		secret.Type = api.SecretType(s.Type)
	}
	if len(s.FileSources) > 0 {
		if err := handleFromFileSources(secret, s.FileSources); err != nil {
			return nil, err
		}
	}
	if len(s.LiteralSources) > 0 {
		if err := handleFromLiteralSources(secret, s.LiteralSources); err != nil {
			return nil, err
		}
	}
	return secret, nil
}

// validate validates required fields are set to support structured generation
func (s SecretGeneratorV1) validate() error {
	if len(s.Name) == 0 {
		return fmt.Errorf("name must be specified")
	}
	return nil
}

// handleFromLiteralSources adds the specified literal source information into the provided secret
func handleFromLiteralSources(secret *api.Secret, literalSources []string) error {
	for _, literalSource := range literalSources {
		keyName, value, err := parseLiteralSource(literalSource)
		if err != nil {
			return err
		}
		err = addKeyFromLiteralToSecret(secret, keyName, []byte(value))
		if err != nil {
			return err
		}
	}
	return nil
}

// handleFromFileSources adds the specified file source information into the provided secret
func handleFromFileSources(secret *api.Secret, fileSources []string) error {
	for _, fileSource := range fileSources {
		keyName, filePath, err := parseFileSource(fileSource)
		if err != nil {
			return err
		}
		info, err := os.Stat(filePath)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return fmt.Errorf("error reading %s: %v", filePath, err.Err)
			default:
				return fmt.Errorf("error reading %s: %v", filePath, err)
			}
		}
		if info.IsDir() {
			if strings.Contains(fileSource, "=") {
				return fmt.Errorf("cannot give a key name for a directory path.")
			}
			fileList, err := ioutil.ReadDir(filePath)
			if err != nil {
				return fmt.Errorf("error listing files in %s: %v", filePath, err)
			}
			for _, item := range fileList {
				itemPath := path.Join(filePath, item.Name())
				if item.Mode().IsRegular() {
					keyName = item.Name()
					err = addKeyFromFileToSecret(secret, keyName, itemPath)
					if err != nil {
						return err
					}
				}
			}
		} else {
			err = addKeyFromFileToSecret(secret, keyName, filePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addKeyFromFileToSecret(secret *api.Secret, keyName, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return addKeyFromLiteralToSecret(secret, keyName, data)
}

func addKeyFromLiteralToSecret(secret *api.Secret, keyName string, data []byte) error {
	if !validation.IsSecretKey(keyName) {
		return fmt.Errorf("%v is not a valid key name for a secret", keyName)
	}
	if _, entryExists := secret.Data[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s, another key by that name already exists: %v.", keyName, secret.Data)
	}
	secret.Data[keyName] = data
	return nil
}
