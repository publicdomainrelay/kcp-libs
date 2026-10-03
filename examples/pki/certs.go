package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"hash/fnv"
	"math/big"
	"strings"
	"sync"
	"time"
)

type ca struct {
	pem string

	serial string
}

var (
	caMu sync.Mutex

	caCache = map[string]ca{}
)

func caFor(name string) (ca, error) {
	if name == "" {
		name = "root"
	}
	caMu.Lock()
	defer caMu.Unlock()
	if cached, ok := caCache[name]; ok {
		return cached, nil
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return ca{}, err
	}
	digest := fnv.New64a()
	_, _ = digest.Write([]byte(name))
	serial := new(big.Int).SetUint64(digest.Sum64())
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Unix(0, 0),
		NotAfter:              time.Unix(4102444800, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return ca{}, err
	}
	entry := ca{
		pem:    string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		serial: colonHex(serial),
	}
	caCache[name] = entry
	return entry, nil
}

func colonHex(value *big.Int) string {
	hex := strings.ToUpper(value.Text(16))
	if len(hex)%2 == 1 {
		hex = "0" + hex
	}
	pairs := make([]string, 0, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		pairs = append(pairs, hex[i:i+2])
	}
	return strings.Join(pairs, ":")
}
