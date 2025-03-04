package oauth2store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"os"
	"time"
)

// The TokenStore interface saves and loads oauth2 Tokens. In conjunction with this package's TokenSource,
// it allows calling applications to persist tokens and reuse them between runs.
//
// The main use case is to persist tokens that require manual action to create (e.g. device authentication flow), so that
// a previously created token may be reused if the application is restarted.
type TokenStore interface {
	Save(token *oauth2.Token) error
	Load() (*oauth2.Token, error)
}

var _ TokenStore = &EncryptedFileTokenStore{}

// An EncryptedFileTokenStore stores an encrypted oauth2.Token to a file.
type EncryptedFileTokenStore struct {
	path       string
	key        []byte
	expiration time.Duration
}

// NewEncryptedFileTokenStore returns an EncryptedFileTokenStore that saved a token to path, using passphrase to generate the encryption key.
// Tokens written older than the expiration date are considered expired and are not loaded.
func NewEncryptedFileTokenStore(path, passphrase string, expiration time.Duration) *EncryptedFileTokenStore {
	key := sha256.Sum256([]byte(passphrase))
	return &EncryptedFileTokenStore{
		path:       path,
		key:        key[:],
		expiration: expiration,
	}
}

// Save stores a token to disk.
func (e EncryptedFileTokenStore) Save(token *oauth2.Token) error {
	// encrypt & write the token
	bytes, err := json.Marshal(token)
	if err == nil {
		bytes, err = encryptAES(bytes, e.key)
	}
	if err == nil {
		err = os.WriteFile(e.path, bytes, 0600)
	}
	return err
}

// Load returns a stored token.  If the token is too old (as specified by the expiration parameter), an error is returned.
func (e EncryptedFileTokenStore) Load() (*oauth2.Token, error) {
	// check if the file hasn't expired
	stats, err := os.Stat(e.path)
	if err != nil {
		return nil, err
	}
	if time.Since(stats.ModTime()) > e.expiration {
		return nil, errors.New("token too old")
	}

	// read & decrypt the token
	bytes, err := os.ReadFile(e.path)
	if err == nil {
		bytes, err = decryptAES(bytes, e.key)
	}
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	err = json.Unmarshal(bytes, &token)
	return &token, err
}

// AES encryption
func encryptAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return aesGCM.Seal(nonce, nonce, data, nil), nil
}

// AES decryption
func decryptAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("invalid ciphertext")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
