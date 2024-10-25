package insomnia

import "encoding/json"

type Export struct {
	Type      string `json:"_type"`
	Format    int    `json:"__export_format"`
	Resources []Resource
}

type Request struct {
	Headers        []Header        `json:"headers"`
	Method         string          `json:"method"`
	Body           Body            `json:"body"`
	Parameters     []Parameter     `json:"parameters"`
	PathParameters []PathParameter `json:"pathParameters"`
	Authentication Authentication  `json:"authentication"`
}
type Body struct {
	Text     string `json:"text"`
	MimeType string `json:"mimeType"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PathParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Environment struct {
	Data json.RawMessage `json:"data"`
}

type Resource struct {
	ID       string `json:"_id"`
	Type     string `json:"_type"`
	Name     string `json:"name"`
	Url      string `json:"url"`
	ParentID string `json:"parentId"`

	Request
	Environment
}

type Authentication struct {
	Type string `json:"type"`

	APIKey
	AtlassianASAP
	AWSIAM
	Basic
	BearerToken
	Digest
	Hawk
	NTLM
	OAuth1
	OAuth2
}

type APIKey struct {
	Disabled bool   `json:"disabled"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	AddTo    string `json:"addTo"`
}

type AtlassianASAP struct {
	Issuer           string `json:"issuer"`
	Subject          string `json:"subject"`
	Audience         string `json:"audience"`
	AdditionalClaims string `json:"additionalClaims"`
	KeyId            string `json:"keyId"`
	PrivateKey       string `json:"privateKey"`
}

type AWSIAM struct {
	Disabled        bool   `json:"disabled"`
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Region          string `json:"region"`
	Service         string `json:"service"`
}

type Basic struct {
	UseISO88591 bool   `json:"useISO88591"`
	Disabled    bool   `json:"disabled"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type BearerToken struct {
	Token  string `json:"token"`
	Prefix string `json:"prefix"`
}

type Digest struct {
	Disabled bool   `json:"disabled"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Hawk struct {
	Algorithm       string `json:"algorithm"`
	Id              string `json:"id"`
	Key             string `json:"key"`
	Ext             string `json:"ext"`
	ValidatePayload bool   `json:"validatePayload"`
}

type NTLM struct {
	Disabled bool   `json:"disabled"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type OAuth1 struct {
	Disabled        bool   `json:"disabled"`
	SignatureMethod string `json:"signatureMethod"`
	ConsumerKey     string `json:"consumerKey"`
	ConsumerSecret  string `json:"consumerSecret"`
	TokenKey        string `json:"tokenKey"`
	TokenSecret     string `json:"tokenSecret"`
	PrivateKey      string `json:"privateKey"`
	Version         string `json:"version"`
	Nonce           string `json:"nonce"`
	Timestamp       string `json:"timestamp"`
	Callback        string `json:"callback"`
	Realm           string `json:"realm"`
	Verifier        string `json:"verifier"`
	IncludeBodyHash bool   `json:"includeBodyHash"`
}

type OAuth2 struct {
	GrantType        string `json:"grantType"`
	AuthorizationUrl string `json:"authorizationUrl"`
	AccessTokenUrl   string `json:"accessTokenUrl"`
	ClientId         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	UsePkce          bool   `json:"usePkce"`
	RedirectUrl      string `json:"redirectUrl"`
	Scope            string `json:"scope"`
	State            string `json:"state"`
	TokenPrefix      string `json:"tokenPrefix"`
	Audience         string `json:"audience"`
	Resource         string `json:"resource"`
	Origin           string `json:"origin"`
}
