package core

import (
	"github.com/ipfs/go-datastore"
	"github.com/qri-io/dataset"
	"github.com/qri-io/dataset/dsfs"
	"github.com/qri-io/qri/repo"
	"testing"
)

func TestDatasetRequestsInit(t *testing.T) {
	cases := []struct {
		p   *InitDatasetParams
		res *dataset.Dataset
		err string
	}{
		{&InitDatasetParams{}, nil, "data file is required"},
		{&InitDatasetParams{Data: badDataFile}, nil, "error determining dataset schema: line 3, column 0: wrong number of fields in line"},
		{&InitDatasetParams{Data: jobsByAutomationFile}, nil, ""},
	}

	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := &dataset.Dataset{}
		err := req.InitDataset(c.p, got)

		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
	}
}

func TestDatasetRequestsList(t *testing.T) {
	cases := []struct {
		p   *ListParams
		res *[]*repo.DatasetRef
		err string
	}{
		{&ListParams{OrderBy: "chaos", Limit: 1, Offset: 0}, nil, ""},
		{&ListParams{OrderBy: "chaos", Limit: 1, Offset: -50}, nil, ""},
		//TODO: get this to work
		// {&ListParams{OrderBy: "chaos", Limit: -6, Offset: -50}, nil, ""},
	}

	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := &[]*repo.DatasetRef{}
		err := req.List(c.p, got)

		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
	}
}

func TestDatasetRequestsGet(t *testing.T) {
	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}
	path, err := mr.GetPath("movies")
	if err != nil {
		t.Errorf("error getting path: %s", err.Error())
		return
	}
	moviesDs, err := dsfs.LoadDataset(ms, path)
	if err != nil {
		t.Errorf("error loading dataset: %s", err.Error())
		return
	}
	cases := []struct {
		p   *GetDatasetParams
		res *dataset.Dataset
		err string
	}{
		//TODO: probably delete some of these
		{&GetDatasetParams{Path: datastore.NewKey("abc"), Name: "ABC", Hash: "123"}, nil, "error loading dataset: error getting file bytes: datastore: key not found"},
		{&GetDatasetParams{Path: path, Name: "ABC", Hash: "123"}, nil, ""},
		{&GetDatasetParams{Path: path, Name: "movies", Hash: "123"}, moviesDs, ""},
		{&GetDatasetParams{Path: path, Name: "cats", Hash: "123"}, moviesDs, ""},
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := &dataset.Dataset{}
		err := req.Get(c.p, got)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
		// if got != c.res && c.checkResult == true {
		// 	t.Errorf("case %d result mismatch: \nexpected \n\t%s, \n\ngot: \n%s", i, c.res, got)
		// }
	}
}

func TestDatasetRequestsSave(t *testing.T) {
	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}
	path, err := mr.GetPath("movies")
	if err != nil {
		t.Errorf("error getting path: %s", err.Error())
		return
	}
	moviesDs, err := dsfs.LoadDataset(ms, path)
	if err != nil {
		t.Errorf("error loading dataset: %s", err.Error())
		return
	}

	cases := []struct {
		p   *SaveParams
		res *dataset.Dataset
		err string
	}{
		//TODO find out why this fails second time but not first
		{&SaveParams{Name: "ABC", Dataset: moviesDs}, nil, ""},
		{&SaveParams{Name: "ABC", Dataset: moviesDs}, nil, "error marshaling dataset abstract structure to json: json: error calling MarshalJSON for type *dataset.Structure: json: error calling MarshalJSON for type dataset.DataFormat: Unknown Data Format"},
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := &dataset.Dataset{}
		err := req.Save(c.p, got)

		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
	}
}

func TestDatasetRequestsDelete(t *testing.T) {
	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}
	path, err := mr.GetPath("movies")
	if err != nil {
		t.Errorf("error getting path: %s", err.Error())
		return
	}

	cases := []struct {
		p   *DeleteParams
		res *dataset.Dataset
		err string
	}{
		{&DeleteParams{Path: path, Name: "movies"}, nil, "delete dataset not yet finished"},
		{&DeleteParams{Path: datastore.NewKey("abc"), Name: "ABC"}, nil, "delete dataset not yet finished"},
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := false
		err := req.Delete(c.p, &got)

		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
	}
}

func TestDatasetRequestsStructuredData(t *testing.T) {
	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}
	path, err := mr.GetPath("movies")
	if err != nil {
		t.Errorf("error getting path: %s", err.Error())
		return
	}
	var df1 dataset.DataFormat = 0
	cases := []struct {
		p   *StructuredDataParams
		res *StructuredData
		err string
	}{
		{&StructuredDataParams{}, nil, "error getting file bytes: datastore: key not found"},
		{&StructuredDataParams{Format: df1, Path: path, Objects: false, Limit: 5, Offset: 0, All: false}, nil, ""},
		{&StructuredDataParams{Format: df1, Path: path, Objects: false, Limit: -5, Offset: -100, All: false}, nil, ""},
		{&StructuredDataParams{Format: df1, Path: path, Objects: false, Limit: -5, Offset: -100, All: true}, nil, ""},
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := &StructuredData{}
		err := req.StructuredData(c.p, got)

		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
	}
}

func TestDatasetRequestsAddDataset(t *testing.T) {
	cases := []struct {
		p   *AddParams
		res *repo.DatasetRef
		err string
	}{
		{&AddParams{Name: "abc", Hash: "hash###"}, nil, "can only add datasets when running an IPFS filestore"},
	}

	mr, ms, err := NewTestRepo()
	if err != nil {
		t.Errorf("error allocating test repo: %s", err.Error())
		return
	}

	req := NewDatasetRequests(ms, mr)
	for i, c := range cases {
		got := &repo.DatasetRef{}
		err := req.AddDataset(c.p, got)

		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch: expected: %s, got: %s", i, c.err, err)
			continue
		}
	}
}