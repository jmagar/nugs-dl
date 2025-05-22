package downloader

import (
	"fmt"
	"time"

	api "nugs-dl/pkg/api"
)

// Moved from main.go

// Quality represents the details of an available audio/video format.
type Quality struct {
	Specs     string
	Extension string
	URL       string
	Format    int // Corresponds to user input format codes (1-5 for audio, maybe extend for video?)
}

// qualityMap maps internal nugs stream URL parts to Quality details.
var qualityMap = map[string]Quality{
	".alac16/": {Specs: "16-bit / 44.1 kHz ALAC", Extension: ".m4a", Format: 1},
	".flac16/": {Specs: "16-bit / 44.1 kHz FLAC", Extension: ".flac", Format: 2},
	".mqa24/":  {Specs: "24-bit / 48 kHz MQA", Extension: ".flac", Format: 3},
	".flac?":   {Specs: "FLAC", Extension: ".flac", Format: 2}, // Generic FLAC fallback?
	".s360/":   {Specs: "360 Reality Audio", Extension: ".mp4", Format: 4},
	".aac150/": {Specs: "150 Kbps AAC", Extension: ".m4a", Format: 5},
	".m4a?":    {Specs: "AAC", Extension: ".m4a", Format: 5}, // Generic AAC fallback?
	".m3u8?":   {Extension: ".m4a", Format: 6},               // Special case for HLS audio
}

// trackFallback defines the quality fallback order if the desired format isn't available.
var trackFallback = map[int]int{
	1: 2, // ALAC -> FLAC
	2: 5, // FLAC -> AAC
	3: 2, // MQA -> FLAC
	4: 3, // 360 -> MQA (or should it be FLAC?)
	// 5 (AAC) has no defined fallback
}

// resolveRes maps video format codes (1-5) to resolution strings.
var resolveRes = map[int]string{
	1: "480",
	2: "720",
	3: "1080",
	4: "1440",
	5: "2160", // Represents 4K
}

// resFallback defines the resolution fallback order if the desired one isn't available.
var resFallback = map[string]string{
	"720":  "480",
	"1080": "720",
	"1440": "1080",
	// 4K (2160) doesn't have a defined fallback in original code, assumes highest available?
}

// --- Structs moved from main/structs.go ---

// Auth represents the response from the nugs token endpoint.
type Auth struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// Payload represents the decoded JWT payload containing legacy tokens.
type Payload struct {
	Nbf         int      `json:"nbf"`
	Exp         int      `json:"exp"`
	Iss         string   `json:"iss"`
	Aud         []string `json:"aud"`
	ClientID    string   `json:"client_id"`
	Sub         string   `json:"sub"` // User ID
	AuthTime    int      `json:"auth_time"`
	Idp         string   `json:"idp"`
	Email       string   `json:"email"`
	LegacyToken string   `json:"legacy_token"`
	LegacyUguid string   `json:"legacy_uguid"`
	Jti         string   `json:"jti"`
	Sid         string   `json:"sid"`
	Iat         int      `json:"iat"`
	Scope       []string `json:"scope"`
	Amr         []string `json:"amr"`
}

// UserInfo contains basic user details from the userinfo endpoint.
type UserInfo struct {
	Sub               string `json:"sub"` // This is the User ID needed for stream params
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
}

// SubInfo contains subscription details.
// Note: This is a complex struct, including Plan and Promo sub-structs.
type SubInfo struct {
	// ... (Keep the full struct definition from structs.go) ...
	StripeMetaData struct {
		SubscriptionID      string      `json:"subscriptionId"`
		InvoiceID           string      `json:"invoiceId"`
		PaymentIntentStatus interface{} `json:"paymentIntentStatus"`
		ReturnURL           interface{} `json:"returnUrl"`
		RedirectURL         interface{} `json:"redirectUrl"`
		PaymentError        interface{} `json:"paymentError"`
	} `json:"stripeMetaData"`
	IsTrialAvailable        bool   `json:"isTrialAvailable"`
	AllowAddNewSubscription bool   `json:"allowAddNewSubscription"`
	ID                      string `json:"id"`
	LegacySubscriptionID    string `json:"legacySubscriptionId"`
	Status                  string `json:"status"`
	IsContentAccessible     bool   `json:"isContentAccessible"`
	StartedAt               string `json:"startedAt"`
	EndsAt                  string `json:"endsAt"`
	TrialEndsAt             string `json:"trialEndsAt"`
	Plan                    struct {
		ID              string      `json:"id"`
		Price           float64     `json:"price"`
		Period          int         `json:"period"`
		TrialPeriodDays int         `json:"trialPeriodDays"`
		PlanID          string      `json:"planId"` // Needed for stream params
		Description     string      `json:"description"`
		ServiceLevel    string      `json:"serviceLevel"`
		StartsAt        interface{} `json:"startsAt"`
		EndsAt          interface{} `json:"endsAt"`
	} `json:"plan"`
	Promo struct {
		// ... Promo details ...
		ID            string      `json:"id"`
		PromoCode     string      `json:"promoCode"`
		PromoPrice    float64     `json:"promoPrice"`
		Description   string      `json:"description"`
		PromoStartsAt interface{} `json:"promoStartsAt"`
		PromoEndsAt   interface{} `json:"promoEndsAt"`
		Plan          struct {
			ID              string      `json:"id"`
			Price           float64     `json:"price"`
			Period          int         `json:"period"`
			TrialPeriodDays int         `json:"trialPeriodDays"`
			PlanID          string      `json:"planId"` // Needed for stream params if promo active
			Description     string      `json:"description"`
			ServiceLevel    string      `json:"serviceLevel"`
			StartsAt        interface{} `json:"startsAt"`
			EndsAt          interface{} `json:"endsAt"`
		} `json:"plan"`
		Gateway string `json:"gateway"`
	} `json:"promo"` // Added closing brace if missing in snippet
}

// StreamParams holds parameters needed for the stream metadata API call.
type StreamParams struct {
	SubscriptionID          string
	SubCostplanIDAccessList string
	UserID                  string
	StartStamp              string
	EndStamp                string
}

// --- Progress Reporting ---

// WriteCounter counts bytes written and calculates download progress.
type WriteCounter struct {
	JobID          string
	Total          int64
	TotalStr       string
	Downloaded     int64
	ProgressChan   chan<- api.ProgressUpdate // Use imported type
	StartTime      int64
	lastUpdateTime int64
}

const progressUpdateInterval = 500 // Milliseconds between progress updates

// Write implements the io.Writer interface for WriteCounter.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += int64(n)

	now := time.Now().UnixMilli()
	
	// Always send first progress update, then throttle subsequent ones
	shouldSendUpdate := wc.lastUpdateTime == 0 || // First update
		(wc.Total > 0 && now-wc.lastUpdateTime >= progressUpdateInterval) || // Regular throttled updates
		(wc.Total <= 0 && now-wc.lastUpdateTime >= progressUpdateInterval/2) // More frequent updates for unknown size
		
	if !shouldSendUpdate {
		return n, nil // Not time to update yet
	}
	wc.lastUpdateTime = now

	var speedBps int64 = 0
	var percentage float64 = 0.0

	if wc.Total > 0 {
		percentage = float64(wc.Downloaded) / float64(wc.Total) * 100.0
	} else {
		// For unknown size, show progress as bytes downloaded instead of percentage
		// Frontend will handle this appropriately
		percentage = -1 // Indicate unknown percentage
	}

	elapsed := now - wc.StartTime
	if elapsed > 500 && wc.Downloaded > 0 { // Avoid calculating speed too early
		speedBps = (wc.Downloaded * 1000) / elapsed
	}

	// Send update over the channel
	if wc.ProgressChan != nil {
		update := api.ProgressUpdate{ // Use imported type
			JobID:           wc.JobID,
			Percentage:      percentage,
			BytesDownloaded: wc.Downloaded,
			TotalBytes:      wc.Total,
			SpeedBPS:        speedBps,
			// Status, Message, CurrentFile are set by the calling function via sendProgress
		}
		
		fmt.Printf("[WriteCounter] Job %s: Downloaded %d/%d bytes (%.1f%%), Speed: %d B/s\n", 
			wc.JobID, wc.Downloaded, wc.Total, percentage, speedBps)
		
		// Use non-blocking send
		select {
		case wc.ProgressChan <- update:
			fmt.Printf("[WriteCounter] Progress update sent for Job %s\n", wc.JobID)
		default:
			fmt.Printf("[WriteCounter Warning] Progress channel full for Job %s, discarding update.\n", wc.JobID)
		}
	} else {
		fmt.Printf("[WriteCounter] Progress channel is nil for Job %s\n", wc.JobID)
	}

	return n, nil
}

// --- Metadata Structs ---

// Product represents a purchasable format or item.
type Product struct {
	ProductStatusType    int           `json:"productStatusType"`
	SkuIDExt             interface{}   `json:"skuIDExt"`
	FormatStr            string        `json:"formatStr"`
	SkuID                int           `json:"skuID"` // Important for getting streams/manifests
	Cost                 int           `json:"cost"`
	CostplanID           int           `json:"costplanID"`
	Pricing              interface{}   `json:"pricing"`
	Bundles              []interface{} `json:"bundles"`
	NumPublicPricePoints int           `json:"numPublicPricePoints"`
	CartLink             string        `json:"cartLink"`
	LiveEventInfo        interface{}   `json:"liveEventInfo"`  // Simplified for brevity
	SaleWindowInfo       interface{}   `json:"saleWindowInfo"` // Simplified for brevity
	IosCost              int           `json:"iosCost"`
	IosPlanName          interface{}   `json:"iosPlanName"`
	GooglePlanName       interface{}   `json:"googlePlanName"`
	GoogleCost           int           `json:"googleCost"`
	NumDiscs             int           `json:"numDiscs"`
	IsSubStreamOnly      int           `json:"isSubStreamOnly"`
}

// ProductFormatList is used specifically in livestream metadata.
type ProductFormatList struct {
	PfType          int         `json:"pfType"`
	FormatStr       string      `json:"formatStr"`
	SkuID           int         `json:"skuID"` // Important for getting streams/manifests
	Cost            int         `json:"cost"`
	CostplanID      int         `json:"costplanID"`
	PfTypeStr       string      `json:"pfTypeStr"`
	LiveEvent       interface{} `json:"liveEvent"`  // Simplified
	Salewindow      interface{} `json:"salewindow"` // Simplified
	SkuCode         string      `json:"skuCode"`
	IsSubStreamOnly int         `json:"isSubStreamOnly"`
}

// Track represents a single audio track's metadata.
type Track struct {
	TrackLabel       string `json:"trackLabel"`
	TrackURL         string `json:"trackURL"`
	SongID           int    `json:"songID"`
	SongTitle        string `json:"songTitle"`
	TotalRunningTime int    `json:"totalRunningTime"`
	DiscNum          int    `json:"discNum"`
	TrackNum         int    `json:"trackNum"`
	SetNum           int    `json:"setNum"`
	ClipURL          string `json:"clipURL"`
	TrackID          int    `json:"trackID"` // Important for getting streams
	TrackExclude     int    `json:"trackExclude"`
	// Simplified other fields
	Products []Product `json:"products"`
}

// AlbArtResp holds the main response data for album/artist/video metadata.
type AlbArtResp struct {
	NumReviews                int                  `json:"numReviews"`
	TotalContainerRunningTime int                  `json:"totalContainerRunningTime"`
	Products                  []Product            `json:"products"`
	Tracks                    []Track              `json:"tracks"`
	Songs                     []Track              `json:"songs"`
	ContainerID               int                  `json:"containerID"`
	ContainerInfo             string               `json:"containerInfo"`
	ArtistName                string               `json:"artistName"`
	AvailabilityTypeStr       string               `json:"availabilityTypeStr"`
	ContainerTypeStr          string               `json:"containerTypeStr"`
	ProductFormatList         []*ProductFormatList `json:"productFormatList"`
	VideoChapters             []interface{}        `json:"videoChapters"`
	// Artwork related fields (added back from original structs.go)
	VodPlayerImage  string      `json:"vodPlayerImage"`
	CoverImage      interface{} `json:"coverImage"` // Can be string or null?
	Img             ImageInfo   `json:"img"`        // Assuming ImageInfo struct exists or defined below
	Pics            []ImageInfo `json:"pics"`
	VenueName       string      `json:"venueName"`       // Added for context
	PerformanceDate string      `json:"performanceDate"` // Added for context
	// Simplified other fields
}

// ImageInfo holds details for an image URL (used in Img and Pics).
// Need to define this based on original structs.go if not already present.
type ImageInfo struct {
	PicID   int    `json:"picID"`
	OrderID int    `json:"orderID"`
	Height  int    `json:"height"`
	Width   int    `json:"width"`
	Caption string `json:"caption"`
	URL     string `json:"url"`
}

// AlbumMeta is the top-level structure for album/video metadata responses.
type AlbumMeta struct {
	MethodName                  string      `json:"methodName"`
	ResponseAvailabilityCode    int         `json:"responseAvailabilityCode"`
	ResponseAvailabilityCodeStr string      `json:"responseAvailabilityCodeStr"`
	Response                    *AlbArtResp `json:"Response"`
}

// PlistItem represents a single item within a playlist.
type PlistItem struct {
	ID                int         `json:"ID"`
	OrderID           int         `json:"orderID"`
	Track             Track       `json:"track"`
	PlaylistContainer interface{} `json:"playlistContainer"` // Simplified
}

// PlistResp holds the main response data for playlist metadata.
type PlistResp struct {
	TotalRunningTime int         `json:"totalRunningTime"`
	ID               int         `json:"ID"`
	UserID           int         `json:"userID"`
	Items            []PlistItem `json:"items"`
	PlayListName     string      `json:"playListName"`
	// Simplified other fields
}

// PlistMeta is the top-level structure for playlist metadata responses.
type PlistMeta struct {
	MethodName string     `json:"methodName"`
	Response   *PlistResp `json:"Response"`
	// Simplified other fields
}

// ArtistResp holds the main response data for artist metadata.
type ArtistResp struct {
	Containers          []*AlbArtResp `json:"containers"`
	TotalMatchedRecords int           `json:"totalMatchedRecords"`
	// Simplified other fields
}

// ArtistMeta is the top-level structure for artist metadata responses.
type ArtistMeta struct {
	MethodName                  string      `json:"methodName"`
	ResponseAvailabilityCode    int         `json:"responseAvailabilityCode"`
	ResponseAvailabilityCodeStr string      `json:"responseAvailabilityCodeStr"`
	Response                    *ArtistResp `json:"Response"`
}

// PurchasedManResp is the response structure for purchased video manifest URLs.
type PurchasedManResp struct {
	FileURL      string `json:"fileURL"`
	ResponseCode int    `json:"responseCode"`
}

// StreamMeta is the response structure for track/video stream URLs.
type StreamMeta struct {
	StreamLink         string      `json:"streamLink"`
	Streamer           string      `json:"streamer"`
	UserID             string      `json:"userID"`
	Mason              interface{} `json:"mason"`
	SubContentAccess   int         `json:"subContentAccess"`
	StashContentAccess int         `json:"stashContentAccess"`
}

// Token response - seems only used by getPlistMeta (secureApi path), maybe move later?
// Keeping here for now as it relates to metadata fetching.
type Token struct {
	MethodName string `json:"methodName"`
	Response   struct {
		TokenValue     string      `json:"tokenValue"`
		ReturnCode     int         `json:"returnCode"`
		ReturnCodeStr  string      `json:"returnCodeStr"`
		NnCustomerAuth interface{} `json:"nnCustomerAuth"`
	} `json:"Response"`
	// Simplified other fields
}

// Placeholder for other structs to be moved later
