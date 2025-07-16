package config

import (
	"encoding/json"
	"gofiber/app/models"
	"gofiber/app/services"

	"time"

	"gofiber/app/utils"

	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	socketio "github.com/doquangtan/socket.io/v4"
)

// AuthSocketHandler handles all authentication-related Socket.IO events
type AuthSocketHandler struct {
	socketService *services.SocketService
}

// NewAuthSocketHandler creates a new auth socket handler instance
func NewAuthSocketHandler(socketService *services.SocketService) *AuthSocketHandler {
	return &AuthSocketHandler{
		socketService: socketService,
	}
}

// SetupAuthHandlers configures all authentication-related Socket.IO event handlers
func (h *AuthSocketHandler) SetupAuthHandlers(socket *socketio.Socket, authFunc func(socket *socketio.Socket, eventName string) (*models.User, error)) {
	// Device info handler
	socket.On("device:info", func(event *socketio.EventPayload) {
		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "device_info",
				Message:   "No device info provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse device info
		deviceInfoData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "device_info",
				Message:   "Invalid device info format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Convert to DeviceInfo struct
		deviceInfoJSON, _ := json.Marshal(deviceInfoData)
		var deviceInfo models.DeviceInfo
		if err := json.Unmarshal(deviceInfoJSON, &deviceInfo); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "device_info",
				Message:   "Failed to parse device info",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Process device info
		response := h.socketService.HandleDeviceInfo(deviceInfo, socket.Id)
		socket.Emit("device:info:ack", response)
	})

	// Login handler
	socket.On("login", func(event *socketio.EventPayload) {
		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "login_data",
				Message:   "No login data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Support both legacy and new encrypted payloads for login
		var loginData map[string]interface{}
		var mobileNo string
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			if encrypted, hasEncrypted := raw["user_data"]; hasEncrypted {
				mobileNo, _ = raw["mobile_no"].(string)
				encStr, _ := encrypted.(string)
				if mobileNo != "" && encStr != "" {
					ciphertext, err := base64.StdEncoding.DecodeString(encStr)
					if err != nil {
						errorResp := models.ConnectionError{
							Status:    "error",
							ErrorCode: models.ErrorCodeInvalidFormat,
							ErrorType: models.ErrorTypeFormat,
							Field:     "user_data",
							Message:   "Failed to base64 decode user_data",
							Timestamp: time.Now().UTC().Format(time.RFC3339),
							SocketID:  socket.Id,
							Event:     "connection_error",
						}
						socket.Emit("connection_error", errorResp)
						return
					}
					decrypted, err := utils.DecryptUserDataWithMobile(encStr, mobileNo)
					if err != nil {
						// Try to print raw decrypted string if possible
						// (simulate what DecryptUserDataWithMobile does internally)
						key := []byte(mobileNo)
						if len(key) < 32 {
							padded := make([]byte, 32)
							copy(padded, key)
							key = padded
						} else {
							key = key[:32]
						}
						iv := make([]byte, 16)
						block, berr := aes.NewCipher(key)
						if berr == nil {
							mode := cipher.NewCBCDecrypter(block, iv)
							plaintext := make([]byte, len(ciphertext))
							mode.CryptBlocks(plaintext, ciphertext)
							// Try to unpad, but print before and after
							if len(plaintext) > 0 {
								unpadLen := int(plaintext[len(plaintext)-1])
								if unpadLen > 0 && unpadLen <= len(plaintext) {
									plaintext = plaintext[:len(plaintext)-unpadLen]
								}
							}
						}
						// Return error to client
						errorResp := models.ConnectionError{
							Status:    "error",
							ErrorCode: models.ErrorCodeInvalidFormat,
							ErrorType: models.ErrorTypeFormat,
							Field:     "user_data",
							Message:   "Failed to decrypt user_data",
							Timestamp: time.Now().UTC().Format(time.RFC3339),
							SocketID:  socket.Id,
							Event:     "connection_error",
						}
						socket.Emit("connection_error", errorResp)
						return
					}
					loginData = decrypted
					if _, ok := loginData["mobile_no"]; !ok && mobileNo != "" {
						loginData["mobile_no"] = mobileNo
					}
				} else {
					errorResp := models.ConnectionError{
						Status:    "error",
						ErrorCode: models.ErrorCodeMissingField,
						ErrorType: models.ErrorTypeField,
						Field:     "mobile_no/user_data",
						Message:   "mobile_no and user_data are required",
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						SocketID:  socket.Id,
						Event:     "connection_error",
					}
					socket.Emit("connection_error", errorResp)
					return
				}
			} else {
				loginData = raw
			}
		} else {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "login_data",
				Message:   "Invalid login data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}
		// Inject socket_id into the login data
		loginData["socket_id"] = socket.Id

		// Convert to LoginRequest struct
		loginJSON, _ := json.Marshal(loginData)
		var loginReq models.LoginRequest
		if err := json.Unmarshal(loginJSON, &loginReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "login_data",
				Message:   "Failed to parse login data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Process login
		response, err := h.socketService.HandleLogin(loginReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "login",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Set socket ID in response
		response.SocketID = socket.Id
		socket.Emit("otp:sent", response)
	})

	// OTP verification handler
	socket.On("verify:otp", func(event *socketio.EventPayload) {
		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "otp_data",
				Message:   "No OTP data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Support both legacy and new encrypted payloads for verify:otp
		var otpData map[string]interface{}
		var mobileNo string
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			if encrypted, hasEncrypted := raw["user_data"]; hasEncrypted {
				mobileNo, _ = raw["mobile_no"].(string)
				encStr, _ := encrypted.(string)
				if mobileNo != "" && encStr != "" {
					decrypted, err := utils.DecryptUserDataWithMobile(encStr, mobileNo)
					if err != nil {
						errorResp := models.ConnectionError{
							Status:    "error",
							ErrorCode: models.ErrorCodeInvalidFormat,
							ErrorType: models.ErrorTypeFormat,
							Field:     "user_data",
							Message:   "Failed to decrypt user_data",
							Timestamp: time.Now().UTC().Format(time.RFC3339),
							SocketID:  socket.Id,
							Event:     "connection_error",
						}
						socket.Emit("connection_error", errorResp)
						return
					}
					otpData = decrypted
					if _, ok := otpData["mobile_no"]; !ok && mobileNo != "" {
						otpData["mobile_no"] = mobileNo
					}
				} else {
					errorResp := models.ConnectionError{
						Status:    "error",
						ErrorCode: models.ErrorCodeMissingField,
						ErrorType: models.ErrorTypeField,
						Field:     "mobile_no/user_data",
						Message:   "mobile_no and user_data are required",
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						SocketID:  socket.Id,
						Event:     "connection_error",
					}
					socket.Emit("connection_error", errorResp)
					return
				}
			} else {
				otpData = raw
			}
		} else {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "otp_data",
				Message:   "Invalid OTP data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Convert to OTPVerificationRequest struct
		otpJSON, _ := json.Marshal(otpData)
		var otpReq models.OTPVerificationRequest
		if err := json.Unmarshal(otpJSON, &otpReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "otp_data",
				Message:   "Failed to parse OTP data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Process OTP verification
		response, err := h.socketService.HandleOTPVerification(otpReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidOTP,
				ErrorType: models.ErrorTypeOTP,
				Field:     "otp",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Set socket ID in response
		response.SocketID = socket.Id
		socket.Emit("otp:verified", response)
	})

	// Set profile handler
	socket.On("set:profile", func(event *socketio.EventPayload) {
		// Authenticate user
		_, err := authFunc(socket, "set:profile")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "profile_data",
				Message:   "No profile data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// --- Begin: JWT + user_data extraction ---
		var profileData map[string]interface{}
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			jwtToken, hasJWT := raw["jwt_token"].(string)
			encStr, hasUserData := raw["user_data"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "jwt_token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			// Validate JWT token (optional: you can extract claims if needed)
			_, err := utils.ValidateSimpleJWTToken(jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "jwt_token",
					Message:   "Failed to validate JWT token",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			// Decrypt user_data using jwt_token as key
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "user_data",
					Message:   "Failed to decrypt user_data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			profileData = decrypted
		} else {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "profile_data",
				Message:   "Invalid profile data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// --- End: JWT + user_data extraction ---

		// Convert to SetProfileRequest struct
		profileJSON, _ := json.Marshal(profileData)
		var profileReq models.SetProfileRequest
		if err := json.Unmarshal(profileJSON, &profileReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "profile_data",
				Message:   "Failed to parse profile data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Process profile setup
		response, err := h.socketService.HandleSetProfile(profileReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeValidation,
				Field:     "profile",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Set socket ID in response
		response.SocketID = socket.Id
		socket.Emit("profile:set", response)
	})

	// Restore session handler (for reconnections)
	socket.On("restore:session", func(event *socketio.EventPayload) {
		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "session_data",
				Message:   "No session data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		var reqData map[string]interface{}
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			jwtToken, hasJWT := raw["jwt_token"].(string)
			encStr, hasUserData := raw["user_data"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "jwt_token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			_, err := utils.ValidateSimpleJWTToken(jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "jwt_token",
					Message:   "Failed to validate JWT token",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "user_data",
					Message:   "Failed to decrypt user_data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			reqData = decrypted
		} else {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "session_data",
				Message:   "Invalid session data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		sessionToken, ok := reqData["session_token"].(string)
		if !ok || sessionToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "session_token",
				Message:   "Session token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		// Try to restore session
		err := h.socketService.GetSessionService().UpdateSessionSocketID(sessionToken, socket.Id)
		if err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "session",
				Message:   "Session not found or expired",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		// Get session data to return user info
		sessionInfo, err := h.socketService.GetSessionService().GetSession(sessionToken)
		if err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "session",
				Message:   "Failed to retrieve session data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		// Send success response
		response := map[string]interface{}{
			"status":        "success",
			"message":       "Session restored successfully",
			"mobile_no":     sessionInfo.MobileNo,
			"session_token": sessionToken,
			"socket_id":     socket.Id,
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"event":         "session:restored",
		}
		socket.Emit("session:restored", response)
	})

	// Logout handler (clears session completely)
	socket.On("logout", func(event *socketio.EventPayload) {
		// Authenticate user first
		_, err := authFunc(socket, "logout")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		// Get session data to clear it
		sessionData, err := h.socketService.GetSessionService().GetSessionBySocket(socket.Id)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "session",
				Message:   "Session not found",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Clear session completely
		err = h.socketService.GetSessionService().DeleteSession(sessionData.SessionToken)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "session",
				Message:   "Failed to logout",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send logout success response
		response := map[string]interface{}{
			"status":    "success",
			"message":   "Logged out successfully",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "logout:success",
		}
		socket.Emit("logout:success", response)
	})

	// Set language handler
	socket.On("set:language", func(event *socketio.EventPayload) {
		// Authenticate user
		_, err := authFunc(socket, "set:language")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "language_data",
				Message:   "No language data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse language request
		langData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "language_data",
				Message:   "Invalid language data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Convert to SetLanguageRequest struct
		langJSON, _ := json.Marshal(langData)
		var langReq models.SetLanguageRequest
		if err := json.Unmarshal(langJSON, &langReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "language_data",
				Message:   "Failed to parse language data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Process language setup
		response, err := h.socketService.HandleSetLanguage(langReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeValidation,
				Field:     "language",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Set socket ID in response
		response.SocketID = socket.Id
		socket.Emit("language:set", response)
	})

}
