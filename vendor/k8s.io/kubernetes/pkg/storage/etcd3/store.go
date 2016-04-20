/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package etcd3

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/conversion"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/storage/etcd"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type store struct {
	client     *clientv3.Client
	codec      runtime.Codec
	versioner  storage.Versioner
	pathPrefix string
}

type elemForDecode struct {
	data []byte
	rev  uint64
}

type objState struct {
	obj  runtime.Object
	meta *storage.ResponseMeta
	rev  int64
	data []byte
}

func newStore(c *clientv3.Client, codec runtime.Codec, prefix string) *store {
	return &store{
		client:     c,
		versioner:  etcd.APIObjectVersioner{},
		codec:      codec,
		pathPrefix: prefix,
	}
}

// Backends implements storage.Interface.Backends.
func (s *store) Backends(ctx context.Context) []string {
	resp, err := s.client.MemberList(ctx)
	if err != nil {
		glog.Errorf("Error obtaining etcd members list: %q", err)
		return nil
	}
	var mlist []string
	for _, member := range resp.Members {
		mlist = append(mlist, member.ClientURLs...)
	}
	return mlist
}

// Codec implements storage.Interface.Codec.
func (s *store) Codec() runtime.Codec {
	return s.codec
}

// Versioner implements storage.Interface.Versioner.
func (s *store) Versioner() storage.Versioner {
	return s.versioner
}

// Get implements storage.Interface.Get.
func (s *store) Get(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool) error {
	key = keyWithPrefix(s.pathPrefix, key)
	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return err
	}

	if len(getResp.Kvs) == 0 {
		if ignoreNotFound {
			return runtime.SetZeroValue(out)
		}
		return storage.NewKeyNotFoundError(key, 0)
	}
	kv := getResp.Kvs[0]
	return decode(s.codec, s.versioner, kv.Value, out, kv.ModRevision)
}

// Create implements storage.Interface.Create.
func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	if version, err := s.versioner.ObjectResourceVersion(obj); err == nil && version != 0 {
		return errors.New("resourceVersion should not be set on objects to be created")
	}
	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		return err
	}
	key = keyWithPrefix(s.pathPrefix, key)

	txnResp, err := s.client.KV.Txn(ctx).If(
		notFound(key),
	).Then(
		clientv3.OpPut(key, string(data)),
	).Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded {
		return storage.NewKeyExistsError(key, 0)
	}

	if out != nil {
		putResp := txnResp.Responses[0].GetResponsePut()
		return decode(s.codec, s.versioner, data, out, putResp.Header.Revision)
	}
	return nil
}

// Delete implements storage.Interface.Delete.
func (s *store) Delete(ctx context.Context, key string, out runtime.Object, precondtions *storage.Preconditions) error {
	v, err := conversion.EnforcePtr(out)
	if err != nil {
		panic("unable to convert output object to pointer")
	}
	key = keyWithPrefix(s.pathPrefix, key)
	if precondtions == nil {
		return s.unconditionalDelete(ctx, key, out)
	}
	return s.conditionalDelete(ctx, key, out, v, precondtions)
}

func (s *store) unconditionalDelete(ctx context.Context, key string, out runtime.Object) error {
	// We need to do get and delete in single transaction in order to
	// know the value and revision before deleting it.
	txnResp, err := s.client.KV.Txn(ctx).If().Then(
		clientv3.OpGet(key),
		clientv3.OpDelete(key),
	).Commit()
	if err != nil {
		return err
	}
	getResp := txnResp.Responses[0].GetResponseRange()
	if len(getResp.Kvs) == 0 {
		return storage.NewKeyNotFoundError(key, 0)
	}

	kv := getResp.Kvs[0]
	return decode(s.codec, s.versioner, kv.Value, out, kv.ModRevision)
}

func (s *store) conditionalDelete(ctx context.Context, key string, out runtime.Object, v reflect.Value, precondtions *storage.Preconditions) error {
	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return err
	}
	for {
		origState, err := s.getState(getResp, key, v, false)
		if err != nil {
			return err
		}
		if err := checkPreconditions(key, precondtions, origState.obj); err != nil {
			return err
		}
		txnResp, err := s.client.KV.Txn(ctx).If(
			clientv3.Compare(clientv3.ModifiedRevision(key), "=", origState.rev),
		).Then(
			clientv3.OpDelete(key),
		).Else(
			clientv3.OpGet(key),
		).Commit()
		if err != nil {
			return err
		}
		if !txnResp.Succeeded {
			getResp = (*clientv3.GetResponse)(txnResp.Responses[0].GetResponseRange())
			glog.V(4).Infof("deletion of %s failed because of a conflict, going to retry", key)
			continue
		}
		return decode(s.codec, s.versioner, origState.data, out, origState.rev)
	}
}

// GuaranteedUpdate implements storage.Interface.GuaranteedUpdate.
func (s *store) GuaranteedUpdate(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool, precondtions *storage.Preconditions, tryUpdate storage.UpdateFunc) error {
	v, err := conversion.EnforcePtr(out)
	if err != nil {
		panic("unable to convert output object to pointer")
	}
	key = keyWithPrefix(s.pathPrefix, key)
	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return err
	}
	for {
		origState, err := s.getState(getResp, key, v, ignoreNotFound)
		if err != nil {
			return err
		}

		if err := checkPreconditions(key, precondtions, origState.obj); err != nil {
			return err
		}

		ret, err := s.updateState(origState, tryUpdate)
		if err != nil {
			return err
		}

		data, err := runtime.Encode(s.codec, ret)
		if err != nil {
			return err
		}
		if bytes.Equal(data, origState.data) {
			return decode(s.codec, s.versioner, origState.data, out, origState.rev)
		}

		txnResp, err := s.client.KV.Txn(ctx).If(
			clientv3.Compare(clientv3.ModifiedRevision(key), "=", origState.rev),
		).Then(
			clientv3.OpPut(key, string(data)),
		).Else(
			clientv3.OpGet(key),
		).Commit()
		if err != nil {
			return err
		}
		if !txnResp.Succeeded {
			getResp = (*clientv3.GetResponse)(txnResp.Responses[0].GetResponseRange())
			glog.V(4).Infof("GuaranteedUpdate of %s failed because of a conflict, going to retry", key)
			continue
		}
		putResp := txnResp.Responses[0].GetResponsePut()
		return decode(s.codec, s.versioner, data, out, putResp.Header.Revision)
	}
}

// GetToList implements storage.Interface.GetToList.
func (s *store) GetToList(ctx context.Context, key string, filter storage.FilterFunc, listObj runtime.Object) error {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return err
	}
	key = keyWithPrefix(s.pathPrefix, key)

	getResp, err := s.client.KV.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(getResp.Kvs) == 0 {
		return nil
	}
	elems := []*elemForDecode{{
		data: getResp.Kvs[0].Value,
		rev:  uint64(getResp.Kvs[0].ModRevision),
	}}
	if err := decodeList(elems, filter, listPtr, s.codec, s.versioner); err != nil {
		return err
	}
	// update version with cluster level revision
	return s.versioner.UpdateList(listObj, uint64(getResp.Header.Revision))
}

// List implements storage.Interface.List.
func (s *store) List(ctx context.Context, key, resourceVersion string, filter storage.FilterFunc, listObj runtime.Object) error {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return err
	}
	key = keyWithPrefix(s.pathPrefix, key)
	// We need to make sure the key ended with "/" so that we only get children "directories".
	// e.g. if we have key "/a", "/a/b", "/ab", getting keys with prefix "/a" will return all three,
	// while with prefix "/a/" will return only "/a/b" which is the correct answer.
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	getResp, err := s.client.KV.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	elems := make([]*elemForDecode, len(getResp.Kvs))
	for i, kv := range getResp.Kvs {
		elems[i] = &elemForDecode{
			data: kv.Value,
			rev:  uint64(kv.ModRevision),
		}
	}
	if err := decodeList(elems, filter, listPtr, s.codec, s.versioner); err != nil {
		return err
	}
	// update version with cluster level revision
	return s.versioner.UpdateList(listObj, uint64(getResp.Header.Revision))
}

// Watch implements storage.Interface.Watch.
func (s *store) Watch(ctx context.Context, key string, resourceVersion string, filter storage.FilterFunc) (watch.Interface, error) {
	panic("TODO: unimplemented")
}

// WatchList implements storage.Interface.WatchList.
func (s *store) WatchList(ctx context.Context, key string, resourceVersion string, filter storage.FilterFunc) (watch.Interface, error) {
	panic("TODO: unimplemented")
}

func (s *store) getState(getResp *clientv3.GetResponse, key string, v reflect.Value, ignoreNotFound bool) (*objState, error) {
	state := &objState{
		obj:  reflect.New(v.Type()).Interface().(runtime.Object),
		meta: &storage.ResponseMeta{},
	}
	if len(getResp.Kvs) == 0 {
		if !ignoreNotFound {
			return nil, storage.NewKeyNotFoundError(key, 0)
		}
		if err := runtime.SetZeroValue(state.obj); err != nil {
			return nil, err
		}
	} else {
		state.rev = getResp.Kvs[0].ModRevision
		state.meta.ResourceVersion = uint64(state.rev)
		state.data = getResp.Kvs[0].Value
		if err := decode(s.codec, s.versioner, state.data, state.obj, state.rev); err != nil {
			return nil, err
		}
	}
	return state, nil
}

func (s *store) updateState(st *objState, userUpdate storage.UpdateFunc) (runtime.Object, error) {
	ret, _, err := userUpdate(st.obj, *st.meta)
	version, err := s.versioner.ObjectResourceVersion(ret)
	if err != nil {
		return nil, err
	}
	if version != 0 {
		// We cannot store object with resourceVersion in etcd. We need to reset it.
		if err := s.versioner.UpdateObject(ret, nil, 0); err != nil {
			return nil, fmt.Errorf("UpdateObject failed: %v", err)
		}
	}
	return ret, nil
}

func keyWithPrefix(prefix, key string) string {
	if strings.HasPrefix(key, prefix) {
		return key
	}
	return path.Join(prefix, key)
}

// decode decodes value of bytes into object. It will also set the object resource version to rev.
// On success, objPtr would be set to the object.
func decode(codec runtime.Codec, versioner storage.Versioner, value []byte, objPtr runtime.Object, rev int64) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		panic("unable to convert output object to pointer")
	}
	_, _, err := codec.Decode(value, nil, objPtr)
	if err != nil {
		return err
	}
	// being unable to set the version does not prevent the object from being extracted
	versioner.UpdateObject(objPtr, nil, uint64(rev))
	return nil
}

// decodeList decodes a list of values into a list of objects, with resource version set to corresponding rev.
// On success, ListPtr would be set to the list of objects.
func decodeList(elems []*elemForDecode, filter storage.FilterFunc, ListPtr interface{}, codec runtime.Codec, versioner storage.Versioner) error {
	v, err := conversion.EnforcePtr(ListPtr)
	if err != nil || v.Kind() != reflect.Slice {
		panic("need ptr to slice")
	}
	for _, elem := range elems {
		obj, _, err := codec.Decode(elem.data, nil, reflect.New(v.Type().Elem()).Interface().(runtime.Object))
		if err != nil {
			return err
		}
		// being unable to set the version does not prevent the object from being extracted
		versioner.UpdateObject(obj, nil, elem.rev)
		if filter(obj) {
			v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
		}
	}
	return nil
}

func checkPreconditions(key string, preconditions *storage.Preconditions, out runtime.Object) error {
	if preconditions == nil {
		return nil
	}
	objMeta, err := api.ObjectMetaFor(out)
	if err != nil {
		return storage.NewInternalErrorf("can't enforce preconditions %v on un-introspectable object %v, got error: %v", *preconditions, out, err)
	}
	if preconditions.UID != nil && *preconditions.UID != objMeta.UID {
		errMsg := fmt.Sprintf("Precondition failed: UID in precondition: %v, UID in object meta: %v", preconditions.UID, objMeta.UID)
		return storage.NewInvalidObjError(key, errMsg)
	}
	return nil
}

func notFound(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModifiedRevision(key), "=", 0)
}
