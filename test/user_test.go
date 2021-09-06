package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"mlauth/pkg/api"
	"mlauth/pkg/srv"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

func TestUserGet(t *testing.T) {
	r := api.Route()
	at, _ := userLogin(t, r, "testusername", "testpassword")
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/api/users/me", nil)
	assert.NoError(t, err)

	req.Header.Set("x-access-token", at)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	body := struct {
		Uid         int       `json:"uid"`
		Username    string    `json:"username"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, 1, body.Uid)
	assert.Equal(t, "testusername", body.Username)
	assert.Equal(t, "testemail@gmail.com", body.Email)
	assert.Equal(t, "test display name", body.DisplayName)
	assert.Equal(t, true, body.IsActive)

	p, err := time.Parse(time.RFC3339, "1999-01-08T04:05:06Z")
	assert.NoError(t, err)
	assert.Equal(t, p, body.CreatedAt)
}

func TestUserRegister(t *testing.T) {
	r := api.Route()
	b, err := json.Marshal(gin.H{
		"username":     "testU",
		"password":     "testPassYou",
		"email":        "testE@outlook.com",
		"display_name": "符号看象限ラブライブ한국어",
	})
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/users", bytes.NewReader(b))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	body := struct {
		Uid         int       `json:"uid"`
		Username    string    `json:"username"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "testU", body.Username)
	assert.Equal(t, "testE@outlook.com", body.Email)
	assert.Equal(t, "符号看象限ラブライブ한국어", body.DisplayName)
	assert.Equal(t, false, body.IsActive)
	assert.Less(t, time.Now().UTC().Sub(body.CreatedAt), 5*time.Second)

	var email, emailBody string
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	select {
	case email = <-srv.SendEmailMockChan:
	case <-ctx.Done():
		assert.NotEmpty(t, email, "Can not get email")
	}
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	select {
	case emailBody = <-srv.SendEmailMockChan:
	case <-ctx.Done():
		assert.NotEmpty(t, emailBody, "Can not get email body")
	}
	cancel()
	assert.Equal(t, "testE@outlook.com", email)

	re := regexp.MustCompile(`/emails/active/?\?active-code=[0-9a-z-]+`)
	p := re.Find([]byte(emailBody))
	assert.NotNil(t, p)

	req, err = http.NewRequest("POST", "/api"+string(p), nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	at, _ := userLogin(t, r, "testU", "testPassYou")
	w = httptest.NewRecorder()

	req, err = http.NewRequest("GET", "/api/users/me", nil)
	assert.NoError(t, err)

	req.Header.Set("x-access-token", at)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	body2 := struct {
		Uid         int       `json:"uid"`
		Username    string    `json:"username"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &body2)
	assert.NoError(t, err)
	assert.Equal(t, body.Uid, body2.Uid)
	assert.Equal(t, true, body2.IsActive)
}

func TestUserEditExceptEmail(t *testing.T) {
	r := api.Route()
	at, _ := userLogin(t, r, "anotherusername", "anotherpassword")

	req, err := http.NewRequest("GET", "/api/users/me", nil)
	assert.NoError(t, err)

	req.Header.Set("x-access-token", at)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	body := struct {
		Uid         int       `json:"uid"`
		Username    string    `json:"username"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, 2, body.Uid)

	b, err := json.Marshal(gin.H{
		"password":     "testPassYou",
		"display_name": "符号看象限ラブライブ한국어",
	})
	assert.NoError(t, err)

	req, err = http.NewRequest("PUT", "/api/users/me", bytes.NewReader(b))
	assert.NoError(t, err)

	req.Header.Set("x-access-token", at)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	body2 := struct {
		Uid         int       `json:"uid"`
		Username    string    `json:"username"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &body2)
	assert.NoError(t, err)
	assert.Equal(t, body.Uid, body2.Uid)
	assert.Equal(t, "符号看象限ラブライブ한국어", body2.DisplayName)

	select {
	case <-srv.SendEmailMockChan:
		<-srv.SendEmailMockChan
		assert.NotNil(t, nil, "No email edit but triggers email sending")
	default:
	}

	at, _ = userLogin(t, r, "anotherusername", "testPassYou")
	req, err = http.NewRequest("GET", "/api/users/me", nil)
	assert.NoError(t, err)

	req.Header.Set("x-access-token", at)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "body: %s", w.Body.String())
}
