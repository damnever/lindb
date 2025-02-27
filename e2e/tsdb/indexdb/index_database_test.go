// Licensed to LinDB under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. LinDB licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build integration
// +build integration

package indexdb

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lindb/lindb/kv"
	"github.com/lindb/lindb/models"
	"github.com/lindb/lindb/pkg/fileutil"
	"github.com/lindb/lindb/series/metric"
	"github.com/lindb/lindb/tsdb/indexdb"
	"github.com/lindb/lindb/tsdb/metadb"
	"github.com/lindb/lindb/tsdb/tblstore/tagindex"
	"github.com/lindb/lindb/tsdb/tblstore/tagkeymeta"
)

var (
	dataPath                                     = "index_database_test"
	indexStore, metaStore                        kv.Store
	forwardFamily, invertedFamily, tagMetaFamily kv.Family
	indexDB                                      indexdb.IndexDatabase
	metadata                                     metadb.Metadata
)

func TestMain(m *testing.M) {
	defer func() {
		kv.Options.Store(&kv.StoreOptions{})
		kv.InitStoreManager(nil)
	}()

	kv.Options.Store(&kv.StoreOptions{Dir: dataPath})

	if err := newIndexDatabase(); err != nil {
		panic(err)
	}
	m.Run()
	_ = kv.GetStoreManager().CloseStore("index")
	_ = kv.GetStoreManager().CloseStore("tag_value")
	_ = indexDB.Close()
	_ = metadata.Close()
	if err := fileutil.RemoveDir(dataPath); err != nil {
		panic(err)
	}
}

func TestIndexDatabase_GetOrCreateSeriesID(t *testing.T) {
	seriesID1, isCreate, err := indexDB.GetOrCreateSeriesID("ns", "metric", metric.ID(10), uint64(1234), models.NewDefaultLimits())
	assert.Equal(t, uint32(1), seriesID1)
	assert.True(t, isCreate)
	assert.NoError(t, err)

	err = indexDB.Close()
	assert.NoError(t, err)

	indexDB, err = indexdb.NewIndexDatabase(
		context.TODO(),
		filepath.Join(dataPath, "meta_db"),
		metadata, forwardFamily,
		invertedFamily)
	assert.NoError(t, err)
	assert.NotNil(t, indexDB)

	seriesID2, isCreate, err := indexDB.GetOrCreateSeriesID("ns", "metric", metric.ID(10), uint64(5678), models.NewDefaultLimits())
	assert.True(t, seriesID2 > seriesID1)
	assert.True(t, isCreate)
	assert.NoError(t, err)
}

func newIndexDatabase() (err error) {
	indexStore, err = kv.GetStoreManager().CreateStore("index", kv.DefaultStoreOption())
	if err != nil {
		return err
	}
	forwardFamily, err = indexStore.CreateFamily(
		"forward",
		kv.FamilyOption{
			CompactThreshold: 0,
			Merger:           string(tagindex.SeriesForwardMerger)})
	if err != nil {
		return err
	}
	invertedFamily, err = indexStore.CreateFamily(
		"inverted",
		kv.FamilyOption{
			CompactThreshold: 0,
			Merger:           string(tagindex.SeriesInvertedMerger)})
	if err != nil {
		return err
	}

	metaStore, err = kv.GetStoreManager().CreateStore("tag_value", kv.DefaultStoreOption())
	if err != nil {
		return err
	}

	tagMetaFamily, err = metaStore.CreateFamily(
		"tag_value",
		kv.FamilyOption{
			CompactThreshold: 0,
			Merger:           string(tagkeymeta.MergerName)})
	if err != nil {
		return err
	}
	metadata, err = metadb.NewMetadata(context.TODO(), "test_db", filepath.Join(dataPath, "metadata"), tagMetaFamily)
	if err != nil {
		return err
	}
	indexDB, err = indexdb.NewIndexDatabase(
		context.TODO(),
		filepath.Join(dataPath, "meta_db"),
		metadata, forwardFamily,
		invertedFamily)
	if err != nil {
		return err
	}
	return nil
}
