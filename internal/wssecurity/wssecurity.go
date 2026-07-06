package wssecurity

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"time"
)

const (
	nsWSSE            = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	nsWSU             = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
	typePasswordDigest = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest"
	encodingBase64     = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
)

type SecurityHeader struct {
	XMLName        xml.Name      `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd Security"`
	MustUnderstand string        `xml:"http://schemas.xmlsoap.org/soap/envelope/ mustUnderstand,attr,omitempty"`
	UsernameToken  UsernameToken `xml:"UsernameToken"`
}

type UsernameToken struct {
	Username string   `xml:"Username"`
	Password Password `xml:"Password"`
	Nonce    Nonce    `xml:"Nonce"`
	Created  string   `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd Created"`
}

type Password struct {
	Type  string `xml:"Type,attr"`
	Value string `xml:",chardata"`
}

type Nonce struct {
	EncodingType string `xml:"EncodingType,attr"`
	Value        string `xml:",chardata"`
}

func Digest(nonce []byte, created string, password string) string {
	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(created))
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Created(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func NewUsernameToken(username, password string) (*SecurityHeader, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return NewUsernameTokenDeterministic(username, password, nonce, time.Now()), nil
}

func NewUsernameTokenDeterministic(username, password string, nonce []byte, created time.Time) *SecurityHeader {
	digest := Digest(nonce, Created(created), password)
	return &SecurityHeader{
		MustUnderstand: "1",
		UsernameToken: UsernameToken{
			Username: username,
			Password: Password{
				Type:  typePasswordDigest,
				Value: digest,
			},
			Nonce: Nonce{
				EncodingType: encodingBase64,
				Value:        base64.StdEncoding.EncodeToString(nonce),
			},
			Created: Created(created),
		},
	}
}
