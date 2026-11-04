package config

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	yaml "gopkg.in/yaml.v2"
)

//go:embed config.yaml
var data []byte

var BuildVersion = "dev"

func UserAgent() string {
	return "mixin-suppergroup/" + BuildVersion
}

// SDKPrivateKey converts supported Ed25519 key encodings to the hex seed used by SDK v3.
func SDKPrivateKey(key string) string {
	if normalized, err := normalizePrivateKey(key); err == nil {
		return normalized
	}
	return key
}

func normalizePrivateKey(key string) (string, error) {
	// Check hex first: a hex seed is also valid unpadded base64 text.
	if decoded, err := hex.DecodeString(key); err == nil && (len(decoded) == ed25519.SeedSize || len(decoded) == ed25519.PrivateKeySize) {
		return hex.EncodeToString(decoded[:ed25519.SeedSize]), nil
	}
	for _, encoding := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.StdEncoding} {
		decoded, err := encoding.Strict().DecodeString(key)
		if err == nil && (len(decoded) == ed25519.SeedSize || len(decoded) == ed25519.PrivateKeySize) {
			return hex.EncodeToString(decoded[:ed25519.SeedSize]), nil
		}
	}
	return "", fmt.Errorf("expected an Ed25519 seed or private key encoded as hex or base64")
}

type PaymentAsset struct {
	Symbol  string `yaml:"symbol" json:"symbol"`
	AssetId string `yaml:"asset_id" json:"asset_id"`
	Amount  string `yaml:"amount" json:"amount"`
}

type Shortcut struct {
	Icon    string `yaml:"icon" json:"icon"`
	LabelEn string `yaml:"label_en" json:"label_en"`
	LabelZh string `yaml:"label_zh" json:"label_zh"`
	Url     string `yaml:"url" json:"url"`
}

type ShortcutGroup struct {
	LabelEn string     `yaml:"label_en" json:"label_en"`
	LabelZh string     `yaml:"label_zh" json:"label_zh"`
	Items   []Shortcut `yaml:"shortcuts" json:"shortcuts"`
}

type Config struct {
	Service struct {
		Name             string   `yaml:"name"`
		Environment      string   `yaml:"enviroment"`
		HTTPListenPort   int      `yaml:"port"`
		HTTPResourceHost string   `yaml:"host"`
		APIRoot          []string `yaml:"api_root"`
		BlazeRoot        []string `yaml:"blaze_root"`
		Retry            int      `yaml:"-"`
	} `yaml:"service"`
	Database struct {
		User     string `yaml:"username"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Name     string `yaml:"database_name"`
	} `yaml:"database"`
	System struct {
		MessageShardModifier   string          `yaml:"message_shard_modifier"`
		MessageShardSize       int64           `yaml:"message_shard_size"`
		PriceAssetsEnable      bool            `yaml:"price_asset_enable"`
		AudioMessageEnable     bool            `yaml:"audio_message_enable"`
		ImageMessageEnable     bool            `yaml:"image_message_enable"`
		VideoMessageEnable     bool            `yaml:"video_message_enable"`
		LiveMessageEnable      bool            `yaml:"live_message_enable"`
		ContactMessageEnable   bool            `yaml:"contact_message_enable"`
		DetectQRCodeEnabled    bool            `yaml:"detect_image"`
		DetectLinkEnabled      bool            `yaml:"detect_link"`
		LimitMessageDuration   int64           `yaml:"limit_message_duration"`
		LimitMessageNumber     int             `yaml:"limit_message_number"`
		OperatorList           []string        `yaml:"operator_list"`
		Operators              map[string]bool `yaml:"-"`
		PayToJoin              bool            `yaml:"pay_to_join"`
		AccpetPaymentAssetList []PaymentAsset  `yaml:"accept_asset_list"`
	} `yaml:"system"`
	Appearance struct {
		HomeWelcomeMessage string          `yaml:"home_welcome_message"`
		HomeShortcutGroups []ShortcutGroup `yaml:"home_shortcut_groups"`
	} `yaml:"appearance"`
	MessageTemplate struct {
		WelcomeMessage          string `yaml:"welcome_message"`
		MessageTipsGuest        string `yaml:"message_tips_guest"`
		MessageTipsHelp         string `yaml:"message_tips_help"`
		GroupRedPacket          string `yaml:"group_redpacket"`
		GroupRedPacketShortDesc string `yaml:"group_redpacket_short_desc"`
		GroupRedPacketDesc      string `yaml:"group_redpacket_desc"`
		GroupOpenedRedPacket    string `yaml:"group_opened_redpacket"`
		MessageProhibit         string `yaml:"message_prohibit"`
		MessageAllow            string `yaml:"message_allow"`
		MessageTipsJoin         string `yaml:"message_tips_join"`
		MessageTipsHelpBtn      string `yaml:"message_tips_help_btn"`
		MessageTipsUnsubscribe  string `yaml:"message_tips_unsubscribe"`
		MessageRewardLabel      string `yaml:"message_reward_label"`
		MessageRewardMemo       string `yaml:"message_reward_memo"`
		MessageTipsTooMany      string `yaml:"message_tips_too_many"`
		MessageTipsSuspended    string `yaml:"message_tips_suspended"`
		MessageCommandsInfo     string `yaml:"message_commands_info"`
		MessageCommandsInfoResp string `yaml:"message_commands_info_resp"`
	} `yaml:"message_template"`
	Mixin struct {
		ClientId        string `yaml:"client_id"`
		ClientSecret    string `yaml:"client_secret"`
		SessionAssetPIN string `yaml:"session_asset_pin"`
		PinToken        string `yaml:"pin_token"`
		SessionId       string `yaml:"session_id"`
		SessionKey      string `yaml:"session_key"`
		Pin             string `yaml:"pin"`
	} `yaml:"mixin"`
}

type ExportedConfig struct {
	MixinClientId          string          `json:"mixin_client_id"`
	HTTPResourceHost       string          `json:"host"`
	AccpetPaymentAssetList []PaymentAsset  `json:"accept_asset_list"`
	HomeWelcomeMessage     string          `json:"home_welcome_message"`
	HomeShortcutGroups     []ShortcutGroup `json:"home_shortcut_groups"`
}

var AppConfig *Config

func Init(env string) {
	var options map[string]*Config
	err := yaml.Unmarshal(data, &options)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	AppConfig = options[env]
	if AppConfig == nil {
		log.Fatalf("configuration environment %q not found", env)
	}
	if _, err := normalizePrivateKey(AppConfig.Mixin.SessionKey); err != nil {
		log.Fatalf("invalid mixin.session_key for environment %q: %v", env, err)
	}
	AppConfig.System.Operators = make(map[string]bool)
	for _, op := range AppConfig.System.OperatorList {
		AppConfig.System.Operators[op] = true
	}
}

func GetExported() ExportedConfig {
	return ExportedConfig{
		MixinClientId:          AppConfig.Mixin.ClientId,
		HTTPResourceHost:       AppConfig.Service.HTTPResourceHost,
		AccpetPaymentAssetList: AppConfig.System.AccpetPaymentAssetList,
		HomeWelcomeMessage:     AppConfig.Appearance.HomeWelcomeMessage,
		HomeShortcutGroups:     AppConfig.Appearance.HomeShortcutGroups,
	}
}
