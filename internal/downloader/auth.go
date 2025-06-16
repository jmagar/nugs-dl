package downloader

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nugs-dl/internal/logger" // Import the logger package
	// Import config locally if needed, or expect it via Downloader struct
	// appConfig "main/internal/config"
)

// Constants moved from main.go related to auth/user info
const (
	clientId    = "Eg7HuH873H65r5rt325UytR5429"
	layout      = "01/02/2006 15:04:05"                                                // Used in parseTimestamps
	userAgent   = "NugsNet/3.26.724 (Android; 7.1.2; Asus; ASUS_Z01QD; Scale/2.0; en)" // Used in auth, getUserInfo, getSubInfo
	authUrl     = "https://id.nugs.net/connect/token"
	subInfoUrl  = "https://subscriptions.nugs.net/api/v1/me/subscriptions"
	userInfoUrl = "https://id.nugs.net/connect/userinfo"
)

// Authenticate performs email/password authentication.
// It takes the Downloader context to access the HTTP client.
func (d *Downloader) Authenticate(email, pwd string) (string, error) {
	logger.Info("Attempting authentication...", "email", email) // Log email for context, be mindful of PII if logs are public
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("grant_type", "password")
	data.Set("scope", "openid profile email nugsnet:api nugsnet:legacyapi offline_access")
	data.Set("username", email)
	data.Set("password", pwd)
	req, err := http.NewRequest(http.MethodPost, authUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform auth request: %w", err)
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed: %s", do.Status)
	}
	var obj Auth
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}
	logger.Info("Authentication successful.", "email", email)
	return obj.AccessToken, nil
}

// GetUserInfo retrieves user details using the access token.
func (d *Downloader) GetUserInfo(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, userInfoUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create user info request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("User-Agent", userAgent)
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform user info request: %w", err)
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get user info failed: %s", do.Status)
	}
	var obj UserInfo
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", fmt.Errorf("failed to decode user info response: %w", err)
	}
	return obj.Sub, nil // Return the User ID (sub)
}

// GetSubInfo retrieves subscription details using the access token.
func (d *Downloader) GetSubInfo(token string) (*SubInfo, error) {
	req, err := http.NewRequest(http.MethodGet, subInfoUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub info request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("User-Agent", userAgent)
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform sub info request: %w", err)
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		// Handle potential expired subscription errors differently?
		return nil, fmt.Errorf("get sub info failed: %s", do.Status)
	}
	var obj SubInfo
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode sub info response: %w", err)
	}
	return &obj, nil
}

// getPlan determines the active plan description and if it's a promo.
// Made internal as it's only used by parseStreamParams now.
func getPlan(subInfo *SubInfo) (string, bool) {
	// Use reflect.ValueOf(subInfo.Plan).IsZero() check for safety
	if subInfo != nil && subInfo.Plan.PlanID != "" { // Check if Plan itself has a PlanID
		return subInfo.Plan.Description, false
	} else if subInfo != nil && subInfo.Promo.Plan.PlanID != "" { // Check promo plan
		return subInfo.Promo.Plan.Description, true
	}
	return "Unknown Plan", false // Fallback case
}

// parseTimestamps converts start/end date strings to Unix timestamps.
// Kept internal as it's only used by parseStreamParams.
func parseTimestamps(start, end string) (string, string) {
	startTime, errStart := time.Parse(layout, start)
	endTime, errEnd := time.Parse(layout, end)
	// Basic error check, return "0" maybe?
	if errStart != nil || errEnd != nil {
		logger.Warn("Could not parse timestamps from subscription info", "startString", start, "endString", end, "startError", errStart, "endError", errEnd)
		return "0", "0"
	}
	parsedStart := strconv.FormatInt(startTime.Unix(), 10)
	parsedEnd := strconv.FormatInt(endTime.Unix(), 10)
	return parsedStart, parsedEnd
}

// ParseStreamParams creates the StreamParams struct needed for download API calls.
func ParseStreamParams(userId string, subInfo *SubInfo) (*StreamParams, error) {
	if userId == "" || subInfo == nil {
		return nil, errors.New("missing user ID or subscription info for stream parameters")
	}

	_, isPromo := getPlan(subInfo)
	startStamp, endStamp := parseTimestamps(subInfo.StartedAt, subInfo.EndsAt)

	streamParams := &StreamParams{
		SubscriptionID: subInfo.LegacySubscriptionID,
		UserID:         userId,
		StartStamp:     startStamp,
		EndStamp:       endStamp,
	}

	if isPromo && subInfo.Promo.Plan.PlanID != "" {
		streamParams.SubCostplanIDAccessList = subInfo.Promo.Plan.PlanID
	} else if !isPromo && subInfo.Plan.PlanID != "" {
		streamParams.SubCostplanIDAccessList = subInfo.Plan.PlanID
	} else {
		// Handle case where neither plan ID is available? Return error?
		return nil, errors.New("could not determine active plan ID from subscription info")
	}

	return streamParams, nil
}

// ExtractLegacyTokens extracts legacy tokens from the main access token (JWT).
func ExtractLegacyTokens(tokenStr string) (legacyToken string, legacyUguid string, err error) {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) < 2 {
		return "", "", errors.New("invalid JWT token format")
	}
	payload := parts[1]
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode token payload: %w", err)
	}
	var obj Payload
	err = json.Unmarshal(decoded, &obj)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal token payload: %w", err)
	}
	return obj.LegacyToken, obj.LegacyUguid, nil
}
