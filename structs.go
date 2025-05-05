package main

type Transport struct{}

// WriteCounter struct moved to internal/downloader/types.go
// type WriteCounter struct {
// 	Total      int64
// 	TotalStr   string
// 	Downloaded int64
// 	Percentage int
// 	StartTime  int64
// }

// Config struct moved to internal/config/config.go
// type Config struct {
// 	Email           string
// 	Password        string
// 	Urls            []string
// 	Format          int
// 	OutPath         string
// 	VideoFormat     int
// 	WantRes         string
// 	Token           string
// 	UseFfmpegEnvVar bool
// 	FfmpegNameStr   string
// 	ForceVideo      bool
// 	SkipVideos		bool
// 	SkipChapters	bool
// }

type Args struct {
	Urls         []string `arg:"positional, required"`
	Format       int      `arg:"-f" default:"-1" help:"Track download format.\n\t\t\t 1 = 16-bit / 44.1 kHz ALAC\n\t\t\t 2 = 16-bit / 44.1 kHz FLAC\n\t\t\t 3 = 24-bit / 48 kHz MQA\n\t\t\t 4 = 360 Reality Audio / best available\n\t\t\t 5 = 150 Kbps AAC"`
	VideoFormat  int      `arg:"-F" default:"-1" help:"Video download format.\n\t\t\t 1 = 480p\n\t\t\t 2 = 720p\n\t\t\t 3 = 1080p\n\t\t\t 4 = 1440p\n\t\t\t 5 = 4K / best available"`
	OutPath      string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
	ForceVideo   bool     `arg:"--force-video" help:"Forces video when it co-exists with audio in release URLs."`
	SkipVideos   bool     `arg:"--skip-videos" help:"Skips videos in artist URLs."`
	SkipChapters bool     `arg:"--skip-chapters" help:"Skips chapters for videos."`
}

// Auth struct moved to internal/downloader/types.go
// type Auth struct {
// 	AccessToken  string `json:"access_token"`
// 	ExpiresIn    int    `json:"expires_in"`
// 	TokenType    string `json:"token_type"`
// 	RefreshToken string `json:"refresh_token"`
// 	Scope        string `json:"scope"`
// }

// Payload struct moved to internal/downloader/types.go
// type Payload struct {
// 	Nbf         int      `json:"nbf"`
// 	Exp         int      `json:"exp"`
// 	Iss         string   `json:"iss"`
// 	Aud         []string `json:"aud"`
// 	ClientID    string   `json:"client_id"`
// 	Sub         string   `json:"sub"`
// 	AuthTime    int      `json:"auth_time"`
// 	Idp         string   `json:"idp"`
// 	Email       string   `json:"email"`
// 	LegacyToken string   `json:"legacy_token"`
// 	LegacyUguid string   `json:"legacy_uguid"`
// 	Jti         string   `json:"jti"`
// 	Sid         string   `json:"sid"`
// 	Iat         int      `json:"iat"`
// 	Scope       []string `json:"scope"`
// 	Amr         []string `json:"amr"`
// }

// UserInfo struct moved to internal/downloader/types.go
// type UserInfo struct {
// 	Sub               string `json:"sub"`
// 	PreferredUsername string `json:"preferred_username"`
// 	Name              string `json:"name"`
// 	Email             string `json:"email"`
// 	EmailVerified     bool   `json:"email_verified"`
// }

// SubInfo struct moved to internal/downloader/types.go
// type SubInfo struct {
// 	StripeMetaData struct {
// 		SubscriptionID      string      `json:"subscriptionId"`
// 		InvoiceID           string      `json:"invoiceId"`
// 		PaymentIntentStatus interface{} `json:"paymentIntentStatus"`
// 		ReturnURL           interface{} `json:"returnUrl"`
// 		RedirectURL         interface{} `json:"redirectUrl"`
// 		PaymentError        interface{} `json:"paymentError"`
// 	} `json:"stripeMetaData"`
// 	IsTrialAvailable        bool   `json:"isTrialAvailable"`
// 	AllowAddNewSubscription bool   `json:"allowAddNewSubscription"`
// 	ID                      string `json:"id"`
// 	LegacySubscriptionID    string `json:"legacySubscriptionId"`
// 	Status                  string `json:"status"`
// 	IsContentAccessible     bool   `json:"isContentAccessible"`
// 	StartedAt               string `json:"startedAt"`
// 	EndsAt                  string `json:"endsAt"`
// 	TrialEndsAt             string `json:"trialEndsAt"`
// 	Plan                    struct {
// 		ID              string      `json:"id"`
// 		Price           float64     `json:"price"`
// 		Period          int         `json:"period"`
// 		TrialPeriodDays int         `json:"trialPeriodDays"`
// 		PlanID          string      `json:"planId"`
// 		Description     string      `json:"description"`
// 		ServiceLevel    string      `json:"serviceLevel"`
// 		StartsAt        interface{} `json:"startsAt"`
// 		EndsAt          interface{} `json:"endsAt"`
// 	} `json:"plan"`
// 	Promo struct {
// 		ID            string      `json:"id"`
// 		PromoCode     string      `json:"promoCode"`
// 		PromoPrice    float64     `json:"promoPrice"`
// 		Description   string      `json:"description"`
// 		PromoStartsAt interface{} `json:"promoStartsAt"`
// 		PromoEndsAt   interface{} `json:"promoEndsAt"`
// 		Plan          struct {
// 			ID              string      `json:"id"`
// 			Price           float64     `json:"price"`
// 			Period          int         `json:"period"`
// 			TrialPeriodDays int         `json:"trialPeriodDays"`
// 			PlanID          string      `json:"planId"`
// 			Description     string      `json:"description"`
// 			ServiceLevel    string      `json:"serviceLevel"`
// 			StartsAt        interface{} `json:"startsAt"`
// 			EndsAt          interface{} `json:"endsAt"`
// 		} `json:"plan"`
// 		Gateway string `json:"gateway"`
// 	}
// }

// StreamParams struct moved to internal/downloader/types.go
// type StreamParams struct {
// 	SubscriptionID          string
// 	SubCostplanIDAccessList string
// 	UserID                  string
// 	StartStamp              string
// 	EndStamp                string
// }

// Product struct moved to internal/downloader/types.go
// type Product struct { ... }

// ProductFormatList struct moved to internal/downloader/types.go
// type ProductFormatList struct { ... }

// Track struct moved to internal/downloader/types.go
// type Track struct { ... }

// AlbArtResp struct moved to internal/downloader/types.go
// type AlbArtResp struct { ... }

// AlbumMeta struct moved to internal/downloader/types.go
// type AlbumMeta struct { ... }

// PlistItem struct moved to internal/downloader/types.go
// type PlistItem struct { ... }

// PlistResp struct moved to internal/downloader/types.go
// type PlistResp struct { ... }

// PlistMeta struct moved to internal/downloader/types.go
// type PlistMeta struct { ... }

// ArtistResp struct moved to internal/downloader/types.go
// type ArtistResp struct { ... }

// ArtistMeta struct moved to internal/downloader/types.go
// type ArtistMeta struct { ... }

// PurchasedManResp struct moved to internal/downloader/types.go
// type PurchasedManResp struct { ... }

// StreamMeta struct moved to internal/downloader/types.go
// type StreamMeta struct { ... }

// Token struct moved to internal/downloader/types.go
// type Token struct { ... }

// Quality struct moved to internal/downloader/types.go
// type Quality struct { ... }
