package config

import (
	"encoding/json"
	"gofiber/app/models"
	"gofiber/app/services"

	"time"

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
func (h *AuthSocketHandler) SetupAuthHandlers(socket *socketio.Socket) {
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

		// Parse login request
		loginData, ok := event.Data[0].(map[string]interface{})
		if !ok {
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

		// Parse OTP request
		otpData, ok := event.Data[0].(map[string]interface{})
		if !ok {
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

		// Parse profile request
		profileData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "profile_data",
				Message:   "Invalid profile data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

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

	// Set language handler
	socket.On("set:language", func(event *socketio.EventPayload) {
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