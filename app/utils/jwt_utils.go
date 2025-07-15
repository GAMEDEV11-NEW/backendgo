package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTSecretKey is the secret key for JWT token signing and verification
const JWTSecretKey = "your-super-secret-jwt-key-for-game-admin-backend-2024"

// EncryptionSecretKey is the secret key for AES encryption/decryption (must be exactly 32 bytes)
const EncryptionSecretKey = "your-32-byte-encryption-key-here!!"

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	MobileNo      string `json:"mobile_no"`
	DeviceID      string `json:"device_id"`
	UserID        string `json:"user_id"`
	EncryptedData string `json:"encrypted_data,omitempty"` // Encrypted additional data
	jwt.RegisteredClaims
}

// EncryptedJWTData represents the data that gets encrypted within the JWT
type EncryptedJWTData struct {
	MobileNo  string    `json:"mobile_no"`
	DeviceID  string    `json:"device_id"`
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	FCMToken  string    `json:"fcm_token"` // Added FCM token
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SimpleJWTData represents the simplified data structure for JWT tokens
type SimpleJWTData struct {
	MobileNo string `json:"mobile_no"`
	DeviceID string `json:"device_id"`
	FCMToken string `json:"fcm_token"`
}

// encryptData encrypts data using AES-256-GCM
func encryptData(data []byte) (string, error) {
	// Ensure the key is exactly 32 bytes
	key := []byte(EncryptionSecretKey)
	if len(key) != 32 {
		// Pad or truncate to 32 bytes
		if len(key) < 32 {
			paddedKey := make([]byte, 32)
			copy(paddedKey, key)
			key = paddedKey
		} else {
			key = key[:32]
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptData decrypts data using AES-256-GCM
func decryptData(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	// Ensure the key is exactly 32 bytes
	key := []byte(EncryptionSecretKey)
	if len(key) != 32 {
		// Pad or truncate to 32 bytes
		if len(key) < 32 {
			paddedKey := make([]byte, 32)
			copy(paddedKey, key)
			key = paddedKey
		} else {
			key = key[:32]
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	return plaintext, nil
}

// GenerateEncryptedJWTToken generates a JWT token with encrypted data for the given user
func GenerateEncryptedJWTToken(mobileNo, deviceID, userID, sessionID string) (string, error) {
	return GenerateEncryptedJWTTokenWithFCM(mobileNo, deviceID, userID, sessionID, "")
}

// GenerateEncryptedJWTTokenWithFCM generates a JWT token with encrypted data including FCM token
func GenerateEncryptedJWTTokenWithFCM(mobileNo, deviceID, userID, sessionID, fcmToken string) (string, error) {
	// Validate input parameters
	if mobileNo == "" {
		return "", fmt.Errorf("mobile number cannot be empty")
	}
	if deviceID == "" {
		return "", fmt.Errorf("device ID cannot be empty")
	}
	if userID == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}
	if sessionID == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	// Create encrypted data
	encryptedData := EncryptedJWTData{
		MobileNo:  mobileNo,
		DeviceID:  deviceID,
		UserID:    userID,
		SessionID: sessionID,
		FCMToken:  fcmToken, // Include FCM token if provided
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Convert to JSON for encryption
	jsonData, err := json.Marshal(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted data: %v", err)
	}

	// Encrypt the data
	encryptedString, err := encryptData(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Create JWT claims
	claims := JWTClaims{
		MobileNo:      mobileNo,
		DeviceID:      deviceID,
		UserID:        userID,
		EncryptedData: encryptedString,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "game-admin-backend",
			Subject:   mobileNo,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

// GenerateJWTToken generates a JWT token for the given user (backward compatibility)
func GenerateJWTToken(mobileNo, deviceID, userID string) (string, error) {
	// Validate input parameters
	if mobileNo == "" {
		return "", fmt.Errorf("mobile number cannot be empty")
	}
	if deviceID == "" {
		return "", fmt.Errorf("device ID cannot be empty")
	}
	if userID == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}

	// Create claims
	claims := JWTClaims{
		MobileNo: mobileNo,
		DeviceID: deviceID,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "game-admin-backend",
			Subject:   mobileNo,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

// VerifyEncryptedJWTToken verifies and decodes an encrypted JWT token
func VerifyEncryptedJWTToken(tokenString string) (*EncryptedJWTData, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %v", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract JWT claims")
	}

	// Decrypt the encrypted data
	if claims.EncryptedData == "" {
		return nil, fmt.Errorf("no encrypted data found in token")
	}

	decryptedData, err := decryptData(claims.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token data: %v", err)
	}

	// Unmarshal the decrypted data
	var encryptedJWTData EncryptedJWTData
	err = json.Unmarshal(decryptedData, &encryptedJWTData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted data: %v", err)
	}

	// Verify mobile number matches
	if encryptedJWTData.MobileNo != claims.MobileNo {
		return nil, fmt.Errorf("mobile number mismatch in encrypted data")
	}

	// Verify device ID matches
	if encryptedJWTData.DeviceID != claims.DeviceID {
		return nil, fmt.Errorf("device ID mismatch in encrypted data")
	}

	// Verify user ID matches
	if encryptedJWTData.UserID != claims.UserID {
		return nil, fmt.Errorf("user ID mismatch in encrypted data")
	}

	// Check if token has expired
	if time.Now().After(encryptedJWTData.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	return &encryptedJWTData, nil
}

// VerifyJWTToken verifies and decodes a JWT token (backward compatibility)
func VerifyJWTToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %v", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract JWT claims")
	}

	return claims, nil
}

// ValidateJWTToken validates a JWT token and returns the claims if valid
func ValidateJWTToken(tokenString string) (*JWTClaims, error) {
	claims, err := VerifyJWTToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Additional validation
	if claims.MobileNo == "" {
		return nil, fmt.Errorf("mobile number is missing in JWT token")
	}

	if claims.DeviceID == "" {
		return nil, fmt.Errorf("device ID is missing in JWT token")
	}

	if claims.UserID == "" {
		return nil, fmt.Errorf("user ID is missing in JWT token")
	}

	return claims, nil
}

// ValidateEncryptedJWTToken validates an encrypted JWT token and returns the decrypted data if valid
func ValidateEncryptedJWTToken(tokenString string) (*EncryptedJWTData, error) {
	encryptedData, err := VerifyEncryptedJWTToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Additional validation
	if encryptedData.MobileNo == "" {
		return nil, fmt.Errorf("mobile number is missing in encrypted JWT token")
	}

	if encryptedData.DeviceID == "" {
		return nil, fmt.Errorf("device ID is missing in encrypted JWT token")
	}

	if encryptedData.UserID == "" {
		return nil, fmt.Errorf("user ID is missing in encrypted JWT token")
	}

	if encryptedData.SessionID == "" {
		return nil, fmt.Errorf("session ID is missing in encrypted JWT token")
	}

	// FCM token is optional, so we don't validate if it's empty
	// if encryptedData.FCMToken == "" {
	// 	return nil, fmt.Errorf("FCM token is missing in encrypted JWT token")
	// }

	return encryptedData, nil
}

// ValidateMobileNumberInToken validates that the mobile number in the token matches the provided mobile number
func ValidateMobileNumberInToken(tokenString, expectedMobileNo string) error {
	if tokenString == "" {
		return fmt.Errorf("token string cannot be empty")
	}
	if expectedMobileNo == "" {
		return fmt.Errorf("expected mobile number cannot be empty")
	}

	// Try encrypted token first
	if encryptedData, err := ValidateEncryptedJWTToken(tokenString); err == nil {
		if encryptedData.MobileNo != expectedMobileNo {
			return fmt.Errorf("mobile number mismatch: expected %s, got %s", expectedMobileNo, encryptedData.MobileNo)
		}
		return nil
	}

	// Fallback to regular JWT token
	claims, err := ValidateJWTToken(tokenString)
	if err != nil {
		return fmt.Errorf("JWT token validation failed: %v", err)
	}

	if claims.MobileNo != expectedMobileNo {
		return fmt.Errorf("mobile number mismatch: expected %s, got %s", expectedMobileNo, claims.MobileNo)
	}

	return nil
}

// RefreshJWTToken creates a new JWT token with extended expiry
func RefreshJWTToken(oldTokenString string) (string, error) {
	if oldTokenString == "" {
		return "", fmt.Errorf("old token string cannot be empty")
	}

	// Try to validate as encrypted token first
	if encryptedData, err := ValidateEncryptedJWTToken(oldTokenString); err == nil {
		return GenerateEncryptedJWTToken(encryptedData.MobileNo, encryptedData.DeviceID, encryptedData.UserID, encryptedData.SessionID)
	}

	// Fallback to regular JWT token
	claims, err := ValidateJWTToken(oldTokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate old token: %v", err)
	}

	return GenerateJWTToken(claims.MobileNo, claims.DeviceID, claims.UserID)
}

// GenerateJWTTokenWithFCM generates a JWT token using mobile_no, device_id, and fcm_token
func GenerateJWTTokenWithFCM(mobileNo, deviceID, fcmToken string) (string, error) {
	// Validate input parameters
	if mobileNo == "" {
		return "", fmt.Errorf("mobile number cannot be empty")
	}
	if deviceID == "" {
		return "", fmt.Errorf("device ID cannot be empty")
	}
	if fcmToken == "" {
		return "", fmt.Errorf("FCM token cannot be empty")
	}

	// Create claims with FCM token
	claims := JWTClaims{
		MobileNo: mobileNo,
		DeviceID: deviceID,
		UserID:   "", // Will be set later when user is created/retrieved
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "game-admin-backend",
			Subject:   mobileNo,
		},
	}

	// Create encrypted data with FCM token
	encryptedData := EncryptedJWTData{
		MobileNo:  mobileNo,
		DeviceID:  deviceID,
		UserID:    "",       // Will be set later
		SessionID: "",       // Will be set later
		FCMToken:  fcmToken, // Include FCM token in encrypted data
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Convert to JSON for encryption
	jsonData, err := json.Marshal(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted data: %v", err)
	}

	// Encrypt the data
	encryptedString, err := encryptData(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Set encrypted data in claims
	claims.EncryptedData = encryptedString

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

// GenerateSimpleJWTToken generates a JWT token with only mobile_no, device_id, and fcm_token
func GenerateSimpleJWTToken(mobileNo, deviceID, fcmToken string) (string, error) {
	// Validate input parameters
	if mobileNo == "" {
		return "", fmt.Errorf("mobile number cannot be empty")
	}
	if deviceID == "" {
		return "", fmt.Errorf("device ID cannot be empty")
	}
	if fcmToken == "" {
		return "", fmt.Errorf("FCM token cannot be empty")
	}

	// Create simple data structure
	simpleData := SimpleJWTData{
		MobileNo: mobileNo,
		DeviceID: deviceID,
		FCMToken: fcmToken,
	}

	// Convert to JSON for encryption
	jsonData, err := json.Marshal(simpleData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal simple data: %v", err)
	}

	// Encrypt the data using AES-256-GCM
	encryptedString, err := encryptData(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Create JWT claims with encrypted data
	claims := JWTClaims{
		MobileNo:      mobileNo,
		DeviceID:      deviceID,
		UserID:        "", // Not used in simple approach
		EncryptedData: encryptedString,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "game-admin-backend",
			Subject:   mobileNo,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}

	return tokenString, nil
}

// VerifySimpleJWTToken verifies and decrypts a simple JWT token
func VerifySimpleJWTToken(tokenString string) (*SimpleJWTData, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %v", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract JWT claims")
	}

	// Decrypt the encrypted data
	if claims.EncryptedData == "" {
		return nil, fmt.Errorf("no encrypted data found in token")
	}

	decryptedData, err := decryptData(claims.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token data: %v", err)
	}

	// Unmarshal the decrypted data
	var simpleJWTData SimpleJWTData
	err = json.Unmarshal(decryptedData, &simpleJWTData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted data: %v", err)
	}

	// Verify mobile number matches
	if simpleJWTData.MobileNo != claims.MobileNo {
		return nil, fmt.Errorf("mobile number mismatch in decrypted data")
	}

	// Verify device ID matches
	if simpleJWTData.DeviceID != claims.DeviceID {
		return nil, fmt.Errorf("device ID mismatch in decrypted data")
	}

	return &simpleJWTData, nil
}

// ValidateSimpleJWTToken validates a simple JWT token and returns the decrypted data if valid
func ValidateSimpleJWTToken(tokenString string) (*SimpleJWTData, error) {
	simpleData, err := VerifySimpleJWTToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Additional validation
	if simpleData.MobileNo == "" {
		return nil, fmt.Errorf("mobile number is missing in simple JWT token")
	}

	if simpleData.DeviceID == "" {
		return nil, fmt.Errorf("device ID is missing in simple JWT token")
	}

	if simpleData.FCMToken == "" {
		return nil, fmt.Errorf("FCM token is missing in simple JWT token")
	}

	return simpleData, nil
}

// DecryptUserData decrypts user_data using the first 32 chars of the JWT token as the key
func DecryptUserData(encryptedData string, jwtToken string) (map[string]interface{}, error) {
	// Use first 32 chars of JWT as key
	key := []byte(jwtToken)
	if len(key) < 32 {
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else {
		key = key[:32]
	}

	iv := make([]byte, 16) // Must match client's IV (all zeros)

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS#7 padding
	plaintext = pkcs7Unpad(plaintext)

	var result map[string]interface{}
	err = json.Unmarshal(plaintext, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// pkcs7Unpad removes PKCS#7 padding
func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
