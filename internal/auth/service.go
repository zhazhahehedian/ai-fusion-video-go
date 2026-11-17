package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	q      *sqlcdb.Queries
	tokens *TokenStore
}

func NewService(db *sql.DB, tokens *TokenStore) *Service {
	return &Service{q: sqlcdb.New(db), tokens: tokens}
}

type LoginResult struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	UserID       int64  `json:"userId"`
	Username     string `json:"username"`
	Nickname     any    `json:"nickname"`
}

type UserInfo struct {
	ID         int64    `json:"id"`
	Username   string   `json:"username"`
	Nickname   any      `json:"nickname"`
	Avatar     any      `json:"avatar"`
	Email      any      `json:"email"`
	Phone      any      `json:"phone"`
	Status     int32    `json:"status"`
	CreateTime string   `json:"createTime"`
	Roles      []string `json:"roles"`
}

func (s *Service) Register(ctx context.Context, username, password, confirmPassword, nickname string) (LoginResult, error) {
	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if n := utf8.RuneCountInString(username); n < 3 || n > 32 {
		return LoginResult{}, apperr.BadRequest("用户名长度为 3-32 位")
	}
	if n := utf8.RuneCountInString(password); n < 6 || n > 32 {
		return LoginResult{}, apperr.BadRequest("密码长度为 6-32 位")
	}
	if password != confirmPassword {
		return LoginResult{}, apperr.BadRequest("两次输入的密码不一致")
	}
	if nickname == "" {
		nickname = username
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return LoginResult{}, err
	}
	userID, err := s.q.CreateUser(ctx, sqlcdb.CreateUserParams{
		Username: username,
		Password: string(hash),
		Nickname: sql.NullString{String: nickname, Valid: true},
	})
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return LoginResult{}, apperr.BadRequest("用户名已存在")
		}
		return LoginResult{}, err
	}

	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return LoginResult{}, err
	}
	pair, err := s.tokens.Create(ctx, Session{
		UserID:        user.ID,
		Username:      user.Username,
		CurrentTeamID: nil,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return buildLoginResult(pair, user), nil
}

func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return LoginResult{}, apperr.BadRequest("用户名不能为空")
	}
	if n := utf8.RuneCountInString(password); n < 6 || n > 32 {
		return LoginResult{}, apperr.BadRequest("密码长度为 6-32 位")
	}

	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginResult{}, apperr.Unauthorized("用户名或密码错误")
		}
		return LoginResult{}, err
	}
	if user.Status != 1 {
		return LoginResult{}, apperr.Unauthorized("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return LoginResult{}, apperr.Unauthorized("用户名或密码错误")
	}

	pair, err := s.tokens.Create(ctx, Session{
		UserID:   user.ID,
		Username: user.Username,
		// currentTeamId: stage-2 debt — no team tables yet
		CurrentTeamID: nil,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return buildLoginResult(pair, user), nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (LoginResult, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return LoginResult{}, apperr.BadRequest("刷新令牌不能为空")
	}

	session, ok, err := s.tokens.PeekRefreshSession(ctx, refreshToken)
	if err != nil {
		return LoginResult{}, err
	}
	if !ok {
		return LoginResult{}, apperr.Unauthorized("刷新令牌无效或账户不可用")
	}

	user, err := s.q.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginResult{}, apperr.Unauthorized("刷新令牌无效或账户不可用")
		}
		return LoginResult{}, err
	}
	if user.Status != 1 || user.Username != session.Username {
		return LoginResult{}, apperr.Unauthorized("刷新令牌无效或账户不可用")
	}

	pair, ok, err := s.tokens.Refresh(ctx, refreshToken)
	if err != nil {
		return LoginResult{}, err
	}
	if !ok {
		return LoginResult{}, apperr.Unauthorized("刷新令牌无效或已过期")
	}
	return buildLoginResult(pair, user), nil
}

func (s *Service) Logout(ctx context.Context, accessToken string) error {
	if strings.TrimSpace(accessToken) == "" {
		return nil
	}
	return s.tokens.RemoveAccess(ctx, accessToken)
}

func (s *Service) UserInfo(ctx context.Context, userID int64) (UserInfo, error) {
	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, apperr.Unauthorized("未登录或登录已过期")
		}
		return UserInfo{}, err
	}
	return UserInfo{
		ID:         user.ID,
		Username:   user.Username,
		Nickname:   nullString(user.Nickname),
		Avatar:     nullString(user.Avatar),
		Email:      nullString(user.Email),
		Phone:      nullString(user.Phone),
		Status:     user.Status,
		CreateTime: user.CreateTime.Format("2006-01-02T15:04:05"),
		Roles:      []string{}, // role tables: technical debt
	}, nil
}

// ResolveAccessToken loads the Redis session and verifies the user still matches.
func (s *Service) ResolveAccessToken(ctx context.Context, accessToken string) (Session, error) {
	session, ok, err := s.tokens.GetAccessSession(ctx, accessToken)
	if err != nil {
		return Session{}, err
	}
	if !ok {
		return Session{}, apperr.Unauthorized("未登录或登录已过期")
	}
	user, err := s.q.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, apperr.Unauthorized("未登录或登录已过期")
		}
		return Session{}, err
	}
	if user.Status != 1 || user.Username != session.Username || user.ID != session.UserID {
		return Session{}, apperr.Unauthorized("未登录或登录已过期")
	}
	return session, nil
}

func buildLoginResult(pair TokenPair, user sqlcdb.SysUser) LoginResult {
	return LoginResult{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		UserID:       user.ID,
		Username:     user.Username,
		Nickname:     nullString(user.Nickname),
	}
}

func nullString(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return v.String
}
