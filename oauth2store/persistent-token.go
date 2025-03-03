package oauth2store

import (
	"cmp"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"os"
	"sync/atomic"
	"time"
)

const defaultMaxTokenFileAge = 24 * 30 * time.Hour

var _ oauth2.TokenSource = &PersistentTokenSource{}

// A PersistentTokenSource implements an oauth2.TokenSource that maintains a stored copy of the active token.
// This allows a process to reuse a valid token from a previous run, avoiding another (possibly manual) authentication flow.
//
// A PersistentTokenSource is implemented as a standard oauth2 TokenSource (which maintains an active token and renews it when required),
// combined with a token that stores an encrypted version of the token on disk.
type PersistentTokenSource struct {
	TokenSource  oauth2.TokenSource
	currentToken atomic.Value
	storedToken  storedToken
}

// NewPersistentTokenSource returns a PersistentTokenSource, which stores the latest oauth2.Token in path,
// using passphrase to generate the encryption key.
func NewPersistentTokenSource(path string, passphrase string) *PersistentTokenSource {
	key := sha256.Sum256([]byte(passphrase))
	return &PersistentTokenSource{
		storedToken: storedToken{
			path: path,
			key:  key[:sha256.Size],
		},
	}
}

// GetStoredToken returns the token currently stored on disk.  If the token is written longer than maxAge, it returns an error.
//
// Note: GetStoredToken does not check if the token is still valid (i.e, not expired):  even if the access token has expired,
// the refresh token may still be valid.  maxAge can be used to reject stored tokens that are not reasonable new enough
// (e.g. Tadoº's OAuth2 refresh tokens remain valid for 30 days).
func (p *PersistentTokenSource) GetStoredToken(maxAge time.Duration) (*oauth2.Token, error) {
	token, modTime, err := p.storedToken.load()
	if err != nil {
		return nil, err
	}
	if time.Since(modTime) > cmp.Or(maxAge, defaultMaxTokenFileAge) {
		return nil, errors.New("stored token is too old")
	}
	return token, nil
}

// Token implements the oauth2.TokenSource interface: it returns a token, possibly requesting a new token
// if the current one is expired, and stores it on disk.
func (p *PersistentTokenSource) Token() (token *oauth2.Token, err error) {
	if p.TokenSource == nil {
		return nil, errors.New("no token source")
	}

	// get an active token
	if token, err = p.TokenSource.Token(); err != nil {
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

type storedToken struct {
	path string
	key  []byte
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

func (e storedToken) load() (*oauth2.Token, time.Time, error) {
	stats, err := os.Stat(e.path)
	if err != nil {
		return nil, time.Time{}, err
	}
	bytes, err := os.ReadFile(e.path)
	if err == nil {
		bytes, err = decryptAES(bytes, e.key)
	}
	if err != nil {
		return nil, time.Time{}, err
	}
	var token oauth2.Token
	err = json.Unmarshal(bytes, &token)
	return &token, stats.ModTime(), err
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
