// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !codes

package testkit

import (
	"flag"
	"testing"
	"time"

	"github.com/pingcap/tidb/ddl/schematracker"
	"github.com/pingcap/tidb/domain"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/session"
	"github.com/pingcap/tidb/store/driver"
	"github.com/pingcap/tidb/store/mockstore"
	"github.com/stretchr/testify/require"
)

// WithTiKV flag is only used for debugging locally with real tikv cluster.
var WithTiKV = flag.String("with-tikv", "", "address of tikv cluster, if set, running test with real tikv cluster")

// CreateMockStore return a new mock kv.Storage.
func CreateMockStore(t testing.TB, opts ...mockstore.MockTiKVStoreOption) kv.Storage {
	if *WithTiKV != "" {
		var d driver.TiKVDriver
		var err error
		store, err := d.Open("tikv://" + *WithTiKV)
		require.NoError(t, err)

		var dom *domain.Domain
		dom, err = session.BootstrapSession(store)
		t.Cleanup(func() {
			dom.Close()
			err := store.Close()
			require.NoError(t, err)
		})
		require.NoError(t, err)
		return store
	}

	store, _ := CreateMockStoreAndDomain(t, opts...)
	return store
}

// CreateMockStoreAndDomain return a new mock kv.Storage and *domain.Domain.
func CreateMockStoreAndDomain(t testing.TB, opts ...mockstore.MockTiKVStoreOption) (kv.Storage, *domain.Domain) {
	store, err := mockstore.NewMockStore(opts...)
	require.NoError(t, err)
	return schematracker.UnwrapStorage(store), bootstrap(t, store, 0)
}

func bootstrap(t testing.TB, store kv.Storage, lease time.Duration) *domain.Domain {
	session.SetSchemaLease(lease)
	session.DisableStats4Test()
	dom, err := session.BootstrapSession(store)
	require.NoError(t, err)

	dom.SetStatsUpdating(true)

	t.Cleanup(func() {
		dom.Close()
		err := store.Close()
		require.NoError(t, err)
	})
	return dom
}

// CreateMockStoreWithSchemaLease return a new mock kv.Storage.
func CreateMockStoreWithSchemaLease(t testing.TB, lease time.Duration, opts ...mockstore.MockTiKVStoreOption) kv.Storage {
	store, _ := CreateMockStoreAndDomainWithSchemaLease(t, lease, opts...)
	return schematracker.UnwrapStorage(store)
}

// CreateMockStoreAndDomainWithSchemaLease return a new mock kv.Storage and *domain.Domain.
func CreateMockStoreAndDomainWithSchemaLease(t testing.TB, lease time.Duration, opts ...mockstore.MockTiKVStoreOption) (kv.Storage, *domain.Domain) {
	store, err := mockstore.NewMockStore(opts...)
	require.NoError(t, err)
	return schematracker.UnwrapStorage(store), bootstrap(t, store, lease)
}
