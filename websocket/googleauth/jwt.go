package googleauth

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"

	. "Redis-Exploration/websocket/models"

	uuid "github.com/satori/go.uuid"
)

type JWTAuthentication struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

const (
	tokenDuration = 72
	expireOffset  = 3600
)

var jwtAuthentication *JWTAuthentication = nil

type MyCustomClaims struct {
	Sub string `json:"sub"`
	jwt.StandardClaims
}

func getPrivateKey() *rsa.PrivateKey {
	abs, err := filepath.Abs("./googleauth/private_key.txt")
	if err != nil {
		panic(err)
	}

	privateKeyFile, err := os.Open(abs)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	abs, err := filepath.Abs("./googleauth/public_key.pub")
	if err != nil {
		panic(err)
	}

	publicKeyFile, err := os.Open(abs)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}

// InitJWTAuthentication instantiates a JWTAuthentication instance
func InitJWTAuthentication() *JWTAuthentication {
	if jwtAuthentication == nil {
		jwtAuthentication = &JWTAuthentication{
			privateKey: getPrivateKey(),
			PublicKey:  getPublicKey(),
		}
	}

	return jwtAuthentication
}

// GenerateToken generates a token based on the userID given
func (authentication *JWTAuthentication) GenerateToken(userID string) (string, error) {
	const JWTExpirationDelta = 72

	// Create the Claims
	claims := MyCustomClaims{
		userID,
		jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(JWTExpirationDelta)).Unix(),
			Issuer:    "Fabulous Kev Kev",
		},
	}
	// token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
	// 	"exp": time.Now().Add(time.Hour * time.Duration(JWTExpirationDelta)).Unix(),
	// 	"iat": time.Now().Unix(),
	// 	"sub": userID,
	// })

	// Sign and get the complete encoded token as a string using the secret
	// tokenString, err := token.SignedString(authentication.privateKey)
	tokenString, err := token.SignedString([]byte("motherfucker"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v %v", tokenString, err)
	return tokenString, nil
}

func (authentication *JWTAuthentication) VerifyToken(userID string, tokenStr string) bool {

	token, err := jwt.ParseWithClaims(tokenStr, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("motherfucker"), nil
	})
	fmt.Println("token parsed")
	fmt.Println(token.Valid)
	fmt.Println(token.Claims)
	fmt.Println(token)

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		fmt.Printf("%v %v\n", claims.Sub, claims.StandardClaims.ExpiresAt)
		fmt.Println(claims.Sub)
		fmt.Println(claims.StandardClaims.ExpiresAt)
		fmt.Println(claims.StandardClaims.IssuedAt)
		fmt.Println(claims.StandardClaims.Issuer)
		res := authentication.GetTokenRemainingValidity(claims.ExpiresAt)
		fmt.Println(res)
	} else {
		fmt.Println(ok)
		fmt.Println(err)
		return false
	}

	return true
}

func (authentication *JWTAuthentication) GetTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainder := tm.Sub(time.Now())
		if remainder > 0 {
			return int(remainder.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

func (authentication *JWTAuthentication) Authenticate(user User) bool {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testing"), 10)
	uuidStr, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	testUser := User{
		UUID:     uuidStr.String(),
		Username: "John Doe",
		Password: string(hashedPassword),
	}

	return user.Username == testUser.Username && bcrypt.CompareHashAndPassword([]byte(testUser.Password), []byte(user.Password)) == nil
}
