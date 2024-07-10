package node

import (
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/google/uuid"

	db "github.com/multisig-labs/tartarus/database"
)

func generateRandomTempFileNames() (string, string, error) {
	// Get the temporary directory path
	tempDir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		return "", "", err
	}

	u1, err := uuid.NewRandom()
	if err != nil {
		return "", "", err
	}
	u2, err := uuid.NewRandom()
	if err != nil {
		return "", "", err
	}

	// Generate two random filenames
	fname1 := filepath.Join(tempDir, u1.String())
	fname2 := filepath.Join(tempDir, u2.String())

	return fname1, fname2, nil
}

func Generate() (db.Node, error) {
	// generate random temp names for the key and cert
	keyFile, certFile, err := generateRandomTempFileNames()
	if err != nil {
		return db.Node{}, err
	}

	err = staking.InitNodeStakingKeyPair(keyFile, certFile)
	if err != nil {
		return db.Node{}, err
	}
	// delete the files once we're done
	defer os.Remove(keyFile)
	defer os.Remove(certFile)

	// load the cert file
	certBytes, err := os.ReadFile(certFile)
	if err != nil {
		return db.Node{}, err
	}

	// load the key file
	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return db.Node{}, err
	}

	tlsCert, err := staking.LoadTLSCertFromBytes(keyBytes, certBytes)
	if err != nil {
		return db.Node{}, err
	}

	stakingCert, err := staking.ParseCertificate(tlsCert.Leaf.Raw)
	if err != nil {
		return db.Node{}, err
	}

	nodeID := ids.NodeIDFromCert(stakingCert)

	blsSecret, err := bls.NewSecretKey()
	if err != nil {
		return db.Node{}, err
	}

	// get the nodeID bytes for the signature
	nodeIDStr := nodeID.String()

	// sign the nodeID
	blsPublic := bls.PublicFromSecretKey(blsSecret)
	blsPublicBytes := bls.PublicKeyToCompressedBytes(blsPublic)
	blsPrivateBytes := bls.SecretKeyToBytes(blsSecret)

	signature := bls.SignProofOfPossession(blsSecret, blsPublicBytes)
	sigBytes := bls.SignatureToBytes(signature)

	blsPublicHex := hex.EncodeToString(blsPublicBytes)
	blsPrivateHex := hex.EncodeToString(blsPrivateBytes)
	sigHex := hex.EncodeToString(sigBytes)

	certString := string(certBytes)
	keyString := string(keyBytes)

	return db.Node{
		NodeID:        nodeIDStr,
		Cert:          certString,
		Key:           keyString,
		BLSPrivateKey: blsPrivateHex,
		BLSPublicKey:  blsPublicHex,
		BLSSignature:  sigHex,
	}, nil
}
