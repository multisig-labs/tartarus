package node

import (
	"encoding/hex"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"

	"github.com/multisig-labs/tartarus/models"
)

func Generate() (models.Node, error) {
	certBytes, keyBytes, err := staking.NewCertAndKeyBytes()
	if err != nil {
		return models.Node{}, err
	}

	tlsCert, err := staking.LoadTLSCertFromBytes(keyBytes, certBytes)
	if err != nil {
		return models.Node{}, err
	}

	stakingCert, err := staking.ParseCertificate(tlsCert.Leaf.Raw)
	if err != nil {
		return models.Node{}, err
	}

	nodeID := ids.NodeIDFromCert(stakingCert)

	blsSecret, err := bls.NewSecretKey()
	if err != nil {
		return models.Node{}, err
	}

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

	return models.Node{
		NodeID:        nodeID.String(),
		Cert:          certString,
		Key:           keyString,
		BLSPrivateKey: blsPrivateHex,
		BLSPublicKey:  blsPublicHex,
		BLSSignature:  sigHex,
	}, nil
}
