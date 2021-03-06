package remote

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/qri-io/dataset"
	"github.com/qri-io/qfs"
	"github.com/qri-io/qfs/cafs"
	"github.com/qri-io/qri/config"
	cfgtest "github.com/qri-io/qri/config/test"
	"github.com/qri-io/qri/dsref"
	"github.com/qri-io/qri/p2p"
	p2ptest "github.com/qri-io/qri/p2p/test"
	"github.com/qri-io/qri/repo"
	"github.com/qri-io/qri/repo/profile"
	reporef "github.com/qri-io/qri/repo/ref"
)

func TestAddDataset(t *testing.T) {
	tr, cleanup := newTestRunner(t)
	defer cleanup()

	var psClient *PeerSyncClient
	var nilClient Client
	nilClient = psClient
	if err := nilClient.AddDataset(tr.Ctx, &reporef.DatasetRef{}, ""); err != ErrNoRemoteClient {
		t.Errorf("nil add mismatch. expected: '%s', got: '%s'", ErrNoRemoteClient, err)
	}

	worldBankRef := writeWorldBankPopulation(tr.Ctx, t, tr.NodeA.Repo)

	cli, err := NewClient(tr.NodeB)
	if err != nil {
		t.Fatal(err)
	}

	tr.NodeA.GoOnline()
	tr.NodeB.GoOnline()

	if err := cli.AddDataset(tr.Ctx, &reporef.DatasetRef{Peername: "foo", Name: "bar"}, ""); err == nil {
		t.Error("expected add of invalid ref to error")
	}

	if err := cli.AddDataset(tr.Ctx, &worldBankRef, ""); err != nil {
		t.Error(err.Error())
	}
}

func TestClientFeedsAndPreviews(t *testing.T) {
	tr, cleanup := newTestRunner(t)
	defer cleanup()

	worldBankRef := writeWorldBankPopulation(tr.Ctx, t, tr.NodeA.Repo)
	publishRef(t, tr.NodeA.Repo, &worldBankRef)

	rem := tr.NodeARemote(t)
	server := tr.RemoteTestServer(rem)
	defer server.Close()

	cli := tr.NodeBClient(t)

	feeds, err := cli.Feeds(tr.Ctx, server.URL)
	if err != nil {
		t.Error(err)
	}

	expect := map[string][]dsref.VersionInfo{
		"recent": {
			{
				Username:   "A",
				Name:       "world_bank_population",
				Path:       "/ipfs/QmVeWbw4DJQqWjKXohgTu5JdhVniLPiyb6z6m1duwvXdQe",
				MetaTitle:  "World Bank Population",
				BodySize:   5,
				BodyRows:   1,
				BodyFormat: "json",
			},
		},
	}

	if diff := cmp.Diff(expect, feeds); diff != "" {
		t.Errorf("feeds result mismatch (-want +got): \n%s", diff)
	}

	ds, err := cli.Preview(tr.Ctx, reporef.ConvertToDsref(worldBankRef), server.URL)
	if err != nil {
		t.Error(err)
	}

	expectDs := &dataset.Dataset{
		Body:     []interface{}{float64(100)},
		BodyPath: "/ipfs/QmWVxUKnBmbiXai1Wgu6SuMzyZwYRqjt5TXL8xxghN5hWL",
		Commit: &dataset.Commit{
			Message:   "created dataset",
			Path:      "/ipfs/QmW27MUFMSvPiE3FpmHhSeBZQEuYAppofudDCLvPXVfSLR",
			Qri:       "cm:0",
			Signature: "XLjvPUsiTxtnhkFajlPosxBl+id/tZJB1RWe9BwPpyqg3toIx6qOkhZtXefDh58rX1L0Id1HU0RkVP8sEl0L54d9C4xv25Uzyv3mAvT9VNN5pzblni5TPvU0mHIbawN57hSiywUP3HQLk8VbjRPo6qjpL5DngwvWXe8mAxTPKWwbV9Zx47tJJWImxJC5vLFRUD1KrRarnhYnGRyGaUiOxssaOnzERw49pA/1dDuFCEWghMpARVgWheZCyHN7rVTs+xH8XOTi8/Zz05bKTlpstm57BcCUENKqJgIt7bjsSIh/gHEc+et1A/kO/DBi3vcoKsA1vZI6lFoJzOwlKRKahg==",
			Title:     "initial commit",
		},
		Meta:     &dataset.Meta{Qri: "md:0", Title: "World Bank Population"},
		Name:     "world_bank_population",
		Path:     "/ipfs/QmVeWbw4DJQqWjKXohgTu5JdhVniLPiyb6z6m1duwvXdQe",
		Peername: "A",
		Qri:      "ds:0",
		Structure: &dataset.Structure{
			Checksum: "QmShoKqAQ98zKKgLrSDGKDCkmAf6Ts1pgk5qPCXkaeshej",
			Depth:    1,
			Entries:  1,
			Format:   "json",
			Length:   5,
			Qri:      "st:0",
			Schema:   map[string]interface{}{"type": string("array")},
		},
	}

	// calling meta has the side-effect of allocating dataset.Meta.meta
	// TODO (b5) - this is bad. we need a meta constructor
	expectDs.Meta.Meta()

	if diff := cmp.Diff(expectDs, ds, cmp.AllowUnexported(dataset.Dataset{}, dataset.Meta{})); diff != "" {
		t.Errorf("preview result mismatch (-want +got): \n%s", diff)
	}
}

func newMemRepoTestNode(t *testing.T) *p2p.QriNode {
	ms := cafs.NewMapstore()
	pi := cfgtest.GetTestPeerInfo(0)
	pro := &profile.Profile{
		Peername: "remote_test_peer",
		ID:       profile.IDFromPeerID(pi.PeerID),
		PrivKey:  pi.PrivKey,
	}
	mr, err := repo.NewMemRepo(pro, ms, newTestFS(ms), profile.NewMemStore())
	if err != nil {
		t.Fatal(err.Error())
	}
	node, err := p2p.NewQriNode(mr, config.DefaultP2PForTesting())
	if err != nil {
		t.Fatal(err.Error())
	}
	return node
}

func newTestFS(cafsys cafs.Filestore) qfs.Filesystem {
	return qfs.NewMux(map[string]qfs.Filesystem{
		"cafs": cafsys,
	})
}

// Convert from test nodes to non-test nodes.
// copied from p2p/peers_test.go
func asQriNodes(testPeers []p2ptest.TestablePeerNode) []*p2p.QriNode {
	// Convert from test nodes to non-test nodes.
	peers := make([]*p2p.QriNode, len(testPeers))
	for i, node := range testPeers {
		peers[i] = node.(*p2p.QriNode)
	}
	return peers
}

func connectMapStores(peers []*p2p.QriNode) {
	for i, s0 := range peers {
		for _, s1 := range peers[i+1:] {
			m0 := (s0.Repo.Store()).(*cafs.MapStore)
			m1 := (s1.Repo.Store()).(*cafs.MapStore)
			m0.AddConnection(m1)
		}
	}
}
