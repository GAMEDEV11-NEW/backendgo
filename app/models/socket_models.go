package models

import "time"

// DeviceInfo represents device information sent by client
type DeviceInfo struct {
	DeviceID        string   `json:"device_id"`
	DeviceType      string   `json:"device_type"`
	Timestamp       string   `json:"timestamp"`
	Manufacturer    string   `json:"manufacturer,omitempty"`
	Model           string   `json:"model,omitempty"`
	FirmwareVersion string   `json:"firmware_version,omitempty"`
	Capabilities    []string `json:"capabilities,omitempty"`
}

// DeviceInfoResponse represents device info acknowledgment response
type DeviceInfoResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	SocketID  string `json:"socket_id"`
	Event     string `json:"event"`
}

// LoginRequest represents login request from client
type LoginRequest struct {
	MobileNo string `json:"mobile_no"`
	DeviceID string `json:"device_id"`
	FCMToken string `json:"fcm_token"`
	Email    string `json:"email,omitempty"`
}

// LoginResponse represents login response to client
type LoginResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	MobileNo     string `json:"mobile_no"`
	DeviceID     string `json:"device_id"`
	SessionToken string `json:"session_token"`
	OTP          int    `json:"otp"`
	IsNewUser    bool   `json:"is_new_user"`
	Timestamp    string `json:"timestamp"`
	SocketID     string `json:"socket_id"`
	Event        string `json:"event"`
}

// OTPVerificationRequest represents OTP verification request
type OTPVerificationRequest struct {
	MobileNo     string `json:"mobile_no"`
	SessionToken string `json:"session_token"`
	OTP          string `json:"otp"`
}

// OTPVerificationResponse represents OTP verification response
type OTPVerificationResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	MobileNo     string `json:"mobile_no"`
	DeviceID     string `json:"device_id"`
	SessionToken string `json:"session_token"`
	JWTToken     string `json:"jwt_token"`
	UserStatus   string `json:"user_status"`
	Timestamp    string `json:"timestamp"`
	SocketID     string `json:"socket_id"`
	Event        string `json:"event"`
}

// ProfileData represents user profile data
type ProfileData struct {
	Avatar      string                 `json:"avatar,omitempty"`
	Bio         string                 `json:"bio,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// SetProfileRequest represents profile setup request
type SetProfileRequest struct {
	MobileNo     string      `json:"mobile_no"`
	SessionToken string      `json:"session_token"`
	FullName     string      `json:"full_name"`
	State        string      `json:"state"`
	ReferralCode string      `json:"referral_code,omitempty"`
	ReferredBy   string      `json:"referred_by,omitempty"`
	ProfileData  ProfileData `json:"profile_data,omitempty"`
}

// SetProfileResponse represents profile setup response
type SetProfileResponse struct {
	Status         string      `json:"status"`
	Message        string      `json:"message"`
	MobileNo       string      `json:"mobile_no"`
	SessionToken   string      `json:"session_token"`
	FullName       string      `json:"full_name"`
	State          string      `json:"state"`
	ReferralCode   string      `json:"referral_code,omitempty"`
	ReferredBy     string      `json:"referred_by,omitempty"`
	ProfileData    ProfileData `json:"profile_data,omitempty"`
	WelcomeMessage string      `json:"welcome_message"`
	NextSteps      string      `json:"next_steps"`
	Timestamp      string      `json:"timestamp"`
	SocketID       string      `json:"socket_id"`
	Event          string      `json:"event"`
}

// UserPreferences represents user preferences for language settings
type UserPreferences struct {
	DateFormat string `json:"date_format,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`
	Currency   string `json:"currency,omitempty"`
}

// SetLanguageRequest represents language settings request
type SetLanguageRequest struct {
	MobileNo        string          `json:"mobile_no"`
	SessionToken    string          `json:"session_token"`
	LanguageCode    string          `json:"language_code"`
	LanguageName    string          `json:"language_name"`
	RegionCode      string          `json:"region_code,omitempty"`
	Timezone        string          `json:"timezone,omitempty"`
	UserPreferences UserPreferences `json:"user_preferences,omitempty"`
}

// LocalizedMessages represents localized message content
type LocalizedMessages struct {
	Welcome       string `json:"welcome"`
	SetupComplete string `json:"setup_complete"`
	ReadyToPlay   string `json:"ready_to_play"`
	NextSteps     string `json:"next_steps"`
}

// SetLanguageResponse represents language settings response
type SetLanguageResponse struct {
	Status            string            `json:"status"`
	Message           string            `json:"message"`
	MobileNo          string            `json:"mobile_no"`
	SessionToken      string            `json:"session_token"`
	LanguageCode      string            `json:"language_code"`
	LanguageName      string            `json:"language_name"`
	RegionCode        string            `json:"region_code,omitempty"`
	Timezone          string            `json:"timezone,omitempty"`
	UserPreferences   UserPreferences   `json:"user_preferences,omitempty"`
	LocalizedMessages LocalizedMessages `json:"localized_messages"`
	Timestamp         string            `json:"timestamp"`
	SocketID          string            `json:"socket_id"`
	Event             string            `json:"event"`
}

// ConnectResponse represents connection response
type ConnectResponse struct {
	Token     int    `json:"token"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	SocketID  string `json:"socket_id"`
	Status    string `json:"status"`
	Event     string `json:"event"`
}

// ConnectionError represents error response
type ConnectionError struct {
	Status    string                 `json:"status"`
	ErrorCode string                 `json:"error_code"`
	ErrorType string                 `json:"error_type"`
	Field     string                 `json:"field,omitempty"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	SocketID  string                 `json:"socket_id"`
	Event     string                 `json:"event"`
}

// GameState represents current game state
type GameState struct {
	Level  int `json:"level"`
	Score  int `json:"score"`
	Health int `json:"health"`
}

// Coordinates represents player coordinates
type Coordinates struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// PlayerActionRequest represents player action in gameplay
type PlayerActionRequest struct {
	ActionType   string      `json:"action_type"`
	PlayerID     string      `json:"player_id"`
	SessionToken string      `json:"session_token"`
	Coordinates  Coordinates `json:"coordinates"`
	Timestamp    string      `json:"timestamp"`
	GameState    GameState   `json:"game_state"`
}

// PlayerActionResponse represents player action response
type PlayerActionResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ActionID string `json:"action_id"`
}

// User represents a user in the system
type User struct {
	ID              string          `json:"id" bson:"_id,omitempty"`
	MobileNo        string          `json:"mobile_no" bson:"mobile_no"`
	Email           string          `json:"email" bson:"email"`
	FullName        string          `json:"full_name" bson:"full_name"`
	State           string          `json:"state" bson:"state"`
	ReferralCode    string          `json:"referral_code" bson:"referral_code,omitempty"`
	ReferredBy      string          `json:"referred_by" bson:"referred_by,omitempty"`
	ProfileData     ProfileData     `json:"profile_data" bson:"profile_data,omitempty"`
	LanguageCode    string          `json:"language_code" bson:"language_code"`
	LanguageName    string          `json:"language_name" bson:"language_name"`
	RegionCode      string          `json:"region_code" bson:"region_code,omitempty"`
	Timezone        string          `json:"timezone" bson:"timezone,omitempty"`
	UserPreferences UserPreferences `json:"user_preferences" bson:"user_preferences,omitempty"`
	Status          string          `json:"status" bson:"status"` // "new_user", "existing_user"
	CreatedAt       time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" bson:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	UserID       string    `json:"user_id" bson:"user_id"`
	SessionToken string    `json:"session_token" bson:"session_token"`
	JWTToken     string    `json:"jwt_token" bson:"jwt_token"`
	MobileNo     string    `json:"mobile_no" bson:"mobile_no"`
	DeviceID     string    `json:"device_id" bson:"device_id"`
	FCMToken     string    `json:"fcm_token" bson:"fcm_token"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
	IsActive     bool      `json:"is_active" bson:"is_active"`
}

// Generic response structures
type HeartbeatResponse struct {
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
}

type WelcomeResponse struct {
	Success    bool                   `json:"success"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	ServerInfo map[string]interface{} `json:"server_info"`
}

// HealthCheckResponse represents health check response
type HealthCheckResponse struct {
	Success   bool   `json:"success"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Game represents a game in the system
type Game struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Icon        string            `json:"icon"`
	Banner      string            `json:"banner"`
	MinPlayers  int               `json:"min_players"`
	MaxPlayers  int               `json:"max_players"`
	Difficulty  string            `json:"difficulty"`
	Rating      float64           `json:"rating"`
	IsActive    bool              `json:"is_active"`
	IsFeatured  bool              `json:"is_featured"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// GameCategory represents a game category
type GameCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	GameCount   int    `json:"game_count"`
}

// StaticMessageRequest represents request for static message
type StaticMessageRequest struct {
	MobileNo     string `json:"mobile_no"`
	SessionToken string `json:"session_token"`
	MessageType  string `json:"message_type"` // "game_list", "announcements", "updates"
}

// StaticMessageResponse represents static message response
type StaticMessageResponse struct {
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	MobileNo     string                 `json:"mobile_no"`
	SessionToken string                 `json:"session_token"`
	MessageType  string                 `json:"message_type"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    string                 `json:"timestamp"`
	SocketID     string                 `json:"socket_id"`
	Event        string                 `json:"event"`
}

// ServerAnnouncement represents server announcement
type ServerAnnouncement struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`     // "info", "warning", "maintenance", "update"
	Priority  string `json:"priority"` // "low", "medium", "high", "critical"
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// GameUpdate represents game update information
type GameUpdate struct {
	ID          string   `json:"id"`
	GameID      string   `json:"game_id"`
	Version     string   `json:"version"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	BugFixes    []string `json:"bug_fixes"`
	IsRequired  bool     `json:"is_required"`
	CreatedAt   string   `json:"created_at"`
}

// MainScreenRequest represents main screen request with authentication
type MainScreenRequest struct {
	MobileNo    string `json:"mobile_no"`
	FCMToken    string `json:"fcm_token"`
	JWTToken    string `json:"jwt_token"`
	DeviceID    string `json:"device_id"`
	MessageType string `json:"message_type"` // "game_list", "announcements", "updates", "dashboard"
}

// MainScreenResponse represents main screen response
type MainScreenResponse struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	MobileNo    string                 `json:"mobile_no"`
	DeviceID    string                 `json:"device_id"`
	MessageType string                 `json:"message_type"`
	Data        map[string]interface{} `json:"data"`
	UserInfo    map[string]interface{} `json:"user_info"`
	Timestamp   string                 `json:"timestamp"`
	SocketID    string                 `json:"socket_id"`
	Event       string                 `json:"event"`
}

// Supported languages
var SupportedLanguages = map[string]string{
	"en": "English",
	"es": "Spanish",
	"fr": "French",
	"de": "German",
	"hi": "Hindi",
	"zh": "Chinese",
	"ja": "Japanese",
	"ko": "Korean",
	"ar": "Arabic",
	"pt": "Portuguese",
	"ru": "Russian",
}

// Error codes and types
const (
	// Error codes
	ErrorCodeMissingField        = "MISSING_FIELD"
	ErrorCodeInvalidFormat       = "INVALID_FORMAT"
	ErrorCodeEmptyField          = "EMPTY_FIELD"
	ErrorCodeInvalidType         = "INVALID_TYPE"
	ErrorCodeInvalidSession      = "INVALID_SESSION"
	ErrorCodeInvalidOTP          = "INVALID_OTP"
	ErrorCodeMaxAttemptsExceeded = "MAX_ATTEMPTS_EXCEEDED"
	ErrorCodeReferralCodeExists  = "REFERRAL_CODE_EXISTS"
	ErrorCodeVerificationError   = "VERIFICATION_ERROR"
	ErrorCodeSessionVerification = "SESSION_VERIFICATION_ERROR"

	// Error types
	ErrorTypeField          = "FIELD_ERROR"
	ErrorTypeFormat         = "FORMAT_ERROR"
	ErrorTypeValue          = "VALUE_ERROR"
	ErrorTypeType           = "TYPE_ERROR"
	ErrorTypeAuthentication = "AUTHENTICATION_ERROR"
	ErrorTypeOTP            = "OTP_ERROR"
	ErrorTypeValidation     = "VALIDATION_ERROR"
	ErrorTypeSystem         = "SYSTEM_ERROR"
)

// ContestRequest represents contest list request
type ContestRequest struct {
	MobileNo    string `json:"mobile_no"`
	FCMToken    string `json:"fcm_token"`
	JWTToken    string `json:"jwt_token"`
	DeviceID    string `json:"device_id"`
	MessageType string `json:"message_type"` // "contest_list", "contest_details", "contest_join"
	ContestID   string `json:"contest_id,omitempty"`
}

// ContestResponse represents contest list response
type ContestResponse struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	MobileNo    string                 `json:"mobile_no"`
	DeviceID    string                 `json:"device_id"`
	MessageType string                 `json:"message_type"`
	Data        map[string]interface{} `json:"data"`
	UserInfo    map[string]interface{} `json:"user_info"`
	Timestamp   string                 `json:"timestamp"`
	SocketID    string                 `json:"socket_id"`
	Event       string                 `json:"event"`
}

// Contest represents a contest in the system
type Contest struct {
	ContestID         string      `json:"contest_id"`
	ContestName       string      `json:"contest_name"`
	ContestWinPrice   interface{} `json:"contest_win_price"`
	ContestEntryFee   interface{} `json:"contest_entryfee"`
	ContestJoinUser   int         `json:"contest_joinuser"`
	ContestActiveUser int         `json:"contest_activeuser"`
	ContestStartTime  string      `json:"contest_starttime"`
	ContestEndTime    string      `json:"contest_endtime"`
}

// ContestJoinRequest represents contest join request
type ContestJoinRequest struct {
	MobileNo  string `json:"mobile_no"`
	FCMToken  string `json:"fcm_token"`
	JWTToken  string `json:"jwt_token"`
	DeviceID  string `json:"device_id"`
	ContestID string `json:"contest_id"`
	TeamName  string `json:"team_name,omitempty"`
	TeamSize  int    `json:"team_size,omitempty"`
}

// ContestJoinResponse represents contest join response
type ContestJoinResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	MobileNo  string                 `json:"mobile_no"`
	DeviceID  string                 `json:"device_id"`
	ContestID string                 `json:"contest_id"`
	TeamID    string                 `json:"team_id,omitempty"`
	JoinTime  string                 `json:"join_time"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
	SocketID  string                 `json:"socket_id"`
	Event     string                 `json:"event"`
}

// ContestGapRequest represents contest price gap request
type ContestGapRequest struct {
	MobileNo    string `json:"mobile_no"`
	FCMToken    string `json:"fcm_token"`
	JWTToken    string `json:"jwt_token"`
	DeviceID    string `json:"device_id"`
	MessageType string `json:"message_type"` // "price_gap", "entry_fee_gap", "win_price_gap"
}

// ContestGapResponse represents contest price gap response
type ContestGapResponse struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	MobileNo    string                 `json:"mobile_no"`
	DeviceID    string                 `json:"device_id"`
	MessageType string                 `json:"message_type"`
	Data        map[string]interface{} `json:"data"`
	UserInfo    map[string]interface{} `json:"user_info"`
	Timestamp   string                 `json:"timestamp"`
	SocketID    string                 `json:"socket_id"`
	Event       string                 `json:"event"`
}
