package config

import (
	"io/ioutil"
	"log"
	"path"

	yaml "gopkg.in/yaml.v2"
)

const ConfigFile = "config.yaml"
const BuildVersion = "BUILD_VERSION"

type PaymentAsset struct {
	Symbol string `yaml:"symbol" json:"symbol"`
	AssetId string `yaml:"asset_id" json:"asset_id"`
	Amount string `yaml:"amount" json:"amount"`
}

type Config struct {
	Service struct {
		Name             string `yaml:"name"`
		Environment      string `yaml:"enviroment"`
		HTTPListenPort   int    `yaml:"port"`
		HTTPResourceHost string `yaml:"host"`
	} `yaml:"service"`
	Database struct {
		DatebaseUser     string `yaml:"username"`
		DatabasePassword string `yaml:"password"`
		DatabaseHost     string `yaml:"host"`
		DatabasePort     string `yaml:"port"`
		DatabaseName     string `yaml:"database_name"`
	} `yaml:"database"`
	System struct {
		MessageShardModifier     string   `yaml:"message_shard_modifier"`
		MessageShardSize         int64    `yaml:"message_shard_size"`
		PriceAssetsEnable        bool     `yaml:"price_asset_enable"`
		OperatorList             []string `yaml:"operator_list"`
		Operators                map[string]bool
		DetectImageEnabled       bool   `yaml:"detect_image"`
		DetectLinkEnabled        bool   `yaml:"detect_link"`
		ProhibitedMessageEnabled bool   `yaml:"prohibited_message"`
		PaymentAssetId           string `yaml:"payment_asset_id"`
		PaymentAmount            string `yaml:"payment_amount"`
		AccpetPaymentAssetList   []PaymentAsset `yaml:"accept_asset_list"`
		AccpetWeChatPayment      bool `yaml:"accept_wechat_payment"`
		WeChatAppId 			 string `yaml:"wechat_app_id"`
		WeChatMchId 			 string `yaml:"wechat_mch_id"`
		WeChatMchKey			 string `yaml:"wechat_mch_key"`
	} `yaml:"system"`
	MessageTemplate struct {
		WelcomeMessage          string `yaml:"welcome_message"`
		GroupRedPacket          string `yaml:"group_redpacket"`
		GroupRedPacketShortDesc string `yaml:"group_redpacket_short_desc"`
		GroupRedPacketDesc      string `yaml:"group_redpacket_desc"`
		GroupOpenedRedPacket    string `yaml:"group_opened_redpacket"`
		MessageTipsGuest        string `yaml:"message_tips_guest"`
		MessageTipsJoin         string `yaml:"message_tips_join"`
		MessageTipsHelp         string `yaml:"message_tips_help"`
		MessageTipsHelpBtn      string `yaml:"message_tips_help_btn"`
		MessageTipsUnsubscribe  string `yaml:"message_tips_unsubscribe"`
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
	} `yaml:"mixin"`
}

type ExportedConfig struct {
	AccpetPaymentAssetList   []PaymentAsset `json:"accept_asset_list"`
	AccpetWeChatPayment      bool `json:"accept_wechat_payment"`
}

var conf *Config

func LoadConfig(dir string) {
	data, err := ioutil.ReadFile(path.Join(dir, ConfigFile))
	if err != nil {
		log.Panicln(err)
	}
	conf = &Config{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	conf.System.Operators = make(map[string]bool)
	for _, op := range conf.System.OperatorList {
		conf.System.Operators[op] = true
	}
}

func Get() *Config {
	return conf
}

func GetExported () ExportedConfig {
	var exc ExportedConfig
	exc.AccpetPaymentAssetList = conf.System.AccpetPaymentAssetList
	exc.AccpetWeChatPayment = conf.System.AccpetWeChatPayment
	return exc
}
