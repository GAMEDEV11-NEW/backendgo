package services

import (
	"gofiber/app/models"
	

	"github.com/gocql/gocql"
)

// SocketService handles all socket-related business logic
// Now acts as a coordinator between AuthService and GameService
type SocketService struct {
	cassandraSession *gocql.Session
	authService      *AuthService
	gameService      *GameService
}

// NewSocketService creates a new socket service instance using Cassandra
func NewSocketService(cassandraSession *gocql.Session) *SocketService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}
	
	service := &SocketService{
		cassandraSession: cassandraSession,
		authService:      NewAuthService(cassandraSession),
		gameService:      NewGameService(cassandraSession),
	}
	return service
}

// Auth-related methods - delegate to AuthService
func (s *SocketService) GenerateSessionToken() (string, error) {
	return s.authService.GenerateSessionToken()
}

func (s *SocketService) GenerateOTP() int {
	return s.authService.GenerateOTP()
}

func (s *SocketService) HandleDeviceInfo(deviceInfo models.DeviceInfo, socketID string) models.DeviceInfoResponse {
	return s.authService.HandleDeviceInfo(deviceInfo, socketID)
}

func (s *SocketService) HandleLogin(loginReq models.LoginRequest) (*models.LoginResponse, error) {
	return s.authService.HandleLogin(loginReq)
}

func (s *SocketService) HandleOTPVerification(otpReq models.OTPVerificationRequest) (*models.OTPVerificationResponse, error) {
	return s.authService.HandleOTPVerification(otpReq)
}

func (s *SocketService) HandleSetProfile(profileReq models.SetProfileRequest) (*models.SetProfileResponse, error) {
	return s.authService.HandleSetProfile(profileReq)
}

func (s *SocketService) HandleSetLanguage(langReq models.SetLanguageRequest) (*models.SetLanguageResponse, error) {
	return s.authService.HandleSetLanguage(langReq)
}

func (s *SocketService) ValidateSession(sessionToken, mobileNo string) bool {
	return s.authService.ValidateSession(sessionToken, mobileNo)
}

func (s *SocketService) CleanupExpiredSessions() error {
	return s.authService.CleanupExpiredSessions()
}

func (s *SocketService) CleanupExpiredOTPs() error {
	return s.authService.CleanupExpiredOTPs()
}

func (s *SocketService) GetLatestOTP(phoneOrEmail, purpose string) (*models.OTPData, error) {
	return s.authService.GetLatestOTP(phoneOrEmail, purpose)
}

func (s *SocketService) ResendOTP(mobileNo string) (int, error) {
	return s.authService.ResendOTP(mobileNo)
}

// Game-related methods - delegate to GameService
func (s *SocketService) HandlePlayerAction(actionReq models.PlayerActionRequest) (*models.PlayerActionResponse, error) {
	return s.gameService.HandlePlayerAction(actionReq)
}

func (s *SocketService) HandleHeartbeat() models.HeartbeatResponse {
	return s.gameService.HandleHeartbeat()
}

func (s *SocketService) HandleWelcome() models.WelcomeResponse {
	return s.gameService.HandleWelcome()
}

func (s *SocketService) HandleHealthCheck() models.HealthCheckResponse {
	return s.gameService.HandleHealthCheck()
}

func (s *SocketService) HandleStaticMessage(staticReq models.StaticMessageRequest) (*models.StaticMessageResponse, error) {
	return s.gameService.HandleStaticMessage(staticReq)
}

func (s *SocketService) GetGameListFromRedis() (map[string]interface{}, error) {
	return s.gameService.GetGameListFromRedis()
}

func (s *SocketService) GetGameListDataPublic() map[string]interface{} {
	return s.gameService.GetGameListDataPublic()
}

func (s *SocketService) HandleMainScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	return s.gameService.HandleMainScreen(mainReq)
}

func (s *SocketService) HandleContestList(contestReq models.ContestRequest) (*models.ContestResponse, error) {
	return s.gameService.HandleContestList(contestReq)
}

func (s *SocketService) HandleContestJoin(joinReq models.ContestJoinRequest) (*models.ContestJoinResponse, error) {
	return s.gameService.HandleContestJoin(joinReq)
}

func (s *SocketService) HandleListContestScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	return s.gameService.HandleListContestScreen(mainReq)
}

func (s *SocketService) HandleContestGap(gapReq models.ContestGapRequest) (*models.ContestGapResponse, error) {
	return s.gameService.HandleContestGap(gapReq)
}

// GetCassandraSession returns the Cassandra session for external access
func (s *SocketService) GetCassandraSession() *gocql.Session {
	return s.cassandraSession
}

// GetAuthService returns the auth service instance
func (s *SocketService) GetAuthService() *AuthService {
	return s.authService
}

// GetGameService returns the game service instance
func (s *SocketService) GetGameService() *GameService {
	return s.gameService
}
