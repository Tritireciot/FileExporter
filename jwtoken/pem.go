package jwtoken

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

//---------------------------------------------------------------------------
// Запакаем в PEM для последующего экспорта
func CreatePEM(key *rsa.PrivateKey) (private *pem.Block, public *pem.Block) {
	// dump private key to file
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}
	return privateKeyBlock, publicKeyBlock
}

//---------------------------------------------------------------------------
func LoadPrivateEM(private *pem.Block) (*rsa.PrivateKey, error) {
	return x509.ParsePKCS1PrivateKey(private.Bytes)
}

//---------------------------------------------------------------------------
func LoadPublicEM(public *pem.Block) (*rsa.PublicKey, error) {
	return x509.ParsePKCS1PublicKey(public.Bytes)
}