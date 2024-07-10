package models

type Node struct {
	NodeID         string `gorm:"unique;not null" json:"node_id"`
	Cert           string `gorm:"unique;not null" json:"cert"`
	Key            string `gorm:"unique;not null" json:"key"`
	BLSPrivateKey  string `gorm:"unique;not null" json:"bls_private"`
	BLSPublicKey   string `gorm:"unique;not null" json:"bls_public"`
	BLSSignature   string `gorm:"unique;not null" json:"bls_signature"`
	ActiveProvider string `json:"active_provider,omitempty"`
}

func (n Node) String() string {
	return n.NodeID
}
