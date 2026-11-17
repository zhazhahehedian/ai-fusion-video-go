package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerReq struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Nickname        string `json:"nickname"`
}

type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

func (d Deps) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	result, err := d.Auth.Register(r.Context(), req.Username, req.Password, req.ConfirmPassword, req.Nickname)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, result)
}

func (d Deps) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	result, err := d.Auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, result)
}

func (d Deps) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	result, err := d.Auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, result)
}

func (d Deps) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := d.Auth.Logout(r.Context(), extractAccessToken(r)); err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, true)
}

func (d Deps) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	info, err := d.Auth.UserInfo(r.Context(), session.UserID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, info)
}
