package service

import (
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// IRiskDataClient defines the interface for risk-related data operations (dependency inversion)
type IRiskDataClient interface {
	GetUserRiskProfile(userID string) (*UserRiskProfile, error)
	GetPositionExposure(userID, symbol string) (*PositionExposure, error)
	GetAccountBalance(userID string) (*AccountBalance, error)
	GetMarketVolatility(symbol string) (*MarketVolatility, error)
	GetUserTradingLimits(userID string) (*TradingLimits, error)
}

// UserRiskProfile represents user's risk tolerance and profile
type UserRiskProfile struct {
	UserID               string
	RiskTolerance        RiskTolerance
	MaxPositionSize      float64
	MaxDailyTradingValue float64
	MaxOrderValue        float64
	IsHighRiskApproved   bool
	ProfileLastUpdated   time.Time
}

// RiskTolerance represents risk tolerance levels
type RiskTolerance int32

const (
	RiskToleranceConservative RiskTolerance = iota
	RiskToleranceModerate
	RiskToleranceAggressive
	RiskToleranceSpeculative
)

// PositionExposure represents current position exposure for a symbol
type PositionExposure struct {
	Symbol          string
	CurrentQuantity float64
	CurrentValue    float64
	AveragePrice    float64
	UnrealizedPnL   float64
	ExposurePercent float64
}

// AccountBalance represents account balance information
type AccountBalance struct {
	TotalBalance     float64
	AvailableBalance float64
	BuyingPower      float64
	LastUpdated      time.Time
}

// MarketVolatility represents market volatility metrics
type MarketVolatility struct {
	Symbol           string
	Volatility30Day  float64
	Beta             float64
	RiskRating       string
	IsHighVolatility bool
	LastCalculated   time.Time
}

// TradingLimits represents user trading limits
type TradingLimits struct {
	DailyTradingLimit   float64
	DailyTradingUsed    float64
	MaxOrderValue       float64
	MaxPositionSize     float64
	RemainingDailyLimit float64
}

// RiskAssessment represents the result of risk assessment
type RiskAssessment struct {
	RiskLevel        RiskLevel
	RiskScore        float64
	IsApproved       bool
	RequiresApproval bool
	RiskFactors      []RiskFactor
	Recommendations  []string
	Warnings         []string
	AssessmentTime   time.Time
}

// RiskLevel represents overall risk level
type RiskLevel int32

const (
	RiskLevelLow RiskLevel = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelVeryHigh
	RiskLevelExtremelyHigh
)

// RiskFactor represents individual risk factors
type RiskFactor struct {
	Factor      string
	Impact      RiskImpact
	Score       float64
	Description string
}

// RiskImpact represents the impact level of a risk factor
type RiskImpact int32

const (
	RiskImpactLow RiskImpact = iota
	RiskImpactMedium
	RiskImpactHigh
	RiskImpactCritical
)

// RiskManagementService handles risk assessment and management business logic
type RiskManagementService interface {
	// AssessOrderRisk performs comprehensive risk assessment for an order
	AssessOrderRisk(order *domain.Order, riskDataClient IRiskDataClient) (*RiskAssessment, error)

	// ValidateRiskLimits validates order against user risk limits
	ValidateRiskLimits(order *domain.Order, riskDataClient IRiskDataClient) error

	// CheckPositionLimits validates position size limits
	CheckPositionLimits(order *domain.Order, riskDataClient IRiskDataClient) error

	// CheckTradingLimits validates trading value limits
	CheckTradingLimits(order *domain.Order, riskDataClient IRiskDataClient) error

	// AssessMarketRisk evaluates market-related risks
	AssessMarketRisk(order *domain.Order, riskDataClient IRiskDataClient) (*RiskAssessment, error)

	// AssessConcentrationRisk evaluates portfolio concentration risk
	AssessConcentrationRisk(order *domain.Order, riskDataClient IRiskDataClient) (*RiskAssessment, error)

	// RequiresManualApproval determines if order needs manual approval
	RequiresManualApproval(riskAssessment *RiskAssessment) bool

	// CalculateRiskScore calculates overall risk score for an order
	CalculateRiskScore(order *domain.Order, riskDataClient IRiskDataClient) (float64, error)
}

type riskManagementService struct {
	// Configuration for risk management
	maxRiskScore            float64
	highRiskThreshold       float64
	concentrationLimit      float64
	volatilityThreshold     float64
	manualApprovalThreshold float64
}

// RiskManagementConfig holds configuration for risk management
type RiskManagementConfig struct {
	MaxRiskScore            float64 // Maximum allowed risk score (0-100)
	HighRiskThreshold       float64 // Threshold for high risk classification
	ConcentrationLimit      float64 // Maximum concentration percentage
	VolatilityThreshold     float64 // Volatility threshold for high risk
	ManualApprovalThreshold float64 // Threshold requiring manual approval
}

// NewRiskManagementService creates a new instance of RiskManagementService
func NewRiskManagementService(config RiskManagementConfig) RiskManagementService {
	return &riskManagementService{
		maxRiskScore:            config.MaxRiskScore,
		highRiskThreshold:       config.HighRiskThreshold,
		concentrationLimit:      config.ConcentrationLimit,
		volatilityThreshold:     config.VolatilityThreshold,
		manualApprovalThreshold: config.ManualApprovalThreshold,
	}
}

// NewRiskManagementServiceWithDefaults creates a service with default configuration
func NewRiskManagementServiceWithDefaults() RiskManagementService {
	return NewRiskManagementService(RiskManagementConfig{
		MaxRiskScore:            80.0, // Max risk score of 80
		HighRiskThreshold:       60.0, // High risk at 60+
		ConcentrationLimit:      20.0, // Max 20% concentration in single position
		VolatilityThreshold:     25.0, // High volatility at 25%+
		ManualApprovalThreshold: 70.0, // Manual approval at 70+ risk score
	})
}

// AssessOrderRisk performs comprehensive risk assessment for an order
func (s *riskManagementService) AssessOrderRisk(order *domain.Order, riskDataClient IRiskDataClient) (*RiskAssessment, error) {
	assessment := &RiskAssessment{
		RiskFactors:     make([]RiskFactor, 0),
		Recommendations: make([]string, 0),
		Warnings:        make([]string, 0),
		AssessmentTime:  time.Now(),
	}

	// Calculate overall risk score
	riskScore, err := s.CalculateRiskScore(order, riskDataClient)
	if err != nil {
		return assessment, fmt.Errorf("failed to calculate risk score: %w", err)
	}

	assessment.RiskScore = riskScore
	assessment.RiskLevel = s.determineRiskLevel(riskScore)

	// Perform individual risk assessments
	if err := s.assessUserRiskProfile(order, riskDataClient, assessment); err != nil {
		return assessment, err
	}

	if err := s.assessPositionRisk(order, riskDataClient, assessment); err != nil {
		return assessment, err
	}

	if err := s.assessMarketRiskFactors(order, riskDataClient, assessment); err != nil {
		return assessment, err
	}

	if err := s.assessTradingLimitsRisk(order, riskDataClient, assessment); err != nil {
		return assessment, err
	}

	// Determine approval status
	assessment.IsApproved = assessment.RiskScore <= s.maxRiskScore
	assessment.RequiresApproval = s.RequiresManualApproval(assessment)

	// Generate recommendations and warnings
	s.generateRiskRecommendations(assessment)

	return assessment, nil
}

// ValidateRiskLimits validates order against user risk limits
func (s *riskManagementService) ValidateRiskLimits(order *domain.Order, riskDataClient IRiskDataClient) error {
	// Check user risk profile
	userProfile, err := riskDataClient.GetUserRiskProfile(order.UserID())
	if err != nil {
		return fmt.Errorf("failed to get user risk profile: %w", err)
	}

	orderValue := order.CalculateOrderValue()

	// Check maximum order value
	if orderValue > userProfile.MaxOrderValue {
		return fmt.Errorf("order value %.2f exceeds user limit %.2f", orderValue, userProfile.MaxOrderValue)
	}

	// Check daily trading limit
	tradingLimits, err := riskDataClient.GetUserTradingLimits(order.UserID())
	if err != nil {
		return fmt.Errorf("failed to get trading limits: %w", err)
	}

	if orderValue > tradingLimits.RemainingDailyLimit {
		return fmt.Errorf("order value %.2f exceeds remaining daily limit %.2f", orderValue, tradingLimits.RemainingDailyLimit)
	}

	return nil
}

// CheckPositionLimits validates position size limits
func (s *riskManagementService) CheckPositionLimits(order *domain.Order, riskDataClient IRiskDataClient) error {
	if order.IsSellOrder() {
		return nil // Sell orders reduce position, no limit check needed
	}

	// Get current position
	currentPosition, err := riskDataClient.GetPositionExposure(order.UserID(), order.Symbol())
	if err != nil {
		return fmt.Errorf("failed to get position exposure: %w", err)
	}

	// Calculate new position value after order
	orderValue := order.CalculateOrderValue()
	newPositionValue := currentPosition.CurrentValue + orderValue

	// Get user risk profile for position limits
	userProfile, err := riskDataClient.GetUserRiskProfile(order.UserID())
	if err != nil {
		return fmt.Errorf("failed to get user risk profile: %w", err)
	}

	// Check maximum position size
	if newPositionValue > userProfile.MaxPositionSize {
		return fmt.Errorf("new position value %.2f would exceed maximum allowed %.2f", newPositionValue, userProfile.MaxPositionSize)
	}

	// Check concentration limits
	accountBalance, err := riskDataClient.GetAccountBalance(order.UserID())
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}

	concentrationPercent := (newPositionValue / accountBalance.TotalBalance) * 100
	if concentrationPercent > s.concentrationLimit {
		return fmt.Errorf("position concentration %.1f%% exceeds limit %.1f%%", concentrationPercent, s.concentrationLimit)
	}

	return nil
}

// CheckTradingLimits validates trading value limits
func (s *riskManagementService) CheckTradingLimits(order *domain.Order, riskDataClient IRiskDataClient) error {
	tradingLimits, err := riskDataClient.GetUserTradingLimits(order.UserID())
	if err != nil {
		return fmt.Errorf("failed to get trading limits: %w", err)
	}

	orderValue := order.CalculateOrderValue()

	// Check daily trading limit
	if orderValue > tradingLimits.RemainingDailyLimit {
		return fmt.Errorf("order value %.2f exceeds remaining daily limit %.2f", orderValue, tradingLimits.RemainingDailyLimit)
	}

	// Check maximum order value
	if orderValue > tradingLimits.MaxOrderValue {
		return fmt.Errorf("order value %.2f exceeds maximum order limit %.2f", orderValue, tradingLimits.MaxOrderValue)
	}

	return nil
}

// AssessMarketRisk evaluates market-related risks
func (s *riskManagementService) AssessMarketRisk(order *domain.Order, riskDataClient IRiskDataClient) (*RiskAssessment, error) {
	assessment := &RiskAssessment{
		RiskFactors:     make([]RiskFactor, 0),
		Recommendations: make([]string, 0),
		Warnings:        make([]string, 0),
		AssessmentTime:  time.Now(),
	}

	// Get market volatility data
	volatility, err := riskDataClient.GetMarketVolatility(order.Symbol())
	if err != nil {
		return assessment, fmt.Errorf("failed to get market volatility: %w", err)
	}

	// Assess volatility risk
	if volatility.IsHighVolatility {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "High Market Volatility",
			Impact:      RiskImpactHigh,
			Score:       volatility.Volatility30Day,
			Description: fmt.Sprintf("Symbol %s has high volatility (%.1f%%)", order.Symbol(), volatility.Volatility30Day),
		})
	}

	// Assess beta risk
	if volatility.Beta > 1.5 {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "High Beta",
			Impact:      RiskImpactMedium,
			Score:       volatility.Beta * 10, // Scale beta for scoring
			Description: fmt.Sprintf("Symbol %s has high beta (%.2f)", order.Symbol(), volatility.Beta),
		})
	}

	// Calculate market risk score
	marketRiskScore := s.calculateMarketRiskScore(volatility)
	assessment.RiskScore = marketRiskScore
	assessment.RiskLevel = s.determineRiskLevel(marketRiskScore)

	return assessment, nil
}

// AssessConcentrationRisk evaluates portfolio concentration risk
func (s *riskManagementService) AssessConcentrationRisk(order *domain.Order, riskDataClient IRiskDataClient) (*RiskAssessment, error) {
	assessment := &RiskAssessment{
		RiskFactors:     make([]RiskFactor, 0),
		Recommendations: make([]string, 0),
		Warnings:        make([]string, 0),
		AssessmentTime:  time.Now(),
	}

	// Skip concentration check for sell orders (they reduce concentration)
	if order.IsSellOrder() {
		assessment.RiskLevel = RiskLevelLow
		assessment.RiskScore = 0
		return assessment, nil
	}

	// Get current position and account balance
	currentPosition, err := riskDataClient.GetPositionExposure(order.UserID(), order.Symbol())
	if err != nil {
		return assessment, fmt.Errorf("failed to get position exposure: %w", err)
	}

	accountBalance, err := riskDataClient.GetAccountBalance(order.UserID())
	if err != nil {
		return assessment, fmt.Errorf("failed to get account balance: %w", err)
	}

	// Calculate concentration after order
	orderValue := order.CalculateOrderValue()
	newPositionValue := currentPosition.CurrentValue + orderValue
	concentrationPercent := (newPositionValue / accountBalance.TotalBalance) * 100

	// Assess concentration risk
	if concentrationPercent > s.concentrationLimit {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "High Concentration",
			Impact:      RiskImpactHigh,
			Score:       concentrationPercent,
			Description: fmt.Sprintf("Position concentration would be %.1f%% (limit: %.1f%%)", concentrationPercent, s.concentrationLimit),
		})
	} else if concentrationPercent > s.concentrationLimit*0.8 {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "Moderate Concentration",
			Impact:      RiskImpactMedium,
			Score:       concentrationPercent * 0.7,
			Description: fmt.Sprintf("Position concentration would be %.1f%%", concentrationPercent),
		})
	}

	// Calculate concentration risk score
	concentrationRiskScore := s.calculateConcentrationRiskScore(concentrationPercent)
	assessment.RiskScore = concentrationRiskScore
	assessment.RiskLevel = s.determineRiskLevel(concentrationRiskScore)

	return assessment, nil
}

// RequiresManualApproval determines if order needs manual approval
func (s *riskManagementService) RequiresManualApproval(riskAssessment *RiskAssessment) bool {
	if riskAssessment.RiskScore >= s.manualApprovalThreshold {
		return true
	}

	// Check for critical risk factors
	for _, factor := range riskAssessment.RiskFactors {
		if factor.Impact == RiskImpactCritical {
			return true
		}
	}

	// Check for very high or extremely high risk levels
	if riskAssessment.RiskLevel >= RiskLevelVeryHigh {
		return true
	}

	return false
}

// CalculateRiskScore calculates overall risk score for an order
func (s *riskManagementService) CalculateRiskScore(order *domain.Order, riskDataClient IRiskDataClient) (float64, error) {
	var totalScore float64
	var scoreComponents int

	// Market risk component
	marketRisk, err := s.AssessMarketRisk(order, riskDataClient)
	if err == nil {
		totalScore += marketRisk.RiskScore * 0.4 // 40% weight
		scoreComponents++
	}

	// Concentration risk component
	concentrationRisk, err := s.AssessConcentrationRisk(order, riskDataClient)
	if err == nil {
		totalScore += concentrationRisk.RiskScore * 0.3 // 30% weight
		scoreComponents++
	}

	// User risk profile component
	userRiskScore, err := s.calculateUserRiskScore(order, riskDataClient)
	if err == nil {
		totalScore += userRiskScore * 0.2 // 20% weight
		scoreComponents++
	}

	// Order size risk component
	orderSizeScore := s.calculateOrderSizeRiskScore(order)
	totalScore += orderSizeScore * 0.1 // 10% weight
	scoreComponents++

	if scoreComponents == 0 {
		return 0, fmt.Errorf("unable to calculate risk score: no components available")
	}

	return totalScore, nil
}

// Helper methods

func (s *riskManagementService) determineRiskLevel(riskScore float64) RiskLevel {
	switch {
	case riskScore >= 80:
		return RiskLevelExtremelyHigh
	case riskScore >= s.highRiskThreshold:
		return RiskLevelVeryHigh
	case riskScore >= 40:
		return RiskLevelHigh
	case riskScore >= 20:
		return RiskLevelMedium
	default:
		return RiskLevelLow
	}
}

func (s *riskManagementService) assessUserRiskProfile(order *domain.Order, riskDataClient IRiskDataClient, assessment *RiskAssessment) error {
	userProfile, err := riskDataClient.GetUserRiskProfile(order.UserID())
	if err != nil {
		return err
	}

	orderValue := order.CalculateOrderValue()

	// Check if order exceeds user's risk tolerance
	toleranceMultiplier := s.getRiskToleranceMultiplier(userProfile.RiskTolerance)
	if orderValue > userProfile.MaxOrderValue*toleranceMultiplier {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "Order Size vs Risk Tolerance",
			Impact:      RiskImpactHigh,
			Score:       (orderValue / userProfile.MaxOrderValue) * 20,
			Description: "Order size may exceed user's risk tolerance",
		})
	}

	return nil
}

func (s *riskManagementService) assessPositionRisk(order *domain.Order, riskDataClient IRiskDataClient, assessment *RiskAssessment) error {
	position, err := riskDataClient.GetPositionExposure(order.UserID(), order.Symbol())
	if err != nil {
		return err
	}

	// Check unrealized losses
	if position.UnrealizedPnL < 0 && abs(position.UnrealizedPnL) > position.CurrentValue*0.2 {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "Significant Unrealized Losses",
			Impact:      RiskImpactHigh,
			Score:       abs(position.UnrealizedPnL/position.CurrentValue) * 100,
			Description: fmt.Sprintf("Position has %.1f%% unrealized losses", abs(position.UnrealizedPnL/position.CurrentValue)*100),
		})
	}

	return nil
}

func (s *riskManagementService) assessMarketRiskFactors(order *domain.Order, riskDataClient IRiskDataClient, assessment *RiskAssessment) error {
	volatility, err := riskDataClient.GetMarketVolatility(order.Symbol())
	if err != nil {
		return err
	}

	if volatility.Volatility30Day > s.volatilityThreshold {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "High Volatility",
			Impact:      RiskImpactMedium,
			Score:       volatility.Volatility30Day,
			Description: fmt.Sprintf("Symbol has %.1f%% volatility", volatility.Volatility30Day),
		})
	}

	return nil
}

func (s *riskManagementService) assessTradingLimitsRisk(order *domain.Order, riskDataClient IRiskDataClient, assessment *RiskAssessment) error {
	limits, err := riskDataClient.GetUserTradingLimits(order.UserID())
	if err != nil {
		return err
	}

	orderValue := order.CalculateOrderValue()
	utilizationPercent := (orderValue / limits.RemainingDailyLimit) * 100

	if utilizationPercent > 80 {
		assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
			Factor:      "High Daily Limit Utilization",
			Impact:      RiskImpactMedium,
			Score:       utilizationPercent * 0.5,
			Description: fmt.Sprintf("Order uses %.1f%% of remaining daily limit", utilizationPercent),
		})
	}

	return nil
}

func (s *riskManagementService) generateRiskRecommendations(assessment *RiskAssessment) {
	switch assessment.RiskLevel {
	case RiskLevelExtremelyHigh:
		assessment.Recommendations = append(assessment.Recommendations,
			"Consider reducing order size significantly",
			"Seek investment advisor consultation",
			"Review risk tolerance settings")
	case RiskLevelVeryHigh:
		assessment.Recommendations = append(assessment.Recommendations,
			"Consider reducing order size",
			"Review position concentration",
			"Monitor market conditions closely")
	case RiskLevelHigh:
		assessment.Recommendations = append(assessment.Recommendations,
			"Consider partial position sizing",
			"Set stop-loss orders",
			"Monitor volatility")
	case RiskLevelMedium:
		assessment.Recommendations = append(assessment.Recommendations,
			"Consider position sizing strategies",
			"Monitor market conditions")
	}

	// Generate warnings based on risk factors
	for _, factor := range assessment.RiskFactors {
		if factor.Impact >= RiskImpactHigh {
			assessment.Warnings = append(assessment.Warnings, factor.Description)
		}
	}
}

func (s *riskManagementService) calculateMarketRiskScore(volatility *MarketVolatility) float64 {
	score := volatility.Volatility30Day

	if volatility.Beta > 1.5 {
		score += (volatility.Beta - 1.0) * 10
	}

	if volatility.IsHighVolatility {
		score += 15
	}

	return min(score, 100)
}

func (s *riskManagementService) calculateConcentrationRiskScore(concentrationPercent float64) float64 {
	if concentrationPercent <= s.concentrationLimit {
		return concentrationPercent * 2
	}

	// Exponential increase for concentrations above limit
	excess := concentrationPercent - s.concentrationLimit
	return s.concentrationLimit*2 + excess*4
}

func (s *riskManagementService) calculateUserRiskScore(order *domain.Order, riskDataClient IRiskDataClient) (float64, error) {
	userProfile, err := riskDataClient.GetUserRiskProfile(order.UserID())
	if err != nil {
		return 0, err
	}

	orderValue := order.CalculateOrderValue()
	utilizationPercent := (orderValue / userProfile.MaxOrderValue) * 100

	// Higher utilization = higher risk score
	return utilizationPercent * 0.8, nil
}

func (s *riskManagementService) calculateOrderSizeRiskScore(order *domain.Order) float64 {
	orderValue := order.CalculateOrderValue()

	// Simple order size risk based on order value
	switch {
	case orderValue >= 100000: // $100k+
		return 30
	case orderValue >= 50000: // $50k+
		return 20
	case orderValue >= 10000: // $10k+
		return 10
	default:
		return 5
	}
}

func (s *riskManagementService) getRiskToleranceMultiplier(tolerance RiskTolerance) float64 {
	switch tolerance {
	case RiskToleranceConservative:
		return 0.5
	case RiskToleranceModerate:
		return 0.8
	case RiskToleranceAggressive:
		return 1.2
	case RiskToleranceSpeculative:
		return 1.5
	default:
		return 1.0
	}
}

// Helper functions

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
