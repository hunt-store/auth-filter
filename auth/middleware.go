package auth

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// 사용자 정보 구조체
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 사용자 검증 함수 (API 서버로 요청)
func isValidUser(email, password string) bool {
	// API 서버 URL
	apiURL := "http://localhost:8080/users"

	// 요청 데이터 준비
	userData := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonData, _ := json.Marshal(userData)

	// API 요청
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending request to API:", err)
		return false
	}
	defer resp.Body.Close()

	// 응답 상태코드 체크
	return resp.StatusCode == http.StatusOK
}

// 회원가입 핸들러
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "잘못된 요청입니다", http.StatusBadRequest)
		return
	}

	// API 서버 URL
	apiURL := "http://localhost:8080/users"

	// 요청 데이터 준비
	registrationRequest := map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}
	jsonData, err := json.Marshal(registrationRequest)
	if err != nil {
		http.Error(w, "데이터 처리 중 오류가 발생했습니다", http.StatusInternalServerError)
		return
	}

	// API 서버에 사용자 추가 요청
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "사용자를 추가하는 데 실패했습니다", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 사용자 등록 성공 시 응답
	w.WriteHeader(http.StatusCreated) // 201 Created
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "사용자 등록이 완료되었습니다."})
}

// 로그인 핸들러
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 사용자 인증 로직
	if isValidUser(user.Email, user.Password) {
		tokenString, err := GenerateJWT(user.Email, user.Password)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// JWT 인증 미들웨어
func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		// 서버에 토큰 검증 요청
		resp, err := http.Post("http://localhost:8080/validate-token", "application/json", bytes.NewBuffer([]byte(`{"token":"`+tokenString+`"}`)))
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 토큰이 유효하면 다음 핸들러로 진행
		next.ServeHTTP(w, r)
	})
}
