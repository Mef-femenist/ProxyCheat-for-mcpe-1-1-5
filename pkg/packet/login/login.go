package login

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"mefproxy/pkg/jose"
	"mefproxy/pkg/jose/jwt"
	"strconv"
	"strings"
	"time"
)

type chain []string

type loginRequest struct {
	
	Chain chain `json:"chain"`
	
	RawToken string `json:"-"`
}
type AuthResult struct {
	PublicKey             *ecdsa.PublicKey `json:"PublicKey"`
	XBOXLiveAuthenticated bool             `json:"XBOXLiveAuthenticated,omitempty"`
}
type IdentityData struct {
	
	XUID string `json:"XUID,omitempty"`
	
	Identity string `json:"identity"`
	
	DisplayName string `json:"displayName"`
	
	TitleID string `json:"titleId,omitempty"`
}
type ClientData struct {
	ClientRandomID int64 `json:"ClientRandomId"`
	
	ADRole    int    `json:"ADRole"`
	GuiScale  int    `json:"GuiScale"`
	TenantId  string `json:"TenantId"`
	UIProfile int    `json:"UIProfile"`
	
	CurrentInputMode int `json:"CurrentInputMode"`
	
	DefaultInputMode int `json:"DefaultInputMode"`
	
	DeviceModel string `json:"DeviceModel"`
	
	DeviceOS int `json:"DeviceOS"`
	
	GameVersion   string `json:"GameVersion"`
	ServerAddress string `json:"ServerAddress"`
	LanguageCode  string `json:"LanguageCode"`
	SkinData      string `json:"SkinData"`
	SkinID        string `json:"SkinId"`
}

func init() {
	
}

func (c identityClaims) Validate(e jwt.Expected) error {
	if err := c.Claims.Validate(e); err != nil {
		return err
	}
	return c.ExtraData.Validate()
}
func (data IdentityData) Validate() error {
	if _, err := strconv.ParseInt(data.XUID, 10, 64); err != nil && len(data.XUID) != 0 {
		return fmt.Errorf("XUID must be parseable as an int64, but got %v", data.XUID)
	}
	if _, err := uuid.Parse(data.Identity); err != nil {
		return fmt.Errorf("UUID must be parseable as a valid UUID, but got %v", data.Identity)
	}
	if len(data.DisplayName) == 0 || len(data.DisplayName) > 15 {
		return fmt.Errorf("DisplayName must not be empty or longer than 15 characters, but got %v characters", len(data.DisplayName))
	}
	if data.DisplayName[0] == ' ' || data.DisplayName[len(data.DisplayName)-1] == ' ' {
		return fmt.Errorf("DisplayName may not have a space as first/last character, but got %v", data.DisplayName)
	}

	if strings.Contains(data.DisplayName, "  ") {
		return fmt.Errorf("DisplayName must only have single spaces, but got %v", data.DisplayName)
	}
	return nil
}

func decodeChain(buf *bytes.Buffer) (chain, error) {
	var chainLength int32
	if err := binary.Read(buf, binary.LittleEndian, &chainLength); err != nil {
		return nil, fmt.Errorf("error reading chain length: %v", err)
	}
	chainData := buf.Next(int(chainLength))

	request := &loginRequest{}
	if err := json.Unmarshal(chainData, request); err != nil {
		return nil, fmt.Errorf("error decoding request chain JSON: %v", err)
	}
	
	if len(request.Chain) == 0 {
		return nil, fmt.Errorf("connection request had no claims in the chain")
	}
	return request.Chain, nil
}

type identityClaims struct {
	jwt.Claims

	ExtraData IdentityData `json:"extraData"`

	IdentityPublicKey string `json:"identityPublicKey"`
}

type request struct {
	
	Chain chain `json:"chain"`
	
	RawToken string `json:"-"`
}

func MarshalPublicKey(key *ecdsa.PublicKey) string {
	data, _ := x509.MarshalPKIXPublicKey(key)
	return base64.StdEncoding.EncodeToString(data)
}

func MarshalPrivateKey(key *ecdsa.PrivateKey) string {
	data, _ := x509.MarshalPKCS8PrivateKey(key)
	return base64.StdEncoding.EncodeToString(data)
}

func EncodeOffline(identityData IdentityData, data ClientData, key *ecdsa.PrivateKey) []byte {
	keyData := MarshalPublicKey(&key.PublicKey)
	claims := jwt.Claims{
		Expiry:    jwt.NewNumericDate(time.Now().Add(31535547 * time.Second)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(454 * -time.Second)),
	}

	signer, _ := jose.NewSigner(jose.SigningKey{Key: key, Algorithm: jose.ES384}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{"x5u": keyData},
	})
	firstJWT, _ := jwt.Signed(signer).Claims(identityClaims{
		Claims:            claims,
		ExtraData:         identityData,
		IdentityPublicKey: keyData,
	}).CompactSerialize()

	request := &request{Chain: chain{firstJWT}}
	
	request.RawToken, _ = jwt.Signed(signer).Claims(data).CompactSerialize()
	
	return loginEncodeRequest(request)
}

func loginEncodeRequest(req *request) []byte {
	chainBytes, _ := json.Marshal(req)
	
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, int32(len(chainBytes)+1))
	_, _ = buf.WriteString(string(chainBytes) + "\n") 
	
	_ = binary.Write(buf, binary.LittleEndian, int32(len(req.RawToken)))
	_, _ = buf.WriteString(req.RawToken)
	return buf.Bytes()
}

func GetLoginEncodedBytes(id IdentityData, cd ClientData) ([]byte, *ecdsa.PrivateKey) {
	
	privHexed, _ := base64.StdEncoding.DecodeString("MIGkAgEBBDCK28jd+vHYjm6oz8Yi+NFC96RHkMEGlnvfxzuhMqKWIEKSwaWxClAz8O1d4a1tM5igBwYFK4EEACKhZANiAAQCAdEySvmCEQ2G8NKVBvLJmaOeE6+QQwFdS/Vj8xQXqO9G6cnunNX1lnhepD3P1ORwBy66Bj3+dPSGsihVaKNbDCA+SWIWtDvKRFG5Jz2G12yro9cumvs9lRMyvJ4u+oI=")
	key, err := x509.ParseECPrivateKey(privHexed)
	if err != nil {
		panic(err)
	}
	
	return EncodeOffline(id, cd, key), key
}
