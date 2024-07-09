package database

import "gorm.io/gorm"

type Node struct {
	gorm.Model
	NodeID         string `gorm:"unique;not null"`
	Cert           string `gorm:"unique;not null"`
	Key            string `gorm:"unique;not null"`
	BLSPrivateKey  string `gorm:"unique;not null"`
	BLSPublicKey   string `gorm:"unique;not null"`
	BLSSignature   string `gorm:"unique;not null"`
	ActiveProvider string
}

func (n Node) String() string {
	return n.NodeID
}
