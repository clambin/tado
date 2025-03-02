package tado

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"os"
	"sync/atomic"
)

var _ oauth2.TokenSource = &persistentTokenSource{}

// A persistentTokenSource implements an oauth2.TokenSource that maintains a stored copy of the active token.
// This allows a process to reuse a valid token from a previous run, avoiding another (possibly manual) authentication flow.
//
// A persistentTokenSource is implemented as a standard oauth2 TokenSource (which maintains an active token and renews it when required),
// combined with a storedToken that stores an encrypted version of the token on disk.
type persistentTokenSource struct {
	TokenSource  oauth2.TokenSource
	currentToken atomic.Value
	storedToken  storedToken
}

func (p *persistentTokenSource) Token() (token *oauth2.Token, err error) {
	if p.TokenSource == nil {
		return nil, errors.New("no token source")
	}

	// get an active token
	token, err = p.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	// save it if needed
	if current := p.currentToken.Load(); current == nil || current.(*oauth2.Token) != token {
		if err = p.storedToken.save(token); err == nil {
			p.currentToken.Store(token)
		}
	}
	return token, err
}

func (p *persistentTokenSource) initialToken() (token *oauth2.Token, err error) {
	token, err = p.storedToken.load()
	if err == nil && !token.Valid() {
		err = errors.New("invalid token")
	}
	if err == nil {
		p.currentToken.Store(token)
	}
	return token, err
}

type storedToken struct {
	path string
	key  []byte
}

func newStoredToken(path string, passphrase string) storedToken {
	key := sha256.Sum256([]byte(passphrase))
	t := storedToken{path: path, key: key[:32]}
	return t
}

func (e storedToken) save(token *oauth2.Token) error {
	bytes, err := json.Marshal(token)
	if err == nil {
		bytes, err = encryptAES(bytes, e.key)
	}
	if err == nil {
		err = os.WriteFile(e.path, bytes, 0600)
	}
	return err
}

func (e storedToken) load() (*oauth2.Token, error) {
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
