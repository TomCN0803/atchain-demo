package idemix

import (
	"errors"
	bccsp "github.com/IBM/idemix/bccsp/schemes"
)

// NewDummyKeyStore instantiate a dummy key store
// that neither loads nor stores keys
func NewDummyKeyStore() bccsp.KeyStore {
	return &dummyKeyStore{}
}

// dummyKeyStore is a read-only KeyStore that neither loads nor stores keys.
type dummyKeyStore struct {
}

// ReadOnly returns true if this KeyStore is read only, false otherwise.
// If ReadOnly is true then StoreKey will fail.
func (ks *dummyKeyStore) ReadOnly() bool {
	return true
}

// GetKey returns a key object whose SKI is the one passed.
func (ks *dummyKeyStore) GetKey(ski []byte) (bccsp.Key, error) {
	return nil, errors.New("key not found. This is a dummy KeyStore")
}

// StoreKey stores the key k in this KeyStore.
// If this KeyStore is read only then the method will fail.
func (ks *dummyKeyStore) StoreKey(k bccsp.Key) error {
	return errors.New("cannot store key. This is a dummy read-only KeyStore")
}
