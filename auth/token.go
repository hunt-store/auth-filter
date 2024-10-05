package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	utils "auth-filter/utility" // utils 패키지 import
)

// AuthRequest 로그인 요청 구조체
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GenerateJWT JWT 생성 함수
func GenerateJWT(email string, password string) (string, error) {
	// .env 파일 로드
	err := utils.LoadEnv()
	if err != nil {
		return "", err
	}

	// API 서버에 인증 요청
	authRequest := AuthRequest{Email: email, Password: password}
	requestBody, err := json.Marshal(authRequest)
	if err != nil {
		return "", err
	}

	// API 서버 주소 생성
	url := utils.GetAPIURL("/users/login")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// 토큰을 API 응답에서 추출
	token, ok := result["data"].(map[string]interface{})["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}
