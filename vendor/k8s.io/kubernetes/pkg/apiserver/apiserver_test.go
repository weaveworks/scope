/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	apierrs "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api/v1"
	apiservertesting "k8s.io/kubernetes/pkg/apiserver/testing"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/diff"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/version"
	"k8s.io/kubernetes/pkg/watch"
	"k8s.io/kubernetes/pkg/watch/versioned"
	"k8s.io/kubernetes/plugin/pkg/admission/admit"
	"k8s.io/kubernetes/plugin/pkg/admission/deny"

	"github.com/emicklei/go-restful"
)

func convert(obj runtime.Object) (runtime.Object, error) {
	return obj, nil
}

// This creates fake API versions, similar to api/latest.go.
var testAPIGroup = "test.group"
var testAPIGroup2 = "test.group2"
var testInternalGroupVersion = unversioned.GroupVersion{Group: testAPIGroup, Version: runtime.APIVersionInternal}
var testGroupVersion = unversioned.GroupVersion{Group: testAPIGroup, Version: "version"}
var newGroupVersion = unversioned.GroupVersion{Group: testAPIGroup, Version: "version2"}
var testGroup2Version = unversioned.GroupVersion{Group: testAPIGroup2, Version: "version"}
var testInternalGroup2Version = unversioned.GroupVersion{Group: testAPIGroup2, Version: runtime.APIVersionInternal}
var prefix = "apis"

var grouplessGroupVersion = unversioned.GroupVersion{Group: "", Version: "v1"}
var grouplessInternalGroupVersion = unversioned.GroupVersion{Group: "", Version: runtime.APIVersionInternal}
var grouplessPrefix = "api"

var groupVersions = []unversioned.GroupVersion{grouplessGroupVersion, testGroupVersion, newGroupVersion}

var codec = api.Codecs.LegacyCodec(groupVersions...)
var grouplessCodec = api.Codecs.LegacyCodec(grouplessGroupVersion)
var testCodec = api.Codecs.LegacyCodec(testGroupVersion)
var newCodec = api.Codecs.LegacyCodec(newGroupVersion)

var accessor = meta.NewAccessor()
var versioner runtime.ResourceVersioner = accessor
var selfLinker runtime.SelfLinker = accessor
var mapper, namespaceMapper meta.RESTMapper // The mappers with namespace and with legacy namespace scopes.
var admissionControl admission.Interface
var requestContextMapper api.RequestContextMapper

func interfacesFor(version unversioned.GroupVersion) (*meta.VersionInterfaces, error) {
	switch version {
	case testGroupVersion:
		return &meta.VersionInterfaces{
			ObjectConvertor:  api.Scheme,
			MetadataAccessor: accessor,
		}, nil
	case newGroupVersion:
		return &meta.VersionInterfaces{
			ObjectConvertor:  api.Scheme,
			MetadataAccessor: accessor,
		}, nil
	case grouplessGroupVersion:
		return &meta.VersionInterfaces{
			ObjectConvertor:  api.Scheme,
			MetadataAccessor: accessor,
		}, nil
	case testGroup2Version:
		return &meta.VersionInterfaces{
			ObjectConvertor:  api.Scheme,
			MetadataAccessor: accessor,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported storage version: %s (valid: %v)", version, groupVersions)
	}
}

func newMapper() *meta.DefaultRESTMapper {
	return meta.NewDefaultRESTMapper([]unversioned.GroupVersion{testGroupVersion, newGroupVersion}, interfacesFor)
}

func addGrouplessTypes() {
	type ListOptions struct {
		runtime.Object
		unversioned.TypeMeta `json:",inline"`
		LabelSelector        string `json:"labelSelector,omitempty"`
		FieldSelector        string `json:"fieldSelector,omitempty"`
		Watch                bool   `json:"watch,omitempty"`
		ResourceVersion      string `json:"resourceVersion,omitempty"`
		TimeoutSeconds       *int64 `json:"timeoutSeconds,omitempty"`
	}
	api.Scheme.AddKnownTypes(grouplessGroupVersion,
		&apiservertesting.Simple{}, &apiservertesting.SimpleList{}, &ListOptions{},
		&api.DeleteOptions{}, &apiservertesting.SimpleGetOptions{}, &apiservertesting.SimpleRoot{})
	api.Scheme.AddKnownTypes(grouplessGroupVersion, &api.Pod{})
	api.Scheme.AddKnownTypes(grouplessInternalGroupVersion,
		&apiservertesting.Simple{}, &apiservertesting.SimpleList{}, &api.ListOptions{},
		&apiservertesting.SimpleGetOptions{}, &apiservertesting.SimpleRoot{})
}

func addTestTypes() {
	type ListOptions struct {
		runtime.Object
		unversioned.TypeMeta `json:",inline"`
		LabelSelector        string `json:"labelSelector,omitempty"`
		FieldSelector        string `json:"fieldSelector,omitempty"`
		Watch                bool   `json:"watch,omitempty"`
		ResourceVersion      string `json:"resourceVersion,omitempty"`
		TimeoutSeconds       *int64 `json:"timeoutSeconds,omitempty"`
	}
	api.Scheme.AddKnownTypes(testGroupVersion,
		&apiservertesting.Simple{}, &apiservertesting.SimpleList{}, &ListOptions{},
		&api.DeleteOptions{}, &apiservertesting.SimpleGetOptions{}, &apiservertesting.SimpleRoot{},
		&SimpleXGSubresource{})
	api.Scheme.AddKnownTypes(testGroupVersion, &api.Pod{})
	api.Scheme.AddKnownTypes(testInternalGroupVersion,
		&apiservertesting.Simple{}, &apiservertesting.SimpleList{}, &api.ListOptions{},
		&apiservertesting.SimpleGetOptions{}, &apiservertesting.SimpleRoot{},
		&SimpleXGSubresource{})
	// Register SimpleXGSubresource in both testGroupVersion and testGroup2Version, and also their
	// their corresponding internal versions, to verify that the desired group version object is
	// served in the tests.
	api.Scheme.AddKnownTypes(testGroup2Version, &SimpleXGSubresource{})
	api.Scheme.AddKnownTypes(testInternalGroup2Version, &SimpleXGSubresource{})
	versioned.AddToGroupVersion(api.Scheme, testGroupVersion)
}

func addNewTestTypes() {
	type ListOptions struct {
		runtime.Object
		unversioned.TypeMeta `json:",inline"`
		LabelSelector        string `json:"labelSelector,omitempty"`
		FieldSelector        string `json:"fieldSelector,omitempty"`
		Watch                bool   `json:"watch,omitempty"`
		ResourceVersion      string `json:"resourceVersion,omitempty"`
		TimeoutSeconds       *int64 `json:"timeoutSeconds,omitempty"`
	}
	api.Scheme.AddKnownTypes(newGroupVersion,
		&apiservertesting.Simple{}, &apiservertesting.SimpleList{}, &ListOptions{},
		&api.DeleteOptions{}, &apiservertesting.SimpleGetOptions{}, &apiservertesting.SimpleRoot{},
		&v1.Pod{},
	)
	versioned.AddToGroupVersion(api.Scheme, newGroupVersion)
}

func init() {
	// Certain API objects are returned regardless of the contents of storage:
	// api.Status is returned in errors

	addGrouplessTypes()
	addTestTypes()
	addNewTestTypes()

	nsMapper := newMapper()

	// enumerate all supported versions, get the kinds, and register with
	// the mapper how to address our resources
	for _, gv := range groupVersions {
		for kind := range api.Scheme.KnownTypes(gv) {
			gvk := gv.WithKind(kind)
			root := bool(kind == "SimpleRoot")
			if root {
				nsMapper.Add(gvk, meta.RESTScopeRoot)
			} else {
				nsMapper.Add(gvk, meta.RESTScopeNamespace)
			}
		}
	}

	mapper = nsMapper
	namespaceMapper = nsMapper
	admissionControl = admit.NewAlwaysAdmit()
	requestContextMapper = api.NewRequestContextMapper()

	api.Scheme.AddFieldLabelConversionFunc(grouplessGroupVersion.String(), "Simple",
		func(label, value string) (string, string, error) {
			return label, value, nil
		},
	)
	api.Scheme.AddFieldLabelConversionFunc(testGroupVersion.String(), "Simple",
		func(label, value string) (string, string, error) {
			return label, value, nil
		},
	)
	api.Scheme.AddFieldLabelConversionFunc(newGroupVersion.String(), "Simple",
		func(label, value string) (string, string, error) {
			return label, value, nil
		},
	)
}

// defaultAPIServer exposes nested objects for testability.
type defaultAPIServer struct {
	http.Handler
	container *restful.Container
}

// uses the default settings
func handle(storage map[string]rest.Storage) http.Handler {
	return handleInternal(storage, admissionControl, selfLinker)
}

// tests with a deny admission controller
func handleDeny(storage map[string]rest.Storage) http.Handler {
	return handleInternal(storage, deny.NewAlwaysDeny(), selfLinker)
}

// tests using the new namespace scope mechanism
func handleNamespaced(storage map[string]rest.Storage) http.Handler {
	return handleInternal(storage, admissionControl, selfLinker)
}

// tests using a custom self linker
func handleLinker(storage map[string]rest.Storage, selfLinker runtime.SelfLinker) http.Handler {
	return handleInternal(storage, admissionControl, selfLinker)
}

func newTestRequestInfoResolver() *RequestInfoResolver {
	return &RequestInfoResolver{sets.NewString("api", "apis"), sets.NewString("api")}
}

func handleInternal(storage map[string]rest.Storage, admissionControl admission.Interface, selfLinker runtime.SelfLinker) http.Handler {
	container := restful.NewContainer()
	container.Router(restful.CurlyRouter{})
	mux := container.ServeMux

	template := APIGroupVersion{
		Storage: storage,

		RequestInfoResolver: newTestRequestInfoResolver(),

		Creater:   api.Scheme,
		Convertor: api.Scheme,
		Typer:     api.Scheme,
		Linker:    selfLinker,
		Mapper:    namespaceMapper,

		ParameterCodec: api.ParameterCodec,

		Admit:   admissionControl,
		Context: requestContextMapper,
	}

	// groupless v1 version
	{
		group := template
		group.Root = "/" + grouplessPrefix
		group.GroupVersion = grouplessGroupVersion
		group.OptionsExternalVersion = &grouplessGroupVersion
		group.Serializer = api.Codecs
		group.StreamSerializer = api.StreamCodecs
		if err := (&group).InstallREST(container); err != nil {
			panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
		}
	}

	// group version 1
	{
		group := template
		group.Root = "/" + prefix
		group.GroupVersion = testGroupVersion
		group.OptionsExternalVersion = &testGroupVersion
		group.Serializer = api.Codecs
		group.StreamSerializer = api.StreamCodecs
		if err := (&group).InstallREST(container); err != nil {
			panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
		}
	}

	// group version 2
	{
		group := template
		group.Root = "/" + prefix
		group.GroupVersion = newGroupVersion
		group.OptionsExternalVersion = &newGroupVersion
		group.Serializer = api.Codecs
		group.StreamSerializer = api.StreamCodecs
		if err := (&group).InstallREST(container); err != nil {
			panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
		}
	}

	ws := new(restful.WebService)
	InstallSupport(mux, ws)
	container.Add(ws)
	return &defaultAPIServer{mux, container}
}

func TestSimpleSetupRight(t *testing.T) {
	s := &apiservertesting.Simple{ObjectMeta: api.ObjectMeta{Name: "aName"}}
	wire, err := runtime.Encode(codec, s)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := runtime.Decode(codec, wire)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, s2) {
		t.Fatalf("encode/decode broken:\n%#v\n%#v\n", s, s2)
	}
}

func TestSimpleOptionsSetupRight(t *testing.T) {
	s := &apiservertesting.SimpleGetOptions{}
	wire, err := runtime.Encode(codec, s)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := runtime.Decode(codec, wire)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, s2) {
		t.Fatalf("encode/decode broken:\n%#v\n%#v\n", s, s2)
	}
}

type SimpleRESTStorage struct {
	errors map[string]error
	list   []apiservertesting.Simple
	item   apiservertesting.Simple

	updated *apiservertesting.Simple
	created *apiservertesting.Simple

	stream *SimpleStream

	deleted       string
	deleteOptions *api.DeleteOptions

	actualNamespace  string
	namespacePresent bool

	// These are set when Watch is called
	fakeWatch                  *watch.FakeWatcher
	requestedLabelSelector     labels.Selector
	requestedFieldSelector     fields.Selector
	requestedResourceVersion   string
	requestedResourceNamespace string

	// The id requested, and location to return for ResourceLocation
	requestedResourceLocationID string
	resourceLocation            *url.URL
	resourceLocationTransport   http.RoundTripper
	expectedResourceNamespace   string

	// If non-nil, called inside the WorkFunc when answering update, delete, create.
	// obj receives the original input to the update, delete, or create call.
	injectedFunction func(obj runtime.Object) (returnObj runtime.Object, err error)
}

func (storage *SimpleRESTStorage) Export(ctx api.Context, name string, opts unversioned.ExportOptions) (runtime.Object, error) {
	obj, err := storage.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	s, ok := obj.(*apiservertesting.Simple)
	if !ok {
		return nil, fmt.Errorf("unexpected object")
	}

	// Set a marker to verify the method was called
	s.Other = "exported"
	return obj, storage.errors["export"]
}

func (storage *SimpleRESTStorage) List(ctx api.Context, options *api.ListOptions) (runtime.Object, error) {
	storage.checkContext(ctx)
	result := &apiservertesting.SimpleList{
		Items: storage.list,
	}
	storage.requestedLabelSelector = labels.Everything()
	if options != nil && options.LabelSelector != nil {
		storage.requestedLabelSelector = options.LabelSelector
	}
	storage.requestedFieldSelector = fields.Everything()
	if options != nil && options.FieldSelector != nil {
		storage.requestedFieldSelector = options.FieldSelector
	}
	return result, storage.errors["list"]
}

type SimpleStream struct {
	version     string
	accept      string
	contentType string
	err         error

	io.Reader
	closed bool
}

func (s *SimpleStream) Close() error {
	s.closed = true
	return nil
}

func (obj *SimpleStream) GetObjectKind() unversioned.ObjectKind { return unversioned.EmptyObjectKind }

func (s *SimpleStream) InputStream(version, accept string) (io.ReadCloser, bool, string, error) {
	s.version = version
	s.accept = accept
	return s, false, s.contentType, s.err
}

type OutputConnect struct {
	response string
}

func (h *OutputConnect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(h.response))
}

func (storage *SimpleRESTStorage) Get(ctx api.Context, id string) (runtime.Object, error) {
	storage.checkContext(ctx)
	if id == "binary" {
		return storage.stream, storage.errors["get"]
	}
	copied, err := api.Scheme.Copy(&storage.item)
	if err != nil {
		panic(err)
	}
	return copied, storage.errors["get"]
}

func (storage *SimpleRESTStorage) checkContext(ctx api.Context) {
	storage.actualNamespace, storage.namespacePresent = api.NamespaceFrom(ctx)
}

func (storage *SimpleRESTStorage) Delete(ctx api.Context, id string, options *api.DeleteOptions) (runtime.Object, error) {
	storage.checkContext(ctx)
	storage.deleted = id
	storage.deleteOptions = options
	if err := storage.errors["delete"]; err != nil {
		return nil, err
	}
	var obj runtime.Object = &unversioned.Status{Status: unversioned.StatusSuccess}
	var err error
	if storage.injectedFunction != nil {
		obj, err = storage.injectedFunction(&apiservertesting.Simple{ObjectMeta: api.ObjectMeta{Name: id}})
	}
	return obj, err
}

func (storage *SimpleRESTStorage) New() runtime.Object {
	return &apiservertesting.Simple{}
}

func (storage *SimpleRESTStorage) NewList() runtime.Object {
	return &apiservertesting.SimpleList{}
}

func (storage *SimpleRESTStorage) Create(ctx api.Context, obj runtime.Object) (runtime.Object, error) {
	storage.checkContext(ctx)
	storage.created = obj.(*apiservertesting.Simple)
	if err := storage.errors["create"]; err != nil {
		return nil, err
	}
	var err error
	if storage.injectedFunction != nil {
		obj, err = storage.injectedFunction(obj)
	}
	return obj, err
}

func (storage *SimpleRESTStorage) Update(ctx api.Context, obj runtime.Object) (runtime.Object, bool, error) {
	storage.checkContext(ctx)
	storage.updated = obj.(*apiservertesting.Simple)
	if err := storage.errors["update"]; err != nil {
		return nil, false, err
	}
	var err error
	if storage.injectedFunction != nil {
		obj, err = storage.injectedFunction(obj)
	}
	return obj, false, err
}

// Implement ResourceWatcher.
func (storage *SimpleRESTStorage) Watch(ctx api.Context, options *api.ListOptions) (watch.Interface, error) {
	storage.checkContext(ctx)
	storage.requestedLabelSelector = labels.Everything()
	if options != nil && options.LabelSelector != nil {
		storage.requestedLabelSelector = options.LabelSelector
	}
	storage.requestedFieldSelector = fields.Everything()
	if options != nil && options.FieldSelector != nil {
		storage.requestedFieldSelector = options.FieldSelector
	}
	storage.requestedResourceVersion = ""
	if options != nil {
		storage.requestedResourceVersion = options.ResourceVersion
	}
	storage.requestedResourceNamespace = api.NamespaceValue(ctx)
	if err := storage.errors["watch"]; err != nil {
		return nil, err
	}
	storage.fakeWatch = watch.NewFake()
	return storage.fakeWatch, nil
}

// Implement Redirector.
var _ = rest.Redirector(&SimpleRESTStorage{})

// Implement Redirector.
func (storage *SimpleRESTStorage) ResourceLocation(ctx api.Context, id string) (*url.URL, http.RoundTripper, error) {
	storage.checkContext(ctx)
	// validate that the namespace context on the request matches the expected input
	storage.requestedResourceNamespace = api.NamespaceValue(ctx)
	if storage.expectedResourceNamespace != storage.requestedResourceNamespace {
		return nil, nil, fmt.Errorf("Expected request namespace %s, but got namespace %s", storage.expectedResourceNamespace, storage.requestedResourceNamespace)
	}
	storage.requestedResourceLocationID = id
	if err := storage.errors["resourceLocation"]; err != nil {
		return nil, nil, err
	}
	// Make a copy so the internal URL never gets mutated
	locationCopy := *storage.resourceLocation
	return &locationCopy, storage.resourceLocationTransport, nil
}

// Implement Connecter
type ConnecterRESTStorage struct {
	connectHandler http.Handler
	handlerFunc    func() http.Handler

	emptyConnectOptions    runtime.Object
	receivedConnectOptions runtime.Object
	receivedID             string
	receivedResponder      rest.Responder
	takesPath              string
}

// Implement Connecter
var _ = rest.Connecter(&ConnecterRESTStorage{})

func (s *ConnecterRESTStorage) New() runtime.Object {
	return &apiservertesting.Simple{}
}

func (s *ConnecterRESTStorage) Connect(ctx api.Context, id string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	s.receivedConnectOptions = options
	s.receivedID = id
	s.receivedResponder = responder
	if s.handlerFunc != nil {
		return s.handlerFunc(), nil
	}
	return s.connectHandler, nil
}

func (s *ConnecterRESTStorage) ConnectMethods() []string {
	return []string{"GET", "POST", "PUT", "DELETE"}
}

func (s *ConnecterRESTStorage) NewConnectOptions() (runtime.Object, bool, string) {
	if len(s.takesPath) > 0 {
		return s.emptyConnectOptions, true, s.takesPath
	}
	return s.emptyConnectOptions, false, ""
}

type LegacyRESTStorage struct {
	*SimpleRESTStorage
}

func (storage LegacyRESTStorage) Delete(ctx api.Context, id string) (runtime.Object, error) {
	return storage.SimpleRESTStorage.Delete(ctx, id, nil)
}

type MetadataRESTStorage struct {
	*SimpleRESTStorage
	types []string
}

func (m *MetadataRESTStorage) ProducesMIMETypes(method string) []string {
	return m.types
}

var _ rest.StorageMetadata = &MetadataRESTStorage{}

type GetWithOptionsRESTStorage struct {
	*SimpleRESTStorage
	optionsReceived runtime.Object
	takesPath       string
}

func (r *GetWithOptionsRESTStorage) Get(ctx api.Context, name string, options runtime.Object) (runtime.Object, error) {
	if _, ok := options.(*apiservertesting.SimpleGetOptions); !ok {
		return nil, fmt.Errorf("Unexpected options object: %#v", options)
	}
	r.optionsReceived = options
	return r.SimpleRESTStorage.Get(ctx, name)
}

func (r *GetWithOptionsRESTStorage) NewGetOptions() (runtime.Object, bool, string) {
	if len(r.takesPath) > 0 {
		return &apiservertesting.SimpleGetOptions{}, true, r.takesPath
	}
	return &apiservertesting.SimpleGetOptions{}, false, ""
}

var _ rest.GetterWithOptions = &GetWithOptionsRESTStorage{}

type NamedCreaterRESTStorage struct {
	*SimpleRESTStorage
	createdName string
}

func (storage *NamedCreaterRESTStorage) Create(ctx api.Context, name string, obj runtime.Object) (runtime.Object, error) {
	storage.checkContext(ctx)
	storage.created = obj.(*apiservertesting.Simple)
	storage.createdName = name
	if err := storage.errors["create"]; err != nil {
		return nil, err
	}
	var err error
	if storage.injectedFunction != nil {
		obj, err = storage.injectedFunction(obj)
	}
	return obj, err
}

type SimpleTypedStorage struct {
	errors   map[string]error
	item     runtime.Object
	baseType runtime.Object

	actualNamespace  string
	namespacePresent bool
}

func (storage *SimpleTypedStorage) New() runtime.Object {
	return storage.baseType
}

func (storage *SimpleTypedStorage) Get(ctx api.Context, id string) (runtime.Object, error) {
	storage.checkContext(ctx)
	copied, err := api.Scheme.Copy(storage.item)
	if err != nil {
		panic(err)
	}
	return copied, storage.errors["get"]
}

func (storage *SimpleTypedStorage) checkContext(ctx api.Context) {
	storage.actualNamespace, storage.namespacePresent = api.NamespaceFrom(ctx)
}

func extractBody(response *http.Response, object runtime.Object) (string, error) {
	return extractBodyDecoder(response, object, codec)
}

func extractBodyDecoder(response *http.Response, object runtime.Object, decoder runtime.Decoder) (string, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return string(body), err
	}
	return string(body), runtime.DecodeInto(decoder, body, object)
}

func TestNotFound(t *testing.T) {
	type T struct {
		Method string
		Path   string
		Status int
	}
	cases := map[string]T{
		// Positive checks to make sure everything is wired correctly
		"groupless GET root":       {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots", http.StatusOK},
		"groupless GET namespaced": {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples", http.StatusOK},

		"groupless GET long prefix": {"GET", "/" + grouplessPrefix + "/", http.StatusNotFound},

		"groupless root PATCH method":                 {"PATCH", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots", http.StatusMethodNotAllowed},
		"groupless root GET missing storage":          {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/blah", http.StatusNotFound},
		"groupless root GET with extra segment":       {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots/bar/baz", http.StatusNotFound},
		"groupless root DELETE without extra segment": {"DELETE", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots", http.StatusMethodNotAllowed},
		"groupless root DELETE with extra segment":    {"DELETE", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots/bar/baz", http.StatusNotFound},
		"groupless root PUT without extra segment":    {"PUT", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots", http.StatusMethodNotAllowed},
		"groupless root PUT with extra segment":       {"PUT", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simpleroots/bar/baz", http.StatusNotFound},
		"groupless root watch missing storage":        {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/watch/", http.StatusNotFound},

		"groupless namespaced PATCH method":                 {"PATCH", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples", http.StatusMethodNotAllowed},
		"groupless namespaced GET long prefix":              {"GET", "/" + grouplessPrefix + "/", http.StatusNotFound},
		"groupless namespaced GET missing storage":          {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/blah", http.StatusNotFound},
		"groupless namespaced GET with extra segment":       {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples/bar/baz", http.StatusNotFound},
		"groupless namespaced POST with extra segment":      {"POST", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples/bar", http.StatusMethodNotAllowed},
		"groupless namespaced DELETE without extra segment": {"DELETE", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples", http.StatusMethodNotAllowed},
		"groupless namespaced DELETE with extra segment":    {"DELETE", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples/bar/baz", http.StatusNotFound},
		"groupless namespaced PUT without extra segment":    {"PUT", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples", http.StatusMethodNotAllowed},
		"groupless namespaced PUT with extra segment":       {"PUT", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/ns/simples/bar/baz", http.StatusNotFound},
		"groupless namespaced watch missing storage":        {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/watch/", http.StatusNotFound},
		"groupless namespaced watch with bad method":        {"POST", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/watch/namespaces/ns/simples/bar", http.StatusMethodNotAllowed},

		// Positive checks to make sure everything is wired correctly
		"GET root": {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots", http.StatusOK},
		// TODO: JTL: "GET root item":       {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots/bar", http.StatusOK},
		"GET namespaced": {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples", http.StatusOK},
		// TODO: JTL: "GET namespaced item": {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples/bar", http.StatusOK},

		"GET long prefix": {"GET", "/" + prefix + "/", http.StatusNotFound},

		"root PATCH method":           {"PATCH", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots", http.StatusMethodNotAllowed},
		"root GET missing storage":    {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/blah", http.StatusNotFound},
		"root GET with extra segment": {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots/bar/baz", http.StatusNotFound},
		// TODO: JTL: "root POST with extra segment":      {"POST", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots/bar", http.StatusMethodNotAllowed},
		"root DELETE without extra segment": {"DELETE", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots", http.StatusMethodNotAllowed},
		"root DELETE with extra segment":    {"DELETE", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots/bar/baz", http.StatusNotFound},
		"root PUT without extra segment":    {"PUT", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots", http.StatusMethodNotAllowed},
		"root PUT with extra segment":       {"PUT", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots/bar/baz", http.StatusNotFound},
		"root watch missing storage":        {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/watch/", http.StatusNotFound},
		// TODO: JTL: "root watch with bad method":        {"POST", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/watch/simpleroot/bar", http.StatusMethodNotAllowed},

		"namespaced PATCH method":                 {"PATCH", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples", http.StatusMethodNotAllowed},
		"namespaced GET long prefix":              {"GET", "/" + prefix + "/", http.StatusNotFound},
		"namespaced GET missing storage":          {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/blah", http.StatusNotFound},
		"namespaced GET with extra segment":       {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples/bar/baz", http.StatusNotFound},
		"namespaced POST with extra segment":      {"POST", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples/bar", http.StatusMethodNotAllowed},
		"namespaced DELETE without extra segment": {"DELETE", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples", http.StatusMethodNotAllowed},
		"namespaced DELETE with extra segment":    {"DELETE", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples/bar/baz", http.StatusNotFound},
		"namespaced PUT without extra segment":    {"PUT", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples", http.StatusMethodNotAllowed},
		"namespaced PUT with extra segment":       {"PUT", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/ns/simples/bar/baz", http.StatusNotFound},
		"namespaced watch missing storage":        {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/watch/", http.StatusNotFound},
		"namespaced watch with bad method":        {"POST", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/watch/namespaces/ns/simples/bar", http.StatusMethodNotAllowed},
	}
	handler := handle(map[string]rest.Storage{
		"simples":     &SimpleRESTStorage{},
		"simpleroots": &SimpleRESTStorage{},
	})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}
	for k, v := range cases {
		request, err := http.NewRequest(v.Method, server.URL+v.Path, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		response, err := client.Do(request)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if response.StatusCode != v.Status {
			t.Errorf("Expected %d for %s (%s), Got %#v", v.Status, v.Method, k, response)
			t.Errorf("MAPPER: %v", mapper)
		}
	}
}

type UnimplementedRESTStorage struct{}

func (UnimplementedRESTStorage) New() runtime.Object {
	return &apiservertesting.Simple{}
}

// TestUnimplementedRESTStorage ensures that if a rest.Storage does not implement a given
// method, that it is literally not registered with the server.  In the past,
// we registered everything, and returned method not supported if it didn't support
// a verb.  Now we literally do not register a storage if it does not implement anything.
// TODO: in future, we should update proxy/redirect
func TestUnimplementedRESTStorage(t *testing.T) {
	type T struct {
		Method  string
		Path    string
		ErrCode int
	}
	cases := map[string]T{
		"groupless GET object":      {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/foo/bar", http.StatusNotFound},
		"groupless GET list":        {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/foo", http.StatusNotFound},
		"groupless POST list":       {"POST", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/foo", http.StatusNotFound},
		"groupless PUT object":      {"PUT", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/foo/bar", http.StatusNotFound},
		"groupless DELETE object":   {"DELETE", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/foo/bar", http.StatusNotFound},
		"groupless watch list":      {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/watch/foo", http.StatusNotFound},
		"groupless watch object":    {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/watch/foo/bar", http.StatusNotFound},
		"groupless proxy object":    {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/proxy/foo/bar", http.StatusNotFound},
		"groupless redirect object": {"GET", "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/redirect/foo/bar", http.StatusNotFound},

		"GET object":      {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/foo/bar", http.StatusNotFound},
		"GET list":        {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/foo", http.StatusNotFound},
		"POST list":       {"POST", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/foo", http.StatusNotFound},
		"PUT object":      {"PUT", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/foo/bar", http.StatusNotFound},
		"DELETE object":   {"DELETE", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/foo/bar", http.StatusNotFound},
		"watch list":      {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/watch/foo", http.StatusNotFound},
		"watch object":    {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/watch/foo/bar", http.StatusNotFound},
		"proxy object":    {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/proxy/foo/bar", http.StatusNotFound},
		"redirect object": {"GET", "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/redirect/foo/bar", http.StatusNotFound},
	}
	handler := handle(map[string]rest.Storage{
		"foo": UnimplementedRESTStorage{},
	})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}
	for k, v := range cases {
		request, err := http.NewRequest(v.Method, server.URL+v.Path, bytes.NewReader([]byte(`{"kind":"Simple","apiVersion":"version"}`)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		response, err := client.Do(request)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.StatusCode != v.ErrCode {
			t.Errorf("%s: expected %d for %s, Got %s", k, v.ErrCode, v.Method, string(data))
			continue
		}
	}
}

func TestVersion(t *testing.T) {
	handler := handle(map[string]rest.Storage{})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	request, err := http.NewRequest("GET", server.URL+"/version", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var info version.Info
	err = json.NewDecoder(response.Body).Decode(&info)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(version.Get(), info) {
		t.Errorf("Expected %#v, Got %#v", version.Get(), info)
	}
}

func TestList(t *testing.T) {
	testCases := []struct {
		url       string
		namespace string
		selfLink  string
		legacy    bool
		label     string
		field     string
	}{
		// Groupless API

		// legacy namespace param is ignored
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple?namespace=",
			namespace: "",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			legacy:    true,
		},
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple?namespace=other",
			namespace: "",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			legacy:    true,
		},
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple?namespace=other&labelSelector=a%3Db&fieldSelector=c%3Dd",
			namespace: "",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			legacy:    true,
			label:     "a=b",
			field:     "c=d",
		},
		// legacy api version is honored
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			namespace: "",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			legacy:    true,
		},
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple",
			namespace: "other",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple",
			legacy:    true,
		},
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple?labelSelector=a%3Db&fieldSelector=c%3Dd",
			namespace: "other",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple",
			legacy:    true,
			label:     "a=b",
			field:     "c=d",
		},
		// list items across all namespaces
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			namespace: "",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			legacy:    true,
		},
		// list items in a namespace in the path
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/default/simple",
			namespace: "default",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/default/simple",
		},
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple",
			namespace: "other",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple",
		},
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple?labelSelector=a%3Db&fieldSelector=c%3Dd",
			namespace: "other",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/namespaces/other/simple",
			label:     "a=b",
			field:     "c=d",
		},
		// list items across all namespaces
		{
			url:       "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
			namespace: "",
			selfLink:  "/" + grouplessPrefix + "/" + grouplessGroupVersion.Version + "/simple",
		},

		// Group API

		// legacy namespace param is ignored
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple?namespace=",
			namespace: "",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			legacy:    true,
		},
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple?namespace=other",
			namespace: "",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			legacy:    true,
		},
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple?namespace=other&labelSelector=a%3Db&fieldSelector=c%3Dd",
			namespace: "",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			legacy:    true,
			label:     "a=b",
			field:     "c=d",
		},
		// legacy api version is honored
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			namespace: "",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			legacy:    true,
		},
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/simple",
			namespace: "other",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/simple",
			legacy:    true,
		},
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/simple?labelSelector=a%3Db&fieldSelector=c%3Dd",
			namespace: "other",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/simple",
			legacy:    true,
			label:     "a=b",
			field:     "c=d",
		},
		// list items across all namespaces
		{
			url:       "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			namespace: "",
			selfLink:  "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple",
			legacy:    true,
		},
		// list items in a namespace in the path
		{
			url:       "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/default/simple",
			namespace: "default",
			selfLink:  "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/default/simple",
		},
		{
			url:       "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/other/simple",
			namespace: "other",
			selfLink:  "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/other/simple",
		},
		{
			url:       "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/other/simple?labelSelector=a%3Db&fieldSelector=c%3Dd",
			namespace: "other",
			selfLink:  "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/other/simple",
			label:     "a=b",
			field:     "c=d",
		},
		// list items across all namespaces
		{
			url:       "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/simple",
			namespace: "",
			selfLink:  "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/simple",
		},
	}
	for i, testCase := range testCases {
		storage := map[string]rest.Storage{}
		simpleStorage := SimpleRESTStorage{expectedResourceNamespace: testCase.namespace}
		storage["simple"] = &simpleStorage
		selfLinker := &setTestSelfLinker{
			t:           t,
			namespace:   testCase.namespace,
			expectedSet: testCase.selfLink,
		}
		var handler = handleInternal(storage, admissionControl, selfLinker)
		server := httptest.NewServer(handler)
		// TODO: Uncomment when fix #19254
		// defer server.Close()

		resp, err := http.Get(server.URL + testCase.url)
		if err != nil {
			t.Errorf("%d: unexpected error: %v", i, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%d: unexpected status: %d from url %s, Expected: %d, %#v", i, resp.StatusCode, testCase.url, http.StatusOK, resp)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("%d: unexpected error: %v", i, err)
				continue
			}
			t.Logf("%d: body: %s", i, string(body))
			continue
		}
		// TODO: future, restore get links
		if !selfLinker.called {
			t.Errorf("%d: never set self link", i)
		}
		if !simpleStorage.namespacePresent {
			t.Errorf("%d: namespace not set", i)
		} else if simpleStorage.actualNamespace != testCase.namespace {
			t.Errorf("%d: unexpected resource namespace: %s", i, simpleStorage.actualNamespace)
		}
		if simpleStorage.requestedLabelSelector == nil || simpleStorage.requestedLabelSelector.String() != testCase.label {
			t.Errorf("%d: unexpected label selector: %v", i, simpleStorage.requestedLabelSelector)
		}
		if simpleStorage.requestedFieldSelector == nil || simpleStorage.requestedFieldSelector.String() != testCase.field {
			t.Errorf("%d: unexpected field selector: %v", i, simpleStorage.requestedFieldSelector)
		}
	}
}

func TestErrorList(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		errors: map[string]error{"list": fmt.Errorf("test Error")},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusInternalServerError, resp)
	}
}

func TestNonEmptyList(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		list: []apiservertesting.Simple{
			{
				ObjectMeta: api.ObjectMeta{Name: "something", Namespace: "other"},
				Other:      "foo",
			},
		},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusOK, resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("Data: %s", string(body))
	}

	var listOut apiservertesting.SimpleList
	body, err := extractBody(resp, &listOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(listOut.Items) != 1 {
		t.Errorf("Unexpected response: %#v", listOut)
		return
	}
	if listOut.Items[0].Other != simpleStorage.list[0].Other {
		t.Errorf("Unexpected data: %#v, %s", listOut.Items[0], string(body))
	}
	if listOut.SelfLink != "/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/simple" {
		t.Errorf("unexpected list self link: %#v", listOut)
	}
	expectedSelfLink := "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/simple/something"
	if listOut.Items[0].ObjectMeta.SelfLink != expectedSelfLink {
		t.Errorf("Unexpected data: %#v, %s", listOut.Items[0].ObjectMeta.SelfLink, expectedSelfLink)
	}
}

func TestSelfLinkSkipsEmptyName(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		list: []apiservertesting.Simple{
			{
				ObjectMeta: api.ObjectMeta{Namespace: "other"},
				Other:      "foo",
			},
		},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusOK, resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("Data: %s", string(body))
	}
	var listOut apiservertesting.SimpleList
	body, err := extractBody(resp, &listOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(listOut.Items) != 1 {
		t.Errorf("Unexpected response: %#v", listOut)
		return
	}
	if listOut.Items[0].Other != simpleStorage.list[0].Other {
		t.Errorf("Unexpected data: %#v, %s", listOut.Items[0], string(body))
	}
	if listOut.SelfLink != "/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/simple" {
		t.Errorf("unexpected list self link: %#v", listOut)
	}
	expectedSelfLink := ""
	if listOut.Items[0].ObjectMeta.SelfLink != expectedSelfLink {
		t.Errorf("Unexpected data: %#v, %s", listOut.Items[0].ObjectMeta.SelfLink, expectedSelfLink)
	}
}

func TestMetadata(t *testing.T) {
	simpleStorage := &MetadataRESTStorage{&SimpleRESTStorage{}, []string{"text/plain"}}
	h := handle(map[string]rest.Storage{"simple": simpleStorage})
	ws := h.(*defaultAPIServer).container.RegisteredWebServices()
	if len(ws) == 0 {
		t.Fatal("no web services registered")
	}
	matches := map[string]int{}
	for _, w := range ws {
		for _, r := range w.Routes() {
			s := strings.Join(r.Produces, ",")
			i := matches[s]
			matches[s] = i + 1
		}
	}
	if matches["text/plain,application/json,application/yaml"] == 0 ||
		matches["application/json,application/yaml"] == 0 ||
		matches["application/json"] == 0 ||
		matches["*/*"] == 0 ||
		len(matches) != 4 {
		t.Errorf("unexpected mime types: %v", matches)
	}
}

func TestExport(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		item: apiservertesting.Simple{
			ObjectMeta: api.ObjectMeta{
				ResourceVersion:   "1234",
				CreationTimestamp: unversioned.NewTime(time.Unix(10, 10)),
			},
			Other: "foo",
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/id",
		name:        "id",
		namespace:   "default",
	}
	storage["simple"] = &simpleStorage
	handler := handleLinker(storage, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/id?export=true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("unexpected response: %#v\n%s\n", resp, string(data))
	}
	var itemOut apiservertesting.Simple
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if itemOut.Name != simpleStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simpleStorage.item, string(body))
	}
	if itemOut.Other != "exported" {
		t.Errorf("Expected: exported, saw: %s", itemOut.Other)
	}

	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestGet(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		item: apiservertesting.Simple{
			Other: "foo",
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/id",
		name:        "id",
		namespace:   "default",
	}
	storage["simple"] = &simpleStorage
	handler := handleLinker(storage, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apiservertesting.Simple
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if itemOut.Name != simpleStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simpleStorage.item, string(body))
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestGetBinary(t *testing.T) {
	simpleStorage := SimpleRESTStorage{
		stream: &SimpleStream{
			contentType: "text/plain",
			Reader:      bytes.NewBufferString("response data"),
		},
	}
	stream := simpleStorage.stream
	server := httptest.NewServer(handle(map[string]rest.Storage{"simple": &simpleStorage}))
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/binary", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req.Header.Add("Accept", "text/other, */*")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !stream.closed || stream.version != testGroupVersion.String() || stream.accept != "text/other, */*" ||
		resp.Header.Get("Content-Type") != stream.contentType || string(body) != "response data" {
		t.Errorf("unexpected stream: %#v", stream)
	}
}

func validateSimpleGetOptionsParams(t *testing.T, route *restful.Route) {
	// Validate name and description
	expectedParams := map[string]string{
		"param1":  "description for param1",
		"param2":  "description for param2",
		"atAPath": "",
	}
	for _, p := range route.ParameterDocs {
		data := p.Data()
		if desc, exists := expectedParams[data.Name]; exists {
			if desc != data.Description {
				t.Errorf("unexpected description for parameter %s: %s\n", data.Name, data.Description)
			}
			delete(expectedParams, data.Name)
		}
	}
	if len(expectedParams) > 0 {
		t.Errorf("did not find all expected parameters: %#v", expectedParams)
	}
}

func TestGetWithOptionsRouteParams(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := GetWithOptionsRESTStorage{
		SimpleRESTStorage: &SimpleRESTStorage{},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	ws := handler.(*defaultAPIServer).container.RegisteredWebServices()
	if len(ws) == 0 {
		t.Fatal("no web services registered")
	}
	routes := ws[0].Routes()
	for i := range routes {
		if routes[i].Method == "GET" && routes[i].Operation == "readNamespacedSimple" {
			validateSimpleGetOptionsParams(t, &routes[i])
			break
		}
	}
}

func TestGetWithOptions(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := GetWithOptionsRESTStorage{
		SimpleRESTStorage: &SimpleRESTStorage{
			item: apiservertesting.Simple{
				Other: "foo",
			},
		},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/id?param1=test1&param2=test2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apiservertesting.Simple
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if itemOut.Name != simpleStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simpleStorage.item, string(body))
	}

	opts, ok := simpleStorage.optionsReceived.(*apiservertesting.SimpleGetOptions)
	if !ok {
		t.Errorf("Unexpected options object received: %#v", simpleStorage.optionsReceived)
		return
	}
	if opts.Param1 != "test1" || opts.Param2 != "test2" {
		t.Errorf("Did not receive expected options: %#v", opts)
	}
}

func TestGetWithOptionsAndPath(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := GetWithOptionsRESTStorage{
		SimpleRESTStorage: &SimpleRESTStorage{
			item: apiservertesting.Simple{
				Other: "foo",
			},
		},
		takesPath: "atAPath",
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/id/a/different/path?param1=test1&param2=test2&atAPath=not")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apiservertesting.Simple
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if itemOut.Name != simpleStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simpleStorage.item, string(body))
	}

	opts, ok := simpleStorage.optionsReceived.(*apiservertesting.SimpleGetOptions)
	if !ok {
		t.Errorf("Unexpected options object received: %#v", simpleStorage.optionsReceived)
		return
	}
	if opts.Param1 != "test1" || opts.Param2 != "test2" || opts.Path != "a/different/path" {
		t.Errorf("Did not receive expected options: %#v", opts)
	}
}
func TestGetAlternateSelfLink(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		item: apiservertesting.Simple{
			Other: "foo",
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/test/simple/id",
		name:        "id",
		namespace:   "test",
	}
	storage["simple"] = &simpleStorage
	handler := handleLinker(storage, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/test/simple/id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apiservertesting.Simple
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if itemOut.Name != simpleStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simpleStorage.item, string(body))
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestGetNamespaceSelfLink(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		item: apiservertesting.Simple{
			Other: "foo",
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/foo/simple/id",
		name:        "id",
		namespace:   "foo",
	}
	storage["simple"] = &simpleStorage
	handler := handleInternal(storage, admissionControl, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/foo/simple/id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apiservertesting.Simple
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if itemOut.Name != simpleStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simpleStorage.item, string(body))
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}
func TestGetMissing(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{
		errors: map[string]error{"get": apierrs.NewNotFound(api.Resource("simples"), "id")},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simple/id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected response %#v", resp)
	}
}

func TestConnect(t *testing.T) {
	responseText := "Hello World"
	itemID := "theID"
	connectStorage := &ConnecterRESTStorage{
		connectHandler: &OutputConnect{
			response: responseText,
		},
	}
	storage := map[string]rest.Storage{
		"simple":         &SimpleRESTStorage{},
		"simple/connect": connectStorage,
	}
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + itemID + "/connect")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if connectStorage.receivedID != itemID {
		t.Errorf("Unexpected item id. Expected: %s. Actual: %s.", itemID, connectStorage.receivedID)
	}
	if string(body) != responseText {
		t.Errorf("Unexpected response. Expected: %s. Actual: %s.", responseText, string(body))
	}
}

func TestConnectResponderObject(t *testing.T) {
	itemID := "theID"
	simple := &apiservertesting.Simple{Other: "foo"}
	connectStorage := &ConnecterRESTStorage{}
	connectStorage.handlerFunc = func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			connectStorage.receivedResponder.Object(http.StatusCreated, simple)
		})
	}
	storage := map[string]rest.Storage{
		"simple":         &SimpleRESTStorage{},
		"simple/connect": connectStorage,
	}
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + itemID + "/connect")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("unexpected response: %#v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if connectStorage.receivedID != itemID {
		t.Errorf("Unexpected item id. Expected: %s. Actual: %s.", itemID, connectStorage.receivedID)
	}
	obj, err := runtime.Decode(codec, body)
	if err != nil {
		t.Fatal(err)
	}
	if !api.Semantic.DeepEqual(obj, simple) {
		t.Errorf("Unexpected response: %#v", obj)
	}
}

func TestConnectResponderError(t *testing.T) {
	itemID := "theID"
	connectStorage := &ConnecterRESTStorage{}
	connectStorage.handlerFunc = func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			connectStorage.receivedResponder.Error(apierrs.NewForbidden(api.Resource("simples"), itemID, errors.New("you are terminated")))
		})
	}
	storage := map[string]rest.Storage{
		"simple":         &SimpleRESTStorage{},
		"simple/connect": connectStorage,
	}
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + itemID + "/connect")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("unexpected response: %#v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if connectStorage.receivedID != itemID {
		t.Errorf("Unexpected item id. Expected: %s. Actual: %s.", itemID, connectStorage.receivedID)
	}
	obj, err := runtime.Decode(codec, body)
	if err != nil {
		t.Fatal(err)
	}
	if obj.(*unversioned.Status).Code != http.StatusForbidden {
		t.Errorf("Unexpected response: %#v", obj)
	}
}

func TestConnectWithOptionsRouteParams(t *testing.T) {
	connectStorage := &ConnecterRESTStorage{
		connectHandler:      &OutputConnect{},
		emptyConnectOptions: &apiservertesting.SimpleGetOptions{},
	}
	storage := map[string]rest.Storage{
		"simple":         &SimpleRESTStorage{},
		"simple/connect": connectStorage,
	}
	handler := handle(storage)
	ws := handler.(*defaultAPIServer).container.RegisteredWebServices()
	if len(ws) == 0 {
		t.Fatal("no web services registered")
	}
	routes := ws[0].Routes()
	for i := range routes {
		switch routes[i].Operation {
		case "connectGetNamespacedSimpleConnect":
		case "connectPostNamespacedSimpleConnect":
		case "connectPutNamespacedSimpleConnect":
		case "connectDeleteNamespacedSimpleConnect":
			validateSimpleGetOptionsParams(t, &routes[i])

		}
	}
}

func TestConnectWithOptions(t *testing.T) {
	responseText := "Hello World"
	itemID := "theID"
	connectStorage := &ConnecterRESTStorage{
		connectHandler: &OutputConnect{
			response: responseText,
		},
		emptyConnectOptions: &apiservertesting.SimpleGetOptions{},
	}
	storage := map[string]rest.Storage{
		"simple":         &SimpleRESTStorage{},
		"simple/connect": connectStorage,
	}
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + itemID + "/connect?param1=value1&param2=value2")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if connectStorage.receivedID != itemID {
		t.Errorf("Unexpected item id. Expected: %s. Actual: %s.", itemID, connectStorage.receivedID)
	}
	if string(body) != responseText {
		t.Errorf("Unexpected response. Expected: %s. Actual: %s.", responseText, string(body))
	}
	if connectStorage.receivedResponder == nil {
		t.Errorf("Unexpected responder")
	}
	opts, ok := connectStorage.receivedConnectOptions.(*apiservertesting.SimpleGetOptions)
	if !ok {
		t.Fatalf("Unexpected options type: %#v", connectStorage.receivedConnectOptions)
	}
	if opts.Param1 != "value1" && opts.Param2 != "value2" {
		t.Errorf("Unexpected options value: %#v", opts)
	}
}

func TestConnectWithOptionsAndPath(t *testing.T) {
	responseText := "Hello World"
	itemID := "theID"
	testPath := "a/b/c/def"
	connectStorage := &ConnecterRESTStorage{
		connectHandler: &OutputConnect{
			response: responseText,
		},
		emptyConnectOptions: &apiservertesting.SimpleGetOptions{},
		takesPath:           "atAPath",
	}
	storage := map[string]rest.Storage{
		"simple":         &SimpleRESTStorage{},
		"simple/connect": connectStorage,
	}
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + itemID + "/connect/" + testPath + "?param1=value1&param2=value2")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if connectStorage.receivedID != itemID {
		t.Errorf("Unexpected item id. Expected: %s. Actual: %s.", itemID, connectStorage.receivedID)
	}
	if string(body) != responseText {
		t.Errorf("Unexpected response. Expected: %s. Actual: %s.", responseText, string(body))
	}
	opts, ok := connectStorage.receivedConnectOptions.(*apiservertesting.SimpleGetOptions)
	if !ok {
		t.Fatalf("Unexpected options type: %#v", connectStorage.receivedConnectOptions)
	}
	if opts.Param1 != "value1" && opts.Param2 != "value2" {
		t.Errorf("Unexpected options value: %#v", opts)
	}
	if opts.Path != testPath {
		t.Errorf("Unexpected path value. Expected: %s. Actual: %s.", testPath, opts.Path)
	}
}

func TestDelete(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	client := http.Client{}
	request, err := http.NewRequest("DELETE", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, nil)
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
	if simpleStorage.deleted != ID {
		t.Errorf("Unexpected delete: %s, expected %s", simpleStorage.deleted, ID)
	}
}

func TestDeleteWithOptions(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	grace := int64(300)
	item := &api.DeleteOptions{
		GracePeriodSeconds: &grace,
	}
	body, err := runtime.Encode(codec, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("DELETE", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %s %#v", request.URL, res)
		s, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf(string(s))
	}
	if simpleStorage.deleted != ID {
		t.Errorf("Unexpected delete: %s, expected %s", simpleStorage.deleted, ID)
	}
	if !api.Semantic.DeepEqual(simpleStorage.deleteOptions, item) {
		t.Errorf("unexpected delete options: %s", diff.ObjectDiff(simpleStorage.deleteOptions, item))
	}
}

func TestLegacyDelete(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = LegacyRESTStorage{&simpleStorage}
	var _ rest.Deleter = storage["simple"].(LegacyRESTStorage)
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	client := http.Client{}
	request, err := http.NewRequest("DELETE", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, nil)
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
	if simpleStorage.deleted != ID {
		t.Errorf("Unexpected delete: %s, expected %s", simpleStorage.deleted, ID)
	}
	if simpleStorage.deleteOptions != nil {
		t.Errorf("unexpected delete options: %#v", simpleStorage.deleteOptions)
	}
}

func TestLegacyDeleteIgnoresOptions(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = LegacyRESTStorage{&simpleStorage}
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := api.NewDeleteOptions(300)
	body, err := runtime.Encode(codec, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("DELETE", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
	if simpleStorage.deleted != ID {
		t.Errorf("Unexpected delete: %s, expected %s", simpleStorage.deleted, ID)
	}
	if simpleStorage.deleteOptions != nil {
		t.Errorf("unexpected delete options: %#v", simpleStorage.deleteOptions)
	}
}

func TestDeleteInvokesAdmissionControl(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handleDeny(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	client := http.Client{}
	request, err := http.NewRequest("DELETE", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestDeleteMissing(t *testing.T) {
	storage := map[string]rest.Storage{}
	ID := "id"
	simpleStorage := SimpleRESTStorage{
		errors: map[string]error{"delete": apierrs.NewNotFound(api.Resource("simples"), ID)},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	client := http.Client{}
	request, err := http.NewRequest("DELETE", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestPatch(t *testing.T) {
	storage := map[string]rest.Storage{}
	ID := "id"
	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: "", // update should allow the client to send an empty namespace
		},
		Other: "bar",
	}
	simpleStorage := SimpleRESTStorage{item: *item}
	storage["simple"] = &simpleStorage
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + ID,
		name:        ID,
		namespace:   api.NamespaceDefault,
	}
	handler := handleLinker(storage, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	client := http.Client{}
	request, err := http.NewRequest("PATCH", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader([]byte(`{"labels":{"foo":"bar"}}`)))
	request.Header.Set("Content-Type", "application/merge-patch+json; charset=UTF-8")
	_, err = client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if simpleStorage.updated == nil || simpleStorage.updated.Labels["foo"] != "bar" {
		t.Errorf("Unexpected update value %#v, expected %#v.", simpleStorage.updated, item)
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestPatchRequiresMatchingName(t *testing.T) {
	storage := map[string]rest.Storage{}
	ID := "id"
	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: "", // update should allow the client to send an empty namespace
		},
		Other: "bar",
	}
	simpleStorage := SimpleRESTStorage{item: *item}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	client := http.Client{}
	request, err := http.NewRequest("PATCH", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader([]byte(`{"metadata":{"name":"idbar"}}`)))
	request.Header.Set("Content-Type", "application/merge-patch+json")
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestUpdate(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	selfLinker := &setTestSelfLinker{
		t:           t,
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + ID,
		name:        ID,
		namespace:   api.NamespaceDefault,
	}
	handler := handleLinker(storage, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: "", // update should allow the client to send an empty namespace
		},
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		// The following cases will fail, so die now
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	_, err = client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if simpleStorage.updated == nil || simpleStorage.updated.Name != item.Name {
		t.Errorf("Unexpected update value %#v, expected %#v.", simpleStorage.updated, item)
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestUpdateInvokesAdmissionControl(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handleDeny(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: api.NamespaceDefault,
		},
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		// The following cases will fail, so die now
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestUpdateRequiresMatchingName(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handleDeny(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		// The following cases will fail, so die now
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestUpdateAllowsMissingNamespace(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name: ID,
		},
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		// The following cases will fail, so die now
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response %#v", response)
	}
}

// when the object name and namespace can't be retrieved, skip name checking
func TestUpdateAllowsMismatchedNamespaceOnError(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	selfLinker := &setTestSelfLinker{
		t:   t,
		err: fmt.Errorf("test error"),
	}
	handler := handleLinker(storage, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: "other", // does not match request
		},
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		// The following cases will fail, so die now
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	_, err = client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if simpleStorage.updated == nil || simpleStorage.updated.Name != item.Name {
		t.Errorf("Unexpected update value %#v, expected %#v.", simpleStorage.updated, item)
	}
	if selfLinker.called {
		t.Errorf("self link ignored")
	}
}

func TestUpdatePreventsMismatchedNamespace(t *testing.T) {
	storage := map[string]rest.Storage{}
	simpleStorage := SimpleRESTStorage{}
	ID := "id"
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: "other",
		},
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		// The following cases will fail, so die now
		t.Fatalf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestUpdateMissing(t *testing.T) {
	storage := map[string]rest.Storage{}
	ID := "id"
	simpleStorage := SimpleRESTStorage{
		errors: map[string]error{"update": apierrs.NewNotFound(api.Resource("simples"), ID)},
	}
	storage["simple"] = &simpleStorage
	handler := handle(storage)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	item := &apiservertesting.Simple{
		ObjectMeta: api.ObjectMeta{
			Name:      ID,
			Namespace: api.NamespaceDefault,
		},
		Other: "bar",
	}
	body, err := runtime.Encode(testCodec, item)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	client := http.Client{}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+ID, bytes.NewReader(body))
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestCreateNotFound(t *testing.T) {
	handler := handle(map[string]rest.Storage{
		"simple": &SimpleRESTStorage{
			// storage.Create can fail with not found error in theory.
			// See http://pr.k8s.io/486#discussion_r15037092.
			errors: map[string]error{"create": apierrs.NewNotFound(api.Resource("simples"), "id")},
		},
	})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{Other: "foo"}
	data, err := runtime.Encode(testCodec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected response %#v", response)
	}
}

func TestCreateChecksDecode(t *testing.T) {
	handler := handle(map[string]rest.Storage{"simple": &SimpleRESTStorage{}})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &api.Pod{}
	data, err := runtime.Encode(codec, simple, testGroupVersion)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v", response)
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !strings.Contains(string(b), "cannot be handled as a Simple") {
		t.Errorf("unexpected response: %s", string(b))
	}
}

// TestUpdateREST tests that you can add new rest implementations to a pre-existing
// web service.
func TestUpdateREST(t *testing.T) {
	makeGroup := func(storage map[string]rest.Storage) *APIGroupVersion {
		return &APIGroupVersion{
			Storage:             storage,
			Root:                "/" + prefix,
			RequestInfoResolver: newTestRequestInfoResolver(),
			Creater:             api.Scheme,
			Convertor:           api.Scheme,
			Typer:               api.Scheme,
			Linker:              selfLinker,

			Admit:   admissionControl,
			Context: requestContextMapper,
			Mapper:  namespaceMapper,

			GroupVersion:           newGroupVersion,
			OptionsExternalVersion: &newGroupVersion,

			Serializer:       api.Codecs,
			StreamSerializer: api.StreamCodecs,
			ParameterCodec:   api.ParameterCodec,
		}
	}

	makeStorage := func(paths ...string) map[string]rest.Storage {
		storage := map[string]rest.Storage{}
		for _, s := range paths {
			storage[s] = &SimpleRESTStorage{}
		}
		return storage
	}

	testREST := func(t *testing.T, container *restful.Container, barCode int) {
		w := httptest.NewRecorder()
		container.ServeHTTP(w, &http.Request{Method: "GET", URL: &url.URL{Path: "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/test/foo/test"}})
		if w.Code != http.StatusOK {
			t.Fatalf("expected OK: %#v", w)
		}

		w = httptest.NewRecorder()
		container.ServeHTTP(w, &http.Request{Method: "GET", URL: &url.URL{Path: "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/test/bar/test"}})
		if w.Code != barCode {
			t.Errorf("expected response code %d for GET to bar but received %d", barCode, w.Code)
		}
	}

	storage1 := makeStorage("foo")
	group1 := makeGroup(storage1)

	storage2 := makeStorage("bar")
	group2 := makeGroup(storage2)

	container := restful.NewContainer()

	// install group1.  Ensure that
	// 1. Foo storage is accessible
	// 2. Bar storage is not accessible
	if err := group1.InstallREST(container); err != nil {
		t.Fatal(err)
	}
	testREST(t, container, http.StatusNotFound)

	// update with group2.  Ensure that
	// 1.  Foo storage is still accessible
	// 2.  Bar storage is now accessible
	if err := group2.UpdateREST(container); err != nil {
		t.Fatal(err)
	}
	testREST(t, container, http.StatusOK)

	// try to update a group that does not have an existing webservice with a matching prefix
	// should not affect the existing registered webservice
	invalidGroup := makeGroup(storage1)
	invalidGroup.Root = "bad"
	if err := invalidGroup.UpdateREST(container); err == nil {
		t.Fatal("expected an error from UpdateREST when updating a non-existing prefix but got none")
	}
	testREST(t, container, http.StatusOK)
}

func TestParentResourceIsRequired(t *testing.T) {
	storage := &SimpleTypedStorage{
		baseType: &apiservertesting.SimpleRoot{}, // a root scoped type
		item:     &apiservertesting.SimpleRoot{},
	}
	group := &APIGroupVersion{
		Storage: map[string]rest.Storage{
			"simple/sub": storage,
		},
		Root:                "/" + prefix,
		RequestInfoResolver: newTestRequestInfoResolver(),
		Creater:             api.Scheme,
		Convertor:           api.Scheme,
		Typer:               api.Scheme,
		Linker:              selfLinker,

		Admit:   admissionControl,
		Context: requestContextMapper,
		Mapper:  namespaceMapper,

		GroupVersion:           newGroupVersion,
		OptionsExternalVersion: &newGroupVersion,

		Serializer:       api.Codecs,
		StreamSerializer: api.StreamCodecs,
		ParameterCodec:   api.ParameterCodec,
	}
	container := restful.NewContainer()
	if err := group.InstallREST(container); err == nil {
		t.Fatal("expected error")
	}

	storage = &SimpleTypedStorage{
		baseType: &apiservertesting.SimpleRoot{}, // a root scoped type
		item:     &apiservertesting.SimpleRoot{},
	}
	group = &APIGroupVersion{
		Storage: map[string]rest.Storage{
			"simple":     &SimpleRESTStorage{},
			"simple/sub": storage,
		},
		Root:                "/" + prefix,
		RequestInfoResolver: newTestRequestInfoResolver(),
		Creater:             api.Scheme,
		Convertor:           api.Scheme,
		Typer:               api.Scheme,
		Linker:              selfLinker,

		Admit:   admissionControl,
		Context: requestContextMapper,
		Mapper:  namespaceMapper,

		GroupVersion:           newGroupVersion,
		OptionsExternalVersion: &newGroupVersion,

		Serializer:       api.Codecs,
		StreamSerializer: api.StreamCodecs,
		ParameterCodec:   api.ParameterCodec,
	}
	container = restful.NewContainer()
	if err := group.InstallREST(container); err != nil {
		t.Fatal(err)
	}

	// resource is NOT registered in the root scope
	w := httptest.NewRecorder()
	container.ServeHTTP(w, &http.Request{Method: "GET", URL: &url.URL{Path: "/" + prefix + "/simple/test/sub"}})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected not found: %#v", w)
	}

	// resource is registered in the namespace scope
	w = httptest.NewRecorder()
	container.ServeHTTP(w, &http.Request{Method: "GET", URL: &url.URL{Path: "/" + prefix + "/" + newGroupVersion.Group + "/" + newGroupVersion.Version + "/namespaces/test/simple/test/sub"}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected OK: %#v", w)
	}
	if storage.actualNamespace != "test" {
		t.Errorf("namespace should be set %#v", storage)
	}
}

func TestCreateWithName(t *testing.T) {
	pathName := "helloworld"
	storage := &NamedCreaterRESTStorage{SimpleRESTStorage: &SimpleRESTStorage{}}
	handler := handle(map[string]rest.Storage{
		"simple":     &SimpleRESTStorage{},
		"simple/sub": storage,
	})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{Other: "foo"}
	data, err := runtime.Encode(testCodec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/"+pathName+"/sub", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected response %#v", response)
	}
	if storage.createdName != pathName {
		t.Errorf("Did not get expected name in create context. Got: %s, Expected: %s", storage.createdName, pathName)
	}
}

func TestUpdateChecksDecode(t *testing.T) {
	handler := handle(map[string]rest.Storage{"simple": &SimpleRESTStorage{}})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &api.Pod{}
	data, err := runtime.Encode(codec, simple, testGroupVersion)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/bar", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v\n%s", response, readBodyOrDie(response.Body))
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !strings.Contains(string(b), "cannot be handled as a Simple") {
		t.Errorf("unexpected response: %s", string(b))
	}
}

func TestParseTimeout(t *testing.T) {
	if d := parseTimeout(""); d != 30*time.Second {
		t.Errorf("blank timeout produces %v", d)
	}
	if d := parseTimeout("not a timeout"); d != 30*time.Second {
		t.Errorf("bad timeout produces %v", d)
	}
	if d := parseTimeout("10s"); d != 10*time.Second {
		t.Errorf("10s timeout produced: %v", d)
	}
}

type setTestSelfLinker struct {
	t           *testing.T
	expectedSet string
	name        string
	namespace   string
	called      bool
	err         error
}

func (s *setTestSelfLinker) Namespace(runtime.Object) (string, error) { return s.namespace, s.err }
func (s *setTestSelfLinker) Name(runtime.Object) (string, error)      { return s.name, s.err }
func (s *setTestSelfLinker) SelfLink(runtime.Object) (string, error)  { return "", s.err }
func (s *setTestSelfLinker) SetSelfLink(obj runtime.Object, selfLink string) error {
	if e, a := s.expectedSet, selfLink; e != a {
		s.t.Errorf("expected '%v', got '%v'", e, a)
	}
	s.called = true
	return s.err
}

func TestCreate(t *testing.T) {
	storage := SimpleRESTStorage{
		injectedFunction: func(obj runtime.Object) (runtime.Object, error) {
			time.Sleep(5 * time.Millisecond)
			return obj, nil
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		name:        "bar",
		namespace:   "default",
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/foo/bar",
	}
	handler := handleLinker(map[string]rest.Storage{"foo": &storage}, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{
		Other: "bar",
	}
	data, err := runtime.Encode(testCodec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/foo", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var itemOut apiservertesting.Simple
	body, err := extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}

	if !reflect.DeepEqual(&itemOut, simple) {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simple, string(body))
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestCreateYAML(t *testing.T) {
	storage := SimpleRESTStorage{
		injectedFunction: func(obj runtime.Object) (runtime.Object, error) {
			time.Sleep(5 * time.Millisecond)
			return obj, nil
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		name:        "bar",
		namespace:   "default",
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/foo/bar",
	}
	handler := handleLinker(map[string]rest.Storage{"foo": &storage}, selfLinker)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	// yaml encoder
	simple := &apiservertesting.Simple{
		Other: "bar",
	}
	serializer, ok := api.Codecs.SerializerForMediaType("application/yaml", nil)
	if !ok {
		t.Fatal("No yaml serializer")
	}
	encoder := api.Codecs.EncoderForVersion(serializer, testGroupVersion)
	decoder := api.Codecs.DecoderToVersion(serializer, testInternalGroupVersion)

	data, err := runtime.Encode(encoder, simple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/foo", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	request.Header.Set("Accept", "application/yaml, application/json")
	request.Header.Set("Content-Type", "application/yaml")

	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var itemOut apiservertesting.Simple
	body, err := extractBodyDecoder(response, &itemOut, decoder)
	if err != nil {
		t.Fatalf("unexpected error: %v %#v", err, response)
	}

	if !reflect.DeepEqual(&itemOut, simple) {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simple, string(body))
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}
func TestCreateInNamespace(t *testing.T) {
	storage := SimpleRESTStorage{
		injectedFunction: func(obj runtime.Object) (runtime.Object, error) {
			time.Sleep(5 * time.Millisecond)
			return obj, nil
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		name:        "bar",
		namespace:   "other",
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/foo/bar",
	}
	handler := handleLinker(map[string]rest.Storage{"foo": &storage}, selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{
		Other: "bar",
	}
	data, err := runtime.Encode(testCodec, simple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/other/foo", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var itemOut apiservertesting.Simple
	body, err := extractBody(response, &itemOut)
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, data)
	}

	if !reflect.DeepEqual(&itemOut, simple) {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simple, string(body))
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	if !selfLinker.called {
		t.Errorf("Never set self link")
	}
}

func TestCreateInvokesAdmissionControl(t *testing.T) {
	storage := SimpleRESTStorage{
		injectedFunction: func(obj runtime.Object) (runtime.Object, error) {
			time.Sleep(5 * time.Millisecond)
			return obj, nil
		},
	}
	selfLinker := &setTestSelfLinker{
		t:           t,
		name:        "bar",
		namespace:   "other",
		expectedSet: "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/other/foo/bar",
	}
	handler := handleInternal(map[string]rest.Storage{"foo": &storage}, deny.NewAlwaysDeny(), selfLinker)
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{
		Other: "bar",
	}
	data, err := runtime.Encode(testCodec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/other/foo", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusForbidden, response)
	}
}

func expectApiStatus(t *testing.T, method, url string, data []byte, code int) *unversioned.Status {
	client := http.Client{}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
		return nil
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error on %s %s: %v", method, url, err)
		return nil
	}
	var status unversioned.Status
	if body, err := extractBody(response, &status); err != nil {
		t.Fatalf("unexpected error on %s %s: %v\nbody:\n%s", method, url, err, body)
		return nil
	}
	if code != response.StatusCode {
		t.Fatalf("Expected %s %s to return %d, Got %d", method, url, code, response.StatusCode)
	}
	return &status
}

func TestDelayReturnsError(t *testing.T) {
	storage := SimpleRESTStorage{
		injectedFunction: func(obj runtime.Object) (runtime.Object, error) {
			return nil, apierrs.NewAlreadyExists(api.Resource("foos"), "bar")
		},
	}
	handler := handle(map[string]rest.Storage{"foo": &storage})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	status := expectApiStatus(t, "DELETE", fmt.Sprintf("%s/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/foo/bar", server.URL), nil, http.StatusConflict)
	if status.Status != unversioned.StatusFailure || status.Message == "" || status.Details == nil || status.Reason != unversioned.StatusReasonAlreadyExists {
		t.Errorf("Unexpected status %#v", status)
	}
}

type UnregisteredAPIObject struct {
	Value string
}

func (obj *UnregisteredAPIObject) GetObjectKind() unversioned.ObjectKind {
	return unversioned.EmptyObjectKind
}

func TestWriteJSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeNegotiated(api.Codecs, newGroupVersion, w, req, http.StatusOK, &UnregisteredAPIObject{"Undecodable"})
	}))
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	// We send a 200 status code before we encode the object, so we expect OK, but there will
	// still be an error object.  This seems ok, the alternative is to validate the object before
	// encoding, but this really should never happen, so it's wasted compute for every API request.
	status := expectApiStatus(t, "GET", server.URL, nil, http.StatusOK)
	if status.Reason != unversioned.StatusReasonUnknown {
		t.Errorf("unexpected reason %#v", status)
	}
	if !strings.Contains(status.Message, "no kind is registered for the type apiserver.UnregisteredAPIObject") {
		t.Errorf("unexpected message %#v", status)
	}
}

type marshalError struct {
	err error
}

func (m *marshalError) MarshalJSON() ([]byte, error) {
	return []byte{}, m.err
}

func TestWriteRAWJSONMarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeRawJSON(http.StatusOK, &marshalError{errors.New("Undecodable")}, w)
	}))
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

func TestCreateTimeout(t *testing.T) {
	testOver := make(chan struct{})
	defer close(testOver)
	storage := SimpleRESTStorage{
		injectedFunction: func(obj runtime.Object) (runtime.Object, error) {
			// Eliminate flakes by ensuring the create operation takes longer than this test.
			<-testOver
			return obj, nil
		},
	}
	handler := handle(map[string]rest.Storage{
		"foo": &storage,
	})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	simple := &apiservertesting.Simple{Other: "foo"}
	data, err := runtime.Encode(testCodec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	itemOut := expectApiStatus(t, "POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/foo?timeout=4ms", data, apierrs.StatusServerTimeout)
	if itemOut.Status != unversioned.StatusFailure || itemOut.Reason != unversioned.StatusReasonTimeout {
		t.Errorf("Unexpected status %#v", itemOut)
	}
}

func TestCORSAllowedOrigins(t *testing.T) {
	table := []struct {
		allowedOrigins []string
		origin         string
		allowed        bool
	}{
		{[]string{}, "example.com", false},
		{[]string{"example.com"}, "example.com", true},
		{[]string{"example.com"}, "not-allowed.com", false},
		{[]string{"not-matching.com", "example.com"}, "example.com", true},
		{[]string{".*"}, "example.com", true},
	}

	for _, item := range table {
		allowedOriginRegexps, err := util.CompileRegexps(item.allowedOrigins)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		handler := CORS(
			handle(map[string]rest.Storage{}),
			allowedOriginRegexps, nil, nil, "true",
		)
		server := httptest.NewServer(handler)
		// TODO: Uncomment when fix #19254
		// defer server.Close()
		client := http.Client{}

		request, err := http.NewRequest("GET", server.URL+"/version", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		request.Header.Set("Origin", item.origin)

		response, err := client.Do(request)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if item.allowed {
			if !reflect.DeepEqual(item.origin, response.Header.Get("Access-Control-Allow-Origin")) {
				t.Errorf("Expected %#v, Got %#v", item.origin, response.Header.Get("Access-Control-Allow-Origin"))
			}

			if response.Header.Get("Access-Control-Allow-Credentials") == "" {
				t.Errorf("Expected Access-Control-Allow-Credentials header to be set")
			}

			if response.Header.Get("Access-Control-Allow-Headers") == "" {
				t.Errorf("Expected Access-Control-Allow-Headers header to be set")
			}

			if response.Header.Get("Access-Control-Allow-Methods") == "" {
				t.Errorf("Expected Access-Control-Allow-Methods header to be set")
			}
		} else {
			if response.Header.Get("Access-Control-Allow-Origin") != "" {
				t.Errorf("Expected Access-Control-Allow-Origin header to not be set")
			}

			if response.Header.Get("Access-Control-Allow-Credentials") != "" {
				t.Errorf("Expected Access-Control-Allow-Credentials header to not be set")
			}

			if response.Header.Get("Access-Control-Allow-Headers") != "" {
				t.Errorf("Expected Access-Control-Allow-Headers header to not be set")
			}

			if response.Header.Get("Access-Control-Allow-Methods") != "" {
				t.Errorf("Expected Access-Control-Allow-Methods header to not be set")
			}
		}
	}
}

func TestCreateChecksAPIVersion(t *testing.T) {
	handler := handle(map[string]rest.Storage{"simple": &SimpleRESTStorage{}})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{}
	//using newCodec and send the request to testVersion URL shall cause a discrepancy in apiVersion
	data, err := runtime.Encode(newCodec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v", response)
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !strings.Contains(string(b), "does not match the expected API version") {
		t.Errorf("unexpected response: %s", string(b))
	}
}

func TestCreateDefaultsAPIVersion(t *testing.T) {
	handler := handle(map[string]rest.Storage{"simple": &SimpleRESTStorage{}})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	delete(m, "apiVersion")
	data, err = json.Marshal(m)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	request, err := http.NewRequest("POST", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple", bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
}

func TestUpdateChecksAPIVersion(t *testing.T) {
	handler := handle(map[string]rest.Storage{"simple": &SimpleRESTStorage{}})
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()
	client := http.Client{}

	simple := &apiservertesting.Simple{ObjectMeta: api.ObjectMeta{Name: "bar"}}
	data, err := runtime.Encode(newCodec, simple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("PUT", server.URL+"/"+prefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/default/simple/bar", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected response %#v", response)
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !strings.Contains(string(b), "does not match the expected API version") {
		t.Errorf("unexpected response: %s", string(b))
	}
}

// SimpleXGSubresource is a cross group subresource, i.e. the subresource does not belong to the
// same group as its parent resource.
type SimpleXGSubresource struct {
	unversioned.TypeMeta `json:",inline"`
	api.ObjectMeta       `json:"metadata"`
	SubresourceInfo      string            `json:"subresourceInfo,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
}

func (obj *SimpleXGSubresource) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }

type SimpleXGSubresourceRESTStorage struct {
	item SimpleXGSubresource
}

func (storage *SimpleXGSubresourceRESTStorage) New() runtime.Object {
	return &SimpleXGSubresource{}
}

func (storage *SimpleXGSubresourceRESTStorage) Get(ctx api.Context, id string) (runtime.Object, error) {
	copied, err := api.Scheme.Copy(&storage.item)
	if err != nil {
		panic(err)
	}
	return copied, nil
}

func TestXGSubresource(t *testing.T) {
	container := restful.NewContainer()
	container.Router(restful.CurlyRouter{})
	mux := container.ServeMux

	itemID := "theID"
	subresourceStorage := &SimpleXGSubresourceRESTStorage{
		item: SimpleXGSubresource{
			SubresourceInfo: "foo",
		},
	}
	storage := map[string]rest.Storage{
		"simple":           &SimpleRESTStorage{},
		"simple/subsimple": subresourceStorage,
	}

	group := APIGroupVersion{
		Storage: storage,

		RequestInfoResolver: newTestRequestInfoResolver(),

		Creater:   api.Scheme,
		Convertor: api.Scheme,
		Typer:     api.Scheme,
		Linker:    selfLinker,
		Mapper:    namespaceMapper,

		ParameterCodec: api.ParameterCodec,

		Admit:   admissionControl,
		Context: requestContextMapper,

		Root:                   "/" + prefix,
		GroupVersion:           testGroupVersion,
		OptionsExternalVersion: &testGroupVersion,
		Serializer:             api.Codecs,
		StreamSerializer:       api.StreamCodecs,

		SubresourceGroupVersionKind: map[string]unversioned.GroupVersionKind{
			"simple/subsimple": testGroup2Version.WithKind("SimpleXGSubresource"),
		},
	}

	if err := (&group).InstallREST(container); err != nil {
		panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
	}

	ws := new(restful.WebService)
	InstallSupport(mux, ws)
	container.Add(ws)

	handler := defaultAPIServer{mux, container}
	server := httptest.NewServer(handler)
	// TODO: Uncomment when fix #19254
	// defer server.Close()

	resp, err := http.Get(server.URL + "/" + prefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/default/simple/" + itemID + "/subsimple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut SimpleXGSubresource
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test if the returned object has the expected group, version and kind
	// We are directly unmarshaling JSON here because TypeMeta cannot be decoded through the
	// installed decoders. TypeMeta cannot be decoded because it is added to the ignored
	// conversion type list in API scheme and hence cannot be converted from input type object
	// to output type object. So it's values don't appear in the decoded output object.
	decoder := json.NewDecoder(strings.NewReader(body))
	var itemFromBody SimpleXGSubresource
	err = decoder.Decode(&itemFromBody)
	if err != nil {
		t.Errorf("unexpected JSON decoding error: %v", err)
	}
	if want := fmt.Sprintf("%s/%s", testGroup2Version.Group, testGroup2Version.Version); itemFromBody.APIVersion != want {
		t.Errorf("unexpected APIVersion got: %+v want: %+v", itemFromBody.APIVersion, want)
	}
	if itemFromBody.Kind != "SimpleXGSubresource" {
		t.Errorf("unexpected Kind got: %+v want: SimpleXGSubresource", itemFromBody.Kind)
	}

	if itemOut.Name != subresourceStorage.item.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, subresourceStorage.item, string(body))
	}
}

func readBodyOrDie(r io.Reader) []byte {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return body
}
