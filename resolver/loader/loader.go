package loader

import (
	"context"
	"fmt"
	"time"

	golog "github.com/ipfs/go-log"
	"github.com/qri-io/dataset"
	"github.com/qri-io/qfs/cafs"
	"github.com/qri-io/qri/base/dsfs"
	"github.com/qri-io/qri/dscache"
	"github.com/qri-io/qri/dscache/dscachefb"
	"github.com/qri-io/qri/dsref"
	"github.com/qri-io/qri/fsi"
	"github.com/qri-io/qri/resolver"
)

var (
	log = golog.Logger("loader")
)

var _ resolver.Resolver = (*DatasetResolver)(nil)

// DatasetResolver is a high-level component that can resolve dataset references
type DatasetResolver struct {
	Dscache *dscache.Dscache
	Store   cafs.Filestore
}

// NewDatasetResolver returns a new DatasetResolver from dscache and a filestore
func NewDatasetResolver(d *dscache.Dscache, store cafs.Filestore) *DatasetResolver {
	return &DatasetResolver{Dscache: d, Store: store}
}

// GetInfo looks up a VersionInfo from an initID
func (dr *DatasetResolver) GetInfo(initID string) *dsref.VersionInfo {
	log.Errorf("TODO(dustmop): Implement me")
	return nil
}

// GetInfoByDsref looks up a VersionInfo from a dataset ref
func (dr *DatasetResolver) GetInfoByDsref(ref dsref.Ref) *dsref.VersionInfo {
	log.Errorf("TODO(dustmop): Implement me")
	return nil
}

// LoadDsref will parse a ref string, resolve it using dscache and fsi, and return the dataset
// along with additional info.
// TODO(dustmop): Remove the info return value after fixing callers that currently rely on it
func (dr *DatasetResolver) LoadDsref(ctx context.Context, refstr string) (*dataset.Dataset, string, dsref.Ref, *dsref.VersionInfo, error) {
	// Parse the refstr
	ref, err := dsref.Parse(refstr)
	if err == dsref.ErrBadCaseName {
		log.Error(dsref.ErrBadCaseShouldRename)
	} else if err != nil {
		return nil, "", ref, nil, err
	}

	// Handle the "me" convenience shortcut
	if ref.Username == "me" && dr.Dscache.DefaultUsername != "" {
		ref.Username = dr.Dscache.DefaultUsername
	}

	// Resolve username to profileID, lookup dataset by profileID + prettyName
	info, err := lookupByName(dr.Dscache, ref)
	if err != nil {
		return nil, "", ref, nil, err
	}

	// Found a versionInfo, fill in ref.
	ref.Name = info.Name
	defaultPath := false
	if ref.Path == "" {
		ref.Path = info.Path
		defaultPath = true
	}

	// Load the dataset head.
	var ds *dataset.Dataset
	if defaultPath && info.FSIPath != "" {
		// Has an FSI Path, load from working directory
		if ds, err = fsi.ReadDir(info.FSIPath); err != nil {
			return nil, "", ref, nil, err
		}
	} else {
		// Load from dsfs
		if ds, err = dsfs.LoadDataset(ctx, dr.Store, ref.Path); err != nil {
			return nil, "", ref, nil, err
		}
	}
	// Set transient info on the returned dataset
	ds.Name = ref.Name
	ds.Peername = ref.Username
	return ds, info.InitID, ref, info, err
}

func lookupByName(dc *dscache.Dscache, ref dsref.Ref) (*dsref.VersionInfo, error) {
	// Convert the username into a profileID
	for i := 0; i < dc.Root.UsersLength(); i++ {
		userAssoc := dscachefb.UserAssoc{}
		dc.Root.Users(&userAssoc, i)
		username := userAssoc.Username()
		profileID := userAssoc.ProfileID()
		if ref.Username == string(username) {
			ref.ProfileID = string(profileID)
			break
		}
	}
	if ref.ProfileID == "" {
		return nil, fmt.Errorf("unknown username %q", ref.Username)
	}
	// Lookup the info, given the profileID/dsname
	for i := 0; i < dc.Root.RefsLength(); i++ {
		r := dscachefb.RefEntryInfo{}
		dc.Root.Refs(&r, i)
		if string(r.ProfileID()) == ref.ProfileID && string(r.PrettyName()) == ref.Name {
			info := convertEntryToVersionInfo(&r)
			return &info, nil
		}
	}
	return nil, fmt.Errorf("dataset ref not found %s/%s", ref.Username, ref.Name)
}

// copied from dscache/dscache.go
func convertEntryToVersionInfo(r *dscachefb.RefEntryInfo) dsref.VersionInfo {
	return dsref.VersionInfo{
		InitID:        string(r.InitID()),
		ProfileID:     string(r.ProfileID()),
		Name:          string(r.PrettyName()),
		Path:          string(r.HeadRef()),
		Published:     r.Published(),
		Foreign:       r.Foreign(),
		MetaTitle:     string(r.MetaTitle()),
		ThemeList:     string(r.ThemeList()),
		BodySize:      int(r.BodySize()),
		BodyRows:      int(r.BodyRows()),
		BodyFormat:    string(r.BodyFormat()),
		NumErrors:     int(r.NumErrors()),
		CommitTime:    time.Unix(r.CommitTime(), 0),
		NumVersions:   int(r.NumVersions()),
		FSIPath:       string(r.FsiPath()),
	}
}
