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

package serviceaccount

import (
	"bytes"
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/registry/secret"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/serviceaccount"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

// RemoveTokenBackoff is the recommended (empirical) retry interval for removing
// a secret reference from a service account when the secret is deleted. It is
// exported for use by custom secret controllers.
var RemoveTokenBackoff = wait.Backoff{
	Steps:    10,
	Duration: 100 * time.Millisecond,
	Jitter:   1.0,
}

// TokensControllerOptions contains options for the TokensController
type TokensControllerOptions struct {
	// TokenGenerator is the generator to use to create new tokens
	TokenGenerator serviceaccount.TokenGenerator
	// ServiceAccountResync is the time.Duration at which to fully re-list service accounts.
	// If zero, re-list will be delayed as long as possible
	ServiceAccountResync time.Duration
	// SecretResync is the time.Duration at which to fully re-list secrets.
	// If zero, re-list will be delayed as long as possible
	SecretResync time.Duration
	// This CA will be added in the secretes of service accounts
	RootCA []byte
}

// NewTokensController returns a new *TokensController.
func NewTokensController(cl clientset.Interface, options TokensControllerOptions) *TokensController {
	e := &TokensController{
		client: cl,
		token:  options.TokenGenerator,
		rootCA: options.RootCA,
	}

	e.serviceAccounts, e.serviceAccountController = framework.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return e.client.Core().ServiceAccounts(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return e.client.Core().ServiceAccounts(api.NamespaceAll).Watch(options)
			},
		},
		&api.ServiceAccount{},
		options.ServiceAccountResync,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    e.serviceAccountAdded,
			UpdateFunc: e.serviceAccountUpdated,
			DeleteFunc: e.serviceAccountDeleted,
		},
		cache.Indexers{"namespace": cache.MetaNamespaceIndexFunc},
	)

	tokenSelector := fields.SelectorFromSet(map[string]string{api.SecretTypeField: string(api.SecretTypeServiceAccountToken)})
	e.secrets, e.secretController = framework.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				options.FieldSelector = tokenSelector
				return e.client.Core().Secrets(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				options.FieldSelector = tokenSelector
				return e.client.Core().Secrets(api.NamespaceAll).Watch(options)
			},
		},
		&api.Secret{},
		options.SecretResync,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    e.secretAdded,
			UpdateFunc: e.secretUpdated,
			DeleteFunc: e.secretDeleted,
		},
		cache.Indexers{"namespace": cache.MetaNamespaceIndexFunc},
	)

	e.serviceAccountsSynced = e.serviceAccountController.HasSynced
	e.secretsSynced = e.secretController.HasSynced

	return e
}

// TokensController manages ServiceAccountToken secrets for ServiceAccount objects
type TokensController struct {
	stopChan chan struct{}

	client clientset.Interface
	token  serviceaccount.TokenGenerator

	rootCA []byte

	serviceAccounts cache.Indexer
	secrets         cache.Indexer

	// Since we join two objects, we'll watch both of them with controllers.
	serviceAccountController *framework.Controller
	secretController         *framework.Controller

	// These are here so tests can inject a 'return true'.
	serviceAccountsSynced func() bool
	secretsSynced         func() bool
}

// Runs controller loops and returns immediately
func (e *TokensController) Run() {
	if e.stopChan == nil {
		e.stopChan = make(chan struct{})
		go e.serviceAccountController.Run(e.stopChan)
		go e.secretController.Run(e.stopChan)
	}
}

// Stop gracefully shuts down this controller
func (e *TokensController) Stop() {
	if e.stopChan != nil {
		close(e.stopChan)
		e.stopChan = nil
	}
}

// serviceAccountAdded reacts to a ServiceAccount creation by creating a corresponding ServiceAccountToken Secret
func (e *TokensController) serviceAccountAdded(obj interface{}) {
	serviceAccount := obj.(*api.ServiceAccount)
	err := e.createSecretIfNeeded(serviceAccount)
	if err != nil {
		glog.Error(err)
	}
}

// serviceAccountUpdated reacts to a ServiceAccount update (or re-list) by ensuring a corresponding ServiceAccountToken Secret exists
func (e *TokensController) serviceAccountUpdated(oldObj interface{}, newObj interface{}) {
	newServiceAccount := newObj.(*api.ServiceAccount)
	err := e.createSecretIfNeeded(newServiceAccount)
	if err != nil {
		glog.Error(err)
	}
}

// serviceAccountDeleted reacts to a ServiceAccount deletion by deleting all corresponding ServiceAccountToken Secrets
func (e *TokensController) serviceAccountDeleted(obj interface{}) {
	serviceAccount, ok := obj.(*api.ServiceAccount)
	if !ok {
		// Unknown type. If we missed a ServiceAccount deletion, the
		// corresponding secrets will be cleaned up during the Secret re-list
		return
	}
	secrets, err := e.listTokenSecrets(serviceAccount)
	if err != nil {
		glog.Error(err)
		return
	}
	for _, secret := range secrets {
		glog.V(4).Infof("Deleting secret %s/%s because service account %s was deleted", secret.Namespace, secret.Name, serviceAccount.Name)
		if err := e.deleteSecret(secret); err != nil {
			glog.Errorf("Error deleting secret %s/%s: %v", secret.Namespace, secret.Name, err)
		}
	}
}

// secretAdded reacts to a Secret create by ensuring the referenced ServiceAccount exists, and by adding a token to the secret if needed
func (e *TokensController) secretAdded(obj interface{}) {
	secret := obj.(*api.Secret)
	serviceAccount, err := e.getServiceAccount(secret, true)
	if err != nil {
		glog.Error(err)
		return
	}
	if serviceAccount == nil {
		glog.V(2).Infof(
			"Deleting new secret %s/%s because service account %s (uid=%s) was not found",
			secret.Namespace, secret.Name,
			secret.Annotations[api.ServiceAccountNameKey], secret.Annotations[api.ServiceAccountUIDKey])
		if err := e.deleteSecret(secret); err != nil {
			glog.Errorf("Error deleting secret %s/%s: %v", secret.Namespace, secret.Name, err)
		}
	} else {
		e.generateTokenIfNeeded(serviceAccount, secret)
	}
}

// secretUpdated reacts to a Secret update (or re-list) by deleting the secret (if the referenced ServiceAccount does not exist)
func (e *TokensController) secretUpdated(oldObj interface{}, newObj interface{}) {
	newSecret := newObj.(*api.Secret)
	newServiceAccount, err := e.getServiceAccount(newSecret, true)
	if err != nil {
		glog.Error(err)
		return
	}
	if newServiceAccount == nil {
		glog.V(2).Infof(
			"Deleting updated secret %s/%s because service account %s (uid=%s) was not found",
			newSecret.Namespace, newSecret.Name,
			newSecret.Annotations[api.ServiceAccountNameKey], newSecret.Annotations[api.ServiceAccountUIDKey])
		if err := e.deleteSecret(newSecret); err != nil {
			glog.Errorf("Error deleting secret %s/%s: %v", newSecret.Namespace, newSecret.Name, err)
		}
	} else {
		e.generateTokenIfNeeded(newServiceAccount, newSecret)
	}
}

// secretDeleted reacts to a Secret being deleted by removing a reference from the corresponding ServiceAccount if needed
func (e *TokensController) secretDeleted(obj interface{}) {
	secret, ok := obj.(*api.Secret)
	if !ok {
		// Unknown type. If we missed a Secret deletion, the corresponding ServiceAccount (if it exists)
		// will get a secret recreated (if needed) during the ServiceAccount re-list
		return
	}

	serviceAccount, err := e.getServiceAccount(secret, false)
	if err != nil {
		glog.Error(err)
		return
	}
	if serviceAccount == nil {
		return
	}

	if err := client.RetryOnConflict(RemoveTokenBackoff, func() error {
		return e.removeSecretReferenceIfNeeded(serviceAccount, secret.Name)
	}); err != nil {
		utilruntime.HandleError(err)
	}
}

// createSecretIfNeeded makes sure at least one ServiceAccountToken secret exists, and is included in the serviceAccount's Secrets list
func (e *TokensController) createSecretIfNeeded(serviceAccount *api.ServiceAccount) error {
	// If the service account references no secrets, short-circuit and create a new one
	if len(serviceAccount.Secrets) == 0 {
		return e.createSecret(serviceAccount)
	}

	// We shouldn't try to validate secret references until the secrets store is synced
	if !e.secretsSynced() {
		return nil
	}

	// If any existing token secrets are referenced by the service account, return
	allSecrets, err := e.listTokenSecrets(serviceAccount)
	if err != nil {
		return err
	}
	referencedSecrets := getSecretReferences(serviceAccount)
	for _, secret := range allSecrets {
		if referencedSecrets.Has(secret.Name) {
			return nil
		}
	}

	// Otherwise create a new token secret
	return e.createSecret(serviceAccount)
}

// createSecret creates a secret of type ServiceAccountToken for the given ServiceAccount
func (e *TokensController) createSecret(serviceAccount *api.ServiceAccount) error {
	// We don't want to update the cache's copy of the service account
	// so add the secret to a freshly retrieved copy of the service account
	serviceAccounts := e.client.Core().ServiceAccounts(serviceAccount.Namespace)
	liveServiceAccount, err := serviceAccounts.Get(serviceAccount.Name)
	if err != nil {
		return err
	}
	if liveServiceAccount.ResourceVersion != serviceAccount.ResourceVersion {
		// our view of the service account is not up to date
		// we'll get notified of an update event later and get to try again
		// this only prevent interactions between successive runs of this controller's event handlers, but that is useful
		glog.V(2).Infof("View of ServiceAccount %s/%s is not up to date, skipping token creation", serviceAccount.Namespace, serviceAccount.Name)
		return nil
	}

	// Build the secret
	secret := &api.Secret{
		ObjectMeta: api.ObjectMeta{
			Name:      secret.Strategy.GenerateName(fmt.Sprintf("%s-token-", serviceAccount.Name)),
			Namespace: serviceAccount.Namespace,
			Annotations: map[string]string{
				api.ServiceAccountNameKey: serviceAccount.Name,
				api.ServiceAccountUIDKey:  string(serviceAccount.UID),
			},
		},
		Type: api.SecretTypeServiceAccountToken,
		Data: map[string][]byte{},
	}

	// Generate the token
	token, err := e.token.GenerateToken(*serviceAccount, *secret)
	if err != nil {
		return err
	}
	secret.Data[api.ServiceAccountTokenKey] = []byte(token)
	secret.Data[api.ServiceAccountNamespaceKey] = []byte(serviceAccount.Namespace)
	if e.rootCA != nil && len(e.rootCA) > 0 {
		secret.Data[api.ServiceAccountRootCAKey] = e.rootCA
	}

	// Save the secret
	if createdToken, err := e.client.Core().Secrets(serviceAccount.Namespace).Create(secret); err != nil {
		return err
	} else {
		// Manually add the new token to the cache store.
		// This prevents the service account update (below) triggering another token creation, if the referenced token couldn't be found in the store
		e.secrets.Add(createdToken)
	}

	liveServiceAccount.Secrets = append(liveServiceAccount.Secrets, api.ObjectReference{Name: secret.Name})

	_, err = serviceAccounts.Update(liveServiceAccount)
	if err != nil {
		// we weren't able to use the token, try to clean it up.
		glog.V(2).Infof("Deleting secret %s/%s because reference couldn't be added (%v)", secret.Namespace, secret.Name, err)
		if err := e.client.Core().Secrets(secret.Namespace).Delete(secret.Name, nil); err != nil {
			glog.Error(err) // if we fail, just log it
		}
	}
	if apierrors.IsConflict(err) {
		// nothing to do.  We got a conflict, that means that the service account was updated.  We simply need to return because we'll get an update notification later
		return nil
	}

	return err
}

// generateTokenIfNeeded populates the token data for the given Secret if not already set
func (e *TokensController) generateTokenIfNeeded(serviceAccount *api.ServiceAccount, secret *api.Secret) error {
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	caData := secret.Data[api.ServiceAccountRootCAKey]
	needsCA := len(e.rootCA) > 0 && bytes.Compare(caData, e.rootCA) != 0

	needsNamespace := len(secret.Data[api.ServiceAccountNamespaceKey]) == 0

	tokenData := secret.Data[api.ServiceAccountTokenKey]
	needsToken := len(tokenData) == 0

	if !needsCA && !needsToken && !needsNamespace {
		return nil
	}

	// Set the CA
	if needsCA {
		secret.Data[api.ServiceAccountRootCAKey] = e.rootCA
	}
	// Set the namespace
	if needsNamespace {
		secret.Data[api.ServiceAccountNamespaceKey] = []byte(secret.Namespace)
	}

	// Generate the token
	if needsToken {
		token, err := e.token.GenerateToken(*serviceAccount, *secret)
		if err != nil {
			return err
		}
		secret.Data[api.ServiceAccountTokenKey] = []byte(token)
	}

	// Set annotations
	secret.Annotations[api.ServiceAccountNameKey] = serviceAccount.Name
	secret.Annotations[api.ServiceAccountUIDKey] = string(serviceAccount.UID)

	// Save the secret
	if _, err := e.client.Core().Secrets(secret.Namespace).Update(secret); err != nil {
		return err
	}
	return nil
}

// deleteSecret deletes the given secret
func (e *TokensController) deleteSecret(secret *api.Secret) error {
	return e.client.Core().Secrets(secret.Namespace).Delete(secret.Name, nil)
}

// removeSecretReferenceIfNeeded updates the given ServiceAccount to remove a reference to the given secretName if needed.
// Returns whether an update was performed, and any error that occurred
func (e *TokensController) removeSecretReferenceIfNeeded(serviceAccount *api.ServiceAccount, secretName string) error {
	// We don't want to update the cache's copy of the service account
	// so remove the secret from a freshly retrieved copy of the service account
	serviceAccounts := e.client.Core().ServiceAccounts(serviceAccount.Namespace)
	serviceAccount, err := serviceAccounts.Get(serviceAccount.Name)
	if err != nil {
		return err
	}

	// Double-check to see if the account still references the secret
	if !getSecretReferences(serviceAccount).Has(secretName) {
		return nil
	}

	secrets := []api.ObjectReference{}
	for _, s := range serviceAccount.Secrets {
		if s.Name != secretName {
			secrets = append(secrets, s)
		}
	}
	serviceAccount.Secrets = secrets

	_, err = serviceAccounts.Update(serviceAccount)
	if err != nil {
		return err
	}

	return nil
}

// getServiceAccount returns the ServiceAccount referenced by the given secret. If the secret is not
// of type ServiceAccountToken, or if the referenced ServiceAccount does not exist, nil is returned
func (e *TokensController) getServiceAccount(secret *api.Secret, fetchOnCacheMiss bool) (*api.ServiceAccount, error) {
	name, _ := serviceAccountNameAndUID(secret)
	if len(name) == 0 {
		return nil, nil
	}

	key := &api.ServiceAccount{ObjectMeta: api.ObjectMeta{Namespace: secret.Namespace}}
	namespaceAccounts, err := e.serviceAccounts.Index("namespace", key)
	if err != nil {
		return nil, err
	}

	for _, obj := range namespaceAccounts {
		serviceAccount := obj.(*api.ServiceAccount)

		if serviceaccount.IsServiceAccountToken(secret, serviceAccount) {
			return serviceAccount, nil
		}
	}

	if fetchOnCacheMiss {
		serviceAccount, err := e.client.Core().ServiceAccounts(secret.Namespace).Get(name)
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		if serviceaccount.IsServiceAccountToken(secret, serviceAccount) {
			return serviceAccount, nil
		}
	}

	return nil, nil
}

// listTokenSecrets returns a list of all of the ServiceAccountToken secrets that
// reference the given service account's name and uid
func (e *TokensController) listTokenSecrets(serviceAccount *api.ServiceAccount) ([]*api.Secret, error) {
	key := &api.Secret{ObjectMeta: api.ObjectMeta{Namespace: serviceAccount.Namespace}}
	namespaceSecrets, err := e.secrets.Index("namespace", key)
	if err != nil {
		return nil, err
	}

	items := []*api.Secret{}
	for _, obj := range namespaceSecrets {
		secret := obj.(*api.Secret)

		if serviceaccount.IsServiceAccountToken(secret, serviceAccount) {
			items = append(items, secret)
		}
	}
	return items, nil
}

// serviceAccountNameAndUID is a helper method to get the ServiceAccount Name and UID from the given secret
// Returns "","" if the secret is not a ServiceAccountToken secret
// If the name or uid annotation is missing, "" is returned instead
func serviceAccountNameAndUID(secret *api.Secret) (string, string) {
	if secret.Type != api.SecretTypeServiceAccountToken {
		return "", ""
	}
	return secret.Annotations[api.ServiceAccountNameKey], secret.Annotations[api.ServiceAccountUIDKey]
}

func getSecretReferences(serviceAccount *api.ServiceAccount) sets.String {
	references := sets.NewString()
	for _, secret := range serviceAccount.Secrets {
		references.Insert(secret.Name)
	}
	return references
}
