package base

import (
	"context"
	"fmt"
	"testing"

	"github.com/qri-io/dataset"
	"github.com/qri-io/qfs/cafs"
	"github.com/qri-io/qri/dsref"
	"github.com/qri-io/qri/repo"
	reporef "github.com/qri-io/qri/repo/ref"
)

func TestRemoveNVersionsFromStore(t *testing.T) {
	ctx := context.Background()
	r := newTestRepo(t)

	bad := []struct {
		description string
		store       repo.Repo
		ref         dsref.Ref
		n           int
		err         string
	}{
		{"No repo", nil, dsref.Ref{}, 0, "need a repo"},
		{"No ref.Path", r, dsref.Ref{}, 0, "need a dataset reference with a path"},
		{"invalid n", r, dsref.Ref{Path: "path"}, -2, "invalid 'n', n should be n >= 0 or n == -1 to indicate removing all versions"},
	}

	for _, c := range bad {
		_, err := RemoveNVersionsFromStore(ctx, c.store, c.ref, c.n)
		if err == nil {
			t.Errorf("case %s expected: '%s', got no error", c.description, c.err)
			continue
		}
		if c.err != err.Error() {
			t.Errorf("case %s error mismatch. expected: '%s', got: '%s'", c.description, c.err, err.Error())
		}
	}

	// create test repo and history
	// create history of 10 versions
	initDs := addCitiesDataset(t, r)
	refs := []*reporef.DatasetRef{&initDs}
	historyTotal := 10
	for i := 2; i <= historyTotal; i++ {
		update := updateCitiesDataset(t, r, fmt.Sprintf("example city data version %d", i))
		refs = append(refs, &update)
	}

	good := []struct {
		description string
		n           int
	}{
		{"remove n == 0 versions", 0},
		{"remove n == 1 versions", 1},
		{"remove n == 3 versions", 3},
		// should not error when n is greater then the length of history
		{"remove n == 10 versions", 10},
	}

	for _, c := range good {
		// remove
		latestRef := reporef.ConvertToDsref(*refs[len(refs)-1])
		_, err := RemoveNVersionsFromStore(ctx, r, latestRef, c.n)
		if err != nil {
			t.Errorf("case '%s', unexpected err: %s", c.description, err.Error())
		}
		// verifyRefsRemoved will return an empty string
		// if the correct number of refs have been removed
		s := verifyRefsRemoved(ctx, r.Store(), refs, c.n)
		if s != "" {
			t.Errorf("case '%s', refs removed incorrectly: %s", c.description, s)
		}
		shorten := len(refs) - c.n
		if shorten < 0 {
			shorten = len(refs)
		}
		refs = refs[:shorten]
	}

	// remove the ds reference to the cities dataset before we populate
	// the repo with cities datasets again
	r.DeleteRef(initDs)

	// create test repo and history
	// create history of 10 versions
	initDs = addCitiesDataset(t, r)
	refs = []*reporef.DatasetRef{&initDs}
	for i := 2; i <= historyTotal; i++ {
		update := updateCitiesDataset(t, r, fmt.Sprintf("example city data version %d", i))
		refs = append(refs, &update)
	}
	latestRef := reporef.ConvertToDsref(*refs[len(refs)-1])
	_, err := RemoveNVersionsFromStore(ctx, r, latestRef, -1)
	if err != nil {
		t.Errorf("case 'remove all', unexpected err: %s", err.Error())
	}
	s := verifyRefsRemoved(ctx, r.Store(), refs, len(refs))
	if s != "" {
		t.Errorf("case 'remove all', refs removed incorrectly: %s", s)
	}

}

// takes store s, where datasets have been added/removed
// takes a list of refs, where refs[0] is the initial (oldest) dataset
// take int n where n is the number of MOST RECENT datasets that should
// have been removed
// assumes that each Dataset has a Meta component with a Title
func verifyRefsRemoved(ctx context.Context, s cafs.Filestore, refs []*reporef.DatasetRef, n int) string {

	// datasets from index len(refs) - n - 1 SHOULD EXISTS
	// we should error if they DON't exist
	errString := ""
	for i, ref := range refs {
		// datasets from index len(refs) - 1 to len(refs) - n SHOULD NOT EXISTS
		// we should error if they exist

		exists, err := s.Has(ctx, ref.Path)
		if err != nil {
			return fmt.Sprintf("error checking ref '%s' with title '%s' from store", ref.Dataset.Path, ref.Dataset.Meta.Title)
		}

		// datasets that are less then len(refs) - n, should exist
		if i < len(refs)-n {
			if exists == true {
				continue
			}
			errString += fmt.Sprintf("\nref '%s' should exist in the store, but does NOT", ref.Dataset.Meta.Title)
			continue
		}

		// datasets that are greater then len(refs) - n, should NOT exist
		if exists == false {
			continue
		}
		errString += fmt.Sprintf("\nref '%s' should NOT exist in the store, but does", ref.Dataset.Meta.Title)

	}
	return errString
}

func TestVerifyRefsRemove(t *testing.T) {
	ctx := context.Background()
	r := newTestRepo(t)
	// create test repo and history
	// create history of 10 versions
	initDs := addCitiesDataset(t, r)

	//
	refs := []*reporef.DatasetRef{&initDs}
	historyTotal := 3
	for i := 2; i <= historyTotal; i++ {
		update := updateCitiesDataset(t, r, fmt.Sprintf("example city data version %d", i))
		refs = append(refs, &update)
	}
	// test that all real refs exist
	// aka n = 0
	s := verifyRefsRemoved(ctx, r.Store(), refs, 0)
	if s != "" {
		t.Errorf("case 'all refs should exist' should return empty string, got '%s'", s)
	}

	// test that when we have refs in the store
	// but we say that there should be no refs in the store
	// we get the proper response:
	s = verifyRefsRemoved(ctx, r.Store(), refs, 2)
	sExpected := "\nref 'example city data version 2' should NOT exist in the store, but does\nref 'example city data version 3' should NOT exist in the store, but does"
	if s != sExpected {
		t.Errorf("case 'all refs exist, but we say 2 should not' response mismatch: expected '%s', got '%s'", sExpected, s)
	}

	for i := 0; i < 3; i++ {
		fakeRef := reporef.DatasetRef{
			Path: fmt.Sprintf("/map/%d", i),
			Dataset: &dataset.Dataset{
				Meta: &dataset.Meta{
					Title: fmt.Sprintf("Fake Ref version %d", i),
				},
			},
		}
		refs = append(refs, &fakeRef)
	}
	// test that all real refs exist in store
	// and all fake refs do not exist in store
	// aka n = 3
	s = verifyRefsRemoved(ctx, r.Store(), refs, 3)
	if s != "" {
		t.Errorf("case '3 fake refs, with n == 3' should return empty string, got '%s'", s)
	}

	// test that when we say we do have refs in the store
	// but we really don't, we get the proper response:
	s = verifyRefsRemoved(ctx, r.Store(), refs, 0)
	sExpected = `
ref 'Fake Ref version 0' should exist in the store, but does NOT
ref 'Fake Ref version 1' should exist in the store, but does NOT
ref 'Fake Ref version 2' should exist in the store, but does NOT`
	if s != sExpected {
		t.Errorf("case 'expect empty refs to exist' response mismatch: expected '%s', got '%s'", sExpected, s)
	}
}
