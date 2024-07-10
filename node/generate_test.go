package node

import (
	"testing"

	"github.com/multisig-labs/tartarus/utils"
)

func TestGenerate(t *testing.T) {
	n, err := Generate()
	if err != nil {
		t.Fatal(err)
	}

	if n.NodeID == "" {
		t.Fatal("NodeID is empty")
	}

	if n.Cert == "" {
		t.Fatal("Cert is empty")
	}

	if n.Key == "" {
		t.Fatal("Key is empty")
	}

	if n.BLSPrivateKey == "" {
		t.Fatal("BLSPrivateKey is empty")
	}

	if n.BLSPublicKey == "" {
		t.Fatal("BLSPublicKey is empty")
	}

	if n.BLSSignature == "" {
		t.Fatal("BLSSignature is empty")
	}

	// test the node against the API
	valid, err := utils.VerifyBLSViaAPI(n.NodeID, n.BLSPublicKey, n.BLSSignature)
	if err != nil {
		t.Fatal(err)
	}
	
	if !valid {
		t.Fatal("node is invalid")
	}

}
