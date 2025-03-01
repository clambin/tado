package oauth2

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"os"
	"sync"
	"sync/atomic"
)

var _ oauth2.TokenSource = &PersistentTokenStore{}

type PersistentTokenStore struct {
	TokenStore  *TokenStore
	TokenSource oauth2.TokenSource
}

func (p PersistentTokenStore) Token() (*oauth2.Token, error) {
	token, err := p.TokenStore.Token()
	if err == nil && token.Valid() {
		return token, nil
	}

	if p.TokenSource == nil {
		return nil, errors.New("no token source")
	}

	if token, err = p.TokenSource.Token(); err == nil {
		err = p.TokenStore.Store(token)
	}
	return token, err
}

type TokenStore struct {
	persistentToken
	current  atomic.Pointer[oauth2.Token]
	loadOnce sync.Once
}

type persistentToken interface {
	save(*oauth2.Token) error
	load() (*oauth2.Token, error)
}

func NewEncryptedTokenStore(path string, passphrase string) *TokenStore {
	s := encryptedToken{path: path}
	s.setEncryptionKey(passphrase)
	return &TokenStore{persistentToken: &s}
}

// Token returns the stored token.  For performance reasons, Token() only read the file the first time it is called.
// After that, we use a cached 'current' copy instead.
func (ts *TokenStore) Token() (*oauth2.Token, error) {
	var err error
	if ts.current.Load() == nil {
		ts.loadOnce.Do(func() { err = ts.load() })
	}
	return ts.current.Load(), err
}

// load reads the token file and sets the cached 'current' token.
func (ts *TokenStore) load() error {
	token, err := ts.persistentToken.load()
	if err == nil {
		ts.current.Store(token)
	}
	return err
}

// Store saves the token to disk and updates the cached 'current' token.
func (ts *TokenStore) Store(token *oauth2.Token) error {
	err := ts.persistentToken.save(token)
	if err == nil {
		ts.current.Store(token)
	}
	return err
}

var _ persistentToken = &encryptedToken{}

type encryptedToken struct {
	path string
	key  []byte
}

func (e *encryptedToken) save(token *oauth2.Token) error {
	bytes, err := json.Marshal(token)
	if err == nil {
		bytes, err = encryptAES(bytes, e.key)
	}
	if err == nil {
		err = os.WriteFile(e.path, bytes, 0600)
	}
	return err
}

func (e *encryptedToken) load() (*oauth2.Token, error) {
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

func (e *encryptedToken) setEncryptionKey(passphrase string) {
	key := sha256.Sum256([]byte(passphrase))
	e.key = key[:32]
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
		return nil, fmt.Errorf("invalid ciphertext")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
