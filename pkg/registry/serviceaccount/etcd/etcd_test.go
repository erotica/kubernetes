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

package etcd

import (
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/rest/resttest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/testapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools/etcdtest"
)

func newEtcdStorage(t *testing.T) (*tools.FakeEtcdClient, tools.StorageInterface) {
	fakeEtcdClient := tools.NewFakeEtcdClient(t)
	fakeEtcdClient.TestIndex = true
	etcdStorage := tools.NewEtcdStorage(fakeEtcdClient, testapi.Codec(), etcdtest.PathPrefix())
	return fakeEtcdClient, etcdStorage
}

func validNewServiceAccount(name string) *api.ServiceAccount {
	return &api.ServiceAccount{
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: api.NamespaceDefault,
		},
		Secrets: []api.ObjectReference{},
	}
}

func TestCreate(t *testing.T) {
	fakeEtcdClient, etcdStorage := newEtcdStorage(t)
	storage := NewStorage(etcdStorage)
	test := resttest.New(t, storage, fakeEtcdClient.SetError)
	serviceAccount := validNewServiceAccount("foo")
	serviceAccount.ObjectMeta = api.ObjectMeta{GenerateName: "foo-"}
	test.TestCreate(
		// valid
		serviceAccount,
		// invalid
		&api.ServiceAccount{},
		&api.ServiceAccount{
			ObjectMeta: api.ObjectMeta{Name: "name with spaces"},
		},
	)
}

func TestUpdate(t *testing.T) {
	fakeEtcdClient, etcdStorage := newEtcdStorage(t)
	storage := NewStorage(etcdStorage)
	test := resttest.New(t, storage, fakeEtcdClient.SetError)
	key, err := storage.KeyFunc(test.TestContext(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	key = etcdtest.AddPrefix(key)

	fakeEtcdClient.ExpectNotFoundGet(key)
	fakeEtcdClient.ChangeIndex = 2
	serviceAccount := validNewServiceAccount("foo")
	existing := validNewServiceAccount("exists")
	existing.Namespace = test.TestNamespace()
	obj, err := storage.Create(test.TestContext(), existing)
	if err != nil {
		t.Fatalf("unable to create object: %v", err)
	}
	older := obj.(*api.ServiceAccount)
	older.ResourceVersion = "1"

	test.TestUpdate(
		serviceAccount,
		existing,
		older,
	)
}
