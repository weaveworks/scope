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

package thirdpartyresourcedata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	apiutil "k8s.io/kubernetes/pkg/api/util"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/runtime"
)

type thirdPartyObjectConverter struct {
	converter runtime.ObjectConvertor
}

func (t *thirdPartyObjectConverter) ConvertToVersion(in runtime.Object, outVersion string) (out runtime.Object, err error) {
	switch in.(type) {
	// This seems weird, but in this case the ThirdPartyResourceData is really just a wrapper on the raw 3rd party data.
	// The actual thing printed/sent to server is the actual raw third party resource data, which only has one version.
	case *extensions.ThirdPartyResourceData:
		return in, nil
	default:
		return t.converter.ConvertToVersion(in, outVersion)
	}
}

func (t *thirdPartyObjectConverter) Convert(in, out interface{}) error {
	return t.converter.Convert(in, out)
}

func (t *thirdPartyObjectConverter) ConvertFieldLabel(version, kind, label, value string) (string, string, error) {
	return t.converter.ConvertFieldLabel(version, kind, label, value)
}

func NewThirdPartyObjectConverter(converter runtime.ObjectConvertor) runtime.ObjectConvertor {
	return &thirdPartyObjectConverter{converter}
}

type thirdPartyResourceDataMapper struct {
	mapper  meta.RESTMapper
	kind    string
	version string
	group   string
}

var _ meta.RESTMapper = &thirdPartyResourceDataMapper{}

func (t *thirdPartyResourceDataMapper) getResource() unversioned.GroupVersionResource {
	plural, _ := meta.KindToResource(t.getKind())

	return plural
}

func (t *thirdPartyResourceDataMapper) getKind() unversioned.GroupVersionKind {
	return unversioned.GroupVersionKind{Group: t.group, Version: t.version, Kind: t.kind}
}

func (t *thirdPartyResourceDataMapper) isThirdPartyResource(partialResource unversioned.GroupVersionResource) bool {
	actualResource := t.getResource()
	if strings.ToLower(partialResource.Resource) != strings.ToLower(actualResource.Resource) {
		return false
	}
	if len(partialResource.Group) != 0 && partialResource.Group != actualResource.Group {
		return false
	}
	if len(partialResource.Version) != 0 && partialResource.Version != actualResource.Version {
		return false
	}

	return true
}

func (t *thirdPartyResourceDataMapper) ResourcesFor(resource unversioned.GroupVersionResource) ([]unversioned.GroupVersionResource, error) {
	if t.isThirdPartyResource(resource) {
		return []unversioned.GroupVersionResource{t.getResource()}, nil
	}
	return t.mapper.ResourcesFor(resource)
}

func (t *thirdPartyResourceDataMapper) KindsFor(resource unversioned.GroupVersionResource) ([]unversioned.GroupVersionKind, error) {
	if t.isThirdPartyResource(resource) {
		return []unversioned.GroupVersionKind{t.getKind()}, nil
	}
	return t.mapper.KindsFor(resource)
}

func (t *thirdPartyResourceDataMapper) ResourceFor(resource unversioned.GroupVersionResource) (unversioned.GroupVersionResource, error) {
	if t.isThirdPartyResource(resource) {
		return t.getResource(), nil
	}
	return t.mapper.ResourceFor(resource)
}

func (t *thirdPartyResourceDataMapper) KindFor(resource unversioned.GroupVersionResource) (unversioned.GroupVersionKind, error) {
	if t.isThirdPartyResource(resource) {
		return t.getKind(), nil
	}
	return t.mapper.KindFor(resource)
}

func (t *thirdPartyResourceDataMapper) RESTMapping(gk unversioned.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	if len(versions) != 1 {
		return nil, fmt.Errorf("unexpected set of versions: %v", versions)
	}
	if gk.Group != t.group {
		return nil, fmt.Errorf("unknown group %q expected %s", gk.Group, t.group)
	}
	if gk.Kind != "ThirdPartyResourceData" {
		return nil, fmt.Errorf("unknown kind %s expected %s", gk.Kind, t.kind)
	}
	if versions[0] != t.version {
		return nil, fmt.Errorf("unknown version %q expected %q", versions[0], t.version)
	}

	// TODO figure out why we're doing this rewriting
	extensionGK := unversioned.GroupKind{Group: extensions.GroupName, Kind: "ThirdPartyResourceData"}

	mapping, err := t.mapper.RESTMapping(extensionGK, registered.GroupOrDie(extensions.GroupName).GroupVersion.Version)
	if err != nil {
		return nil, err
	}
	mapping.ObjectConvertor = &thirdPartyObjectConverter{mapping.ObjectConvertor}
	return mapping, nil
}

func (t *thirdPartyResourceDataMapper) AliasesForResource(resource string) ([]string, bool) {
	return t.mapper.AliasesForResource(resource)
}

func (t *thirdPartyResourceDataMapper) ResourceSingularizer(resource string) (singular string, err error) {
	return t.mapper.ResourceSingularizer(resource)
}

func NewMapper(mapper meta.RESTMapper, kind, version, group string) meta.RESTMapper {
	return &thirdPartyResourceDataMapper{
		mapper:  mapper,
		kind:    kind,
		version: version,
		group:   group,
	}
}

type thirdPartyResourceDataCodecFactory struct {
	runtime.NegotiatedSerializer
	kind     string
	encodeGV unversioned.GroupVersion
	decodeGV unversioned.GroupVersion
}

func NewNegotiatedSerializer(s runtime.NegotiatedSerializer, kind string, encodeGV, decodeGV unversioned.GroupVersion) runtime.NegotiatedSerializer {
	return &thirdPartyResourceDataCodecFactory{
		NegotiatedSerializer: s,

		kind:     kind,
		encodeGV: encodeGV,
		decodeGV: decodeGV,
	}
}

func (t *thirdPartyResourceDataCodecFactory) EncoderForVersion(s runtime.Serializer, gv unversioned.GroupVersion) runtime.Encoder {
	return NewCodec(runtime.NewCodec(
		t.NegotiatedSerializer.EncoderForVersion(s, gv),
		t.NegotiatedSerializer.DecoderToVersion(s, t.decodeGV),
	), t.kind)
}

func (t *thirdPartyResourceDataCodecFactory) DecoderToVersion(s runtime.Serializer, gv unversioned.GroupVersion) runtime.Decoder {
	return NewCodec(runtime.NewCodec(
		t.NegotiatedSerializer.EncoderForVersion(s, t.encodeGV),
		t.NegotiatedSerializer.DecoderToVersion(s, gv),
	), t.kind)
}

type thirdPartyResourceDataCodec struct {
	delegate runtime.Codec
	kind     string
}

func NewCodec(codec runtime.Codec, kind string) runtime.Codec {
	return &thirdPartyResourceDataCodec{codec, kind}
}

func parseObject(data []byte) (map[string]interface{}, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		fmt.Printf("Invalid JSON:\n%s\n", string(data))
		return nil, err
	}
	mapObj, ok := obj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected object: %#v", obj)
	}
	return mapObj, nil
}

func (t *thirdPartyResourceDataCodec) populate(data []byte) (runtime.Object, error) {
	mapObj, err := parseObject(data)
	if err != nil {
		return nil, err
	}
	return t.populateFromObject(mapObj, data)
}

func (t *thirdPartyResourceDataCodec) populateFromObject(mapObj map[string]interface{}, data []byte) (runtime.Object, error) {
	typeMeta := unversioned.TypeMeta{}
	if err := json.Unmarshal(data, &typeMeta); err != nil {
		return nil, err
	}
	switch typeMeta.Kind {
	case t.kind:
		result := &extensions.ThirdPartyResourceData{}
		if err := t.populateResource(result, mapObj, data); err != nil {
			return nil, err
		}
		return result, nil
	case t.kind + "List":
		list := &extensions.ThirdPartyResourceDataList{}
		if err := t.populateListResource(list, mapObj); err != nil {
			return nil, err
		}
		return list, nil
	default:
		return nil, fmt.Errorf("unexpected kind: %s, expected %s", typeMeta.Kind, t.kind)
	}
}

func (t *thirdPartyResourceDataCodec) populateResource(objIn *extensions.ThirdPartyResourceData, mapObj map[string]interface{}, data []byte) error {
	metadata, ok := mapObj["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected object for metadata: %#v", mapObj["metadata"])
	}

	metadataData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(metadataData, &objIn.ObjectMeta); err != nil {
		return err
	}
	// Override API Version with the ThirdPartyResourceData value
	// TODO: fix this hard code
	objIn.APIVersion = v1beta1.SchemeGroupVersion.String()

	objIn.Data = data
	return nil
}

func (t *thirdPartyResourceDataCodec) Decode(data []byte, gvk *unversioned.GroupVersionKind, into runtime.Object) (runtime.Object, *unversioned.GroupVersionKind, error) {
	if into == nil {
		obj, err := t.populate(data)
		if err != nil {
			return nil, nil, err
		}
		return obj, gvk, nil
	}
	thirdParty, ok := into.(*extensions.ThirdPartyResourceData)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected object: %#v", into)
	}

	var dataObj interface{}
	if err := json.Unmarshal(data, &dataObj); err != nil {
		return nil, nil, err
	}
	mapObj, ok := dataObj.(map[string]interface{})
	if !ok {

		return nil, nil, fmt.Errorf("unexpected object: %#v", dataObj)
	}
	/*if gvk.Kind != "ThirdPartyResourceData" {
		return nil, nil, fmt.Errorf("unexpected kind: %s", gvk.Kind)
	}*/
	actual := &unversioned.GroupVersionKind{}
	if kindObj, found := mapObj["kind"]; !found {
		if gvk == nil {
			return nil, nil, runtime.NewMissingKindErr(string(data))
		}
		mapObj["kind"] = gvk.Kind
		actual.Kind = gvk.Kind
	} else {
		kindStr, ok := kindObj.(string)
		if !ok {
			return nil, nil, fmt.Errorf("unexpected object for 'kind': %v", kindObj)
		}
		if kindStr != t.kind {
			return nil, nil, fmt.Errorf("kind doesn't match, expecting: %s, got %s", gvk.Kind, kindStr)
		}
		actual.Kind = t.kind
	}
	if versionObj, found := mapObj["apiVersion"]; !found {
		if gvk == nil {
			return nil, nil, runtime.NewMissingVersionErr(string(data))
		}
		mapObj["apiVersion"] = gvk.GroupVersion().String()
		actual.Group, actual.Version = gvk.Group, gvk.Version
	} else {
		versionStr, ok := versionObj.(string)
		if !ok {
			return nil, nil, fmt.Errorf("unexpected object for 'apiVersion': %v", versionObj)
		}
		if gvk != nil && versionStr != gvk.GroupVersion().String() {
			return nil, nil, fmt.Errorf("version doesn't match, expecting: %v, got %s", gvk.GroupVersion(), versionStr)
		}
		gv, err := unversioned.ParseGroupVersion(versionStr)
		if err != nil {
			return nil, nil, err
		}
		actual.Group, actual.Version = gv.Group, gv.Version
	}

	mapObj, err := parseObject(data)
	if err != nil {
		return nil, actual, err
	}
	if err := t.populateResource(thirdParty, mapObj, data); err != nil {
		return nil, actual, err
	}
	return thirdParty, actual, nil
}

func (t *thirdPartyResourceDataCodec) populateListResource(objIn *extensions.ThirdPartyResourceDataList, mapObj map[string]interface{}) error {
	items, ok := mapObj["items"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected object for items: %#v", mapObj["items"])
	}
	objIn.Items = make([]extensions.ThirdPartyResourceData, len(items))
	for ix := range items {
		objData, err := json.Marshal(items[ix])
		if err != nil {
			return err
		}
		objMap, err := parseObject(objData)
		if err != nil {
			return err
		}
		if err := t.populateResource(&objIn.Items[ix], objMap, objData); err != nil {
			return err
		}
	}
	return nil
}

const template = `{
  "kind": "%s",
  "items": [ %s ]
}`

func encodeToJSON(obj *extensions.ThirdPartyResourceData, stream io.Writer) error {
	var objOut interface{}
	if err := json.Unmarshal(obj.Data, &objOut); err != nil {
		return err
	}
	objMap, ok := objOut.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected type: %v", objOut)
	}
	objMap["metadata"] = obj.ObjectMeta
	encoder := json.NewEncoder(stream)
	return encoder.Encode(objMap)
}

func (t *thirdPartyResourceDataCodec) EncodeToStream(obj runtime.Object, stream io.Writer, overrides ...unversioned.GroupVersion) (err error) {
	switch obj := obj.(type) {
	case *extensions.ThirdPartyResourceData:
		return encodeToJSON(obj, stream)
	case *extensions.ThirdPartyResourceDataList:
		// TODO: There must be a better way to do this...
		dataStrings := make([]string, len(obj.Items))
		for ix := range obj.Items {
			buff := &bytes.Buffer{}
			err := encodeToJSON(&obj.Items[ix], buff)
			if err != nil {
				return err
			}
			dataStrings[ix] = buff.String()
		}
		fmt.Fprintf(stream, template, t.kind+"List", strings.Join(dataStrings, ","))
		return nil
	case *unversioned.Status, *unversioned.APIResourceList:
		return t.delegate.EncodeToStream(obj, stream, overrides...)
	default:
		return fmt.Errorf("unexpected object to encode: %#v", obj)
	}
}

func NewObjectCreator(group, version string, delegate runtime.ObjectCreater) runtime.ObjectCreater {
	return &thirdPartyResourceDataCreator{group, version, delegate}
}

type thirdPartyResourceDataCreator struct {
	group    string
	version  string
	delegate runtime.ObjectCreater
}

func (t *thirdPartyResourceDataCreator) New(kind unversioned.GroupVersionKind) (out runtime.Object, err error) {
	switch kind.Kind {
	case "ThirdPartyResourceData":
		if apiutil.GetGroupVersion(t.group, t.version) != kind.GroupVersion().String() {
			return nil, fmt.Errorf("unknown kind %v", kind)
		}
		return &extensions.ThirdPartyResourceData{}, nil
	case "ThirdPartyResourceDataList":
		if apiutil.GetGroupVersion(t.group, t.version) != kind.GroupVersion().String() {
			return nil, fmt.Errorf("unknown kind %v", kind)
		}
		return &extensions.ThirdPartyResourceDataList{}, nil
	// TODO: this list needs to be formalized higher in the chain
	case "ListOptions", "WatchEvent":
		if apiutil.GetGroupVersion(t.group, t.version) == kind.GroupVersion().String() {
			// Translate third party group to external group.
			gvk := registered.EnabledVersionsForGroup(api.GroupName)[0].WithKind(kind.Kind)
			return t.delegate.New(gvk)
		}
		return t.delegate.New(kind)
	default:
		return t.delegate.New(kind)
	}
}

func NewThirdPartyParameterCodec(p runtime.ParameterCodec) runtime.ParameterCodec {
	return &thirdPartyParameterCodec{p}
}

type thirdPartyParameterCodec struct {
	delegate runtime.ParameterCodec
}

func (t *thirdPartyParameterCodec) DecodeParameters(parameters url.Values, from unversioned.GroupVersion, into runtime.Object) error {
	return t.delegate.DecodeParameters(parameters, v1.SchemeGroupVersion, into)
}

func (t *thirdPartyParameterCodec) EncodeParameters(obj runtime.Object, to unversioned.GroupVersion) (url.Values, error) {
	return t.delegate.EncodeParameters(obj, v1.SchemeGroupVersion)
}
