package auth

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/json"
    "fmt"
    "io"
    "sync"
    "time"
    
    "golang.org/x/oauth2"
)

// MemoryTokenStore implementa TokenStore usando memoria
type MemoryTokenStore struct {
    tokens map[string]*EncryptedToken
    gcm    cipher.AEAD
    mu     sync.RWMutex
}

// EncryptedToken representa un token encriptado
type EncryptedToken struct {
    Data      []byte    `json:"data"`
    Nonce     []byte    `json:"nonce"`
    Timestamp time.Time `json:"timestamp"`
}

// NewMemoryTokenStore crea un nuevo almacén de tokens en memoria
func NewMemoryTokenStore(encryptionKey []byte) (*MemoryTokenStore, error) {
    if len(encryptionKey) != 32 {
        return nil, fmt.Errorf("la clave de encriptación debe tener 32 bytes")
    }
    
    block, err := aes.NewCipher(encryptionKey)
    if err != nil {
        return nil, fmt.Errorf("error creando cipher: %w", err)
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("error creando GCM: %w", err)
    }
    
    store := &MemoryTokenStore{
        tokens: make(map[string]*EncryptedToken),
        gcm:    gcm,
    }
    
    // Iniciar limpieza periódica
    go store.cleanupRoutine()
    
    return store, nil
}

// StoreToken almacena un token encriptado
func (s *MemoryTokenStore) StoreToken(ctx context.Context, userID string, token *oauth2.Token) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Serializar token
    tokenData, err := json.Marshal(token)
    if err != nil {
        return fmt.Errorf("error serializando token: %w", err)
    }
    
    // Generar nonce
    nonce := make([]byte, s.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return fmt.Errorf("error generando nonce: %w", err)
    }
    
    // Encriptar token
    encryptedData := s.gcm.Seal(nil, nonce, tokenData, nil)
    
    s.tokens[userID] = &EncryptedToken{
        Data:      encryptedData,
        Nonce:     nonce,
        Timestamp: time.Now(),
    }
    
    return nil
}

// GetToken obtiene un token desencriptado
func (s *MemoryTokenStore) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
    s.mu.RLock()
    encryptedToken, exists := s.tokens[userID]
    s.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("token no encontrado para usuario %s", userID)
    }
    
    // Desencriptar token
    tokenData, err := s.gcm.Open(nil, encryptedToken.Nonce, encryptedToken.Data, nil)
    if err != nil {
        return nil, fmt.Errorf("error desencriptando token: %w", err)
    }
    
    // Deserializar token
    var token oauth2.Token
    if err := json.Unmarshal(tokenData, &token); err != nil {
        return nil, fmt.Errorf("error deserializando token: %w", err)
    }
    
    return &token, nil
}

// DeleteToken elimina un token
func (s *MemoryTokenStore) DeleteToken(ctx context.Context, userID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    delete(s.tokens, userID)
    return nil
}

func (s *MemoryTokenStore) cleanupRoutine() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        s.cleanup()
    }
}

func (s *MemoryTokenStore) cleanup() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Eliminar tokens antiguos (más de 24 horas)
    cutoff := time.Now().Add(-24 * time.Hour)
    for userID, token := range s.tokens {
        if token.Timestamp.Before(cutoff) {
            delete(s.tokens, userID)
        }
    }
}
