package user_registration

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
	"unicode"
)

const (
	maxBytesPerHash          int  = 72
	defaultPasswordMinLength uint = 8
	defaultPasswordMaxLength uint = 32
)

type resetCode struct {
	Email  string
	Expiry time.Time
}

type UserRegistration struct {
	userSource           UserSource
	mailSender           MailSender
	passwordRequirements *PasswordRequirements
	resetCodes           map[string]resetCode
}

type PasswordRequirements struct {
	MinLength   *uint // default
	MaxLength   *uint
	MinLowers   *uint
	MinUppers   *uint
	MinNumbers  *uint
	MinSpecials *uint
}

type NewUserRegistrationConfig struct {
	UserSource           UserSource
	MailSender           MailSender
	PasswordRequirements *PasswordRequirements
}

func NewUserRegistration(cfg *NewUserRegistrationConfig) (*UserRegistration, error) {
	if cfg == nil {
		return nil, errors.New("NewUserRegistrationConfig cannot be a nil pointer")
	}

	if cfg.UserSource == nil {
		return nil, errors.New("UserSource cannot be a nil pointer")
	}

	if cfg.PasswordRequirements == nil {
		return nil, errors.New("PasswordRequirements cannot be a nil pointer")
	}

	return &UserRegistration{
		userSource:           cfg.UserSource,
		mailSender:           cfg.MailSender,
		passwordRequirements: cfg.PasswordRequirements,
		resetCodes:           make(map[string]resetCode),
	}, nil
}

func (u *UserRegistration) HasMailSender() bool {
	return u.mailSender != nil
}

func (u *UserRegistration) Register(email, password, confirmPassword string) (bool, string, string, string, error) {
	user, err := u.userSource.Select(email)
	if err != nil {
		return false, "", "", "", err
	}

	if user != nil {
		return false, "email already registered", "", "", nil
	}

	ok := u.verifyPassword(password)
	if !ok {
		return false, "", u.passwordError(), "", nil
	}

	if password != confirmPassword {
		return false, "", "", "passwords are not the same", nil
	}

	var code = ""

	if u.HasMailSender() {
		code, err = getCode(email)
		if err != nil {
			return false, "", "", "", err
		}

		defer func() {
			err = u.mailSender.Confirm(email, code)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return false, "", "", "", err
	}

	err = u.userSource.Insert(User{
		Email:            email,
		Password:         hashed,
		ConfirmationCode: code,
		CreatedAt:        time.Now(),
		ConfirmedAt:      nil,
	})
	if err != nil {
		return false, "", "", "", err
	}

	return true, "", "", "", nil
}

func (u *UserRegistration) Reset(code, password, confirmPassword string) (bool, string, string, error) {
	email, err := u.ValidateResetCode(code)
	if err != nil {
		return false, "", "", err
	}

	ok := u.verifyPassword(password)
	if !ok {
		return false, u.passwordError(), "", nil
	}

	if password != confirmPassword {
		return false, "", "passwords are not the same", nil
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return false, "", "", err
	}

	user, err := u.userSource.Select(email)
	if err != nil {
		return false, "", "", err
	}

	if user == nil {
		return false, "", "", errors.New("user does not exist")
	}

	user.Password = hashed
	if user.ConfirmedAt == nil {
		now := time.Now()
		user.ConfirmedAt = &now
	}

	err = u.userSource.Update(*user)
	if err != nil {
		return false, "", "", err
	}

	delete(u.resetCodes, code)

	return true, "", "", nil
}

func getCode(prefix string) (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	if prefix != "" {
		randomBytes = append([]byte(prefix+":"), randomBytes...)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

func (u *UserRegistration) Login(email, password string) (*User, string, string, error) {
	user, err := u.userSource.Select(email)
	if err != nil {
		return nil, "", "", err
	}

	if user != nil {
		if checkPasswordHash(password, user.Password) {
			if u.HasMailSender() && user.ConfirmedAt == nil {
				return nil, "email not confirmed yet, check your inbox", "", nil
			}

			return user, "", "", nil
		}
	}

	return nil, "invalid email and/or password", "invalid email and/or password", nil
}

func (u *UserRegistration) ValidateResetCode(code string) (string, error) {
	t, ok := u.resetCodes[code]
	if !ok {
		return "", errors.New("password reset code invalid or expired")
	}

	if time.Now().After(t.Expiry) {
		return "", errors.New("password reset code invalid or expired")
	}

	return t.Email, nil
}

func (u *UserRegistration) GetUser(email string) (*User, error) {
	return u.userSource.Select(email)
}

func (u *UserRegistration) Confirm(code string) error {
	decoded, err := base64.URLEncoding.DecodeString(code)
	if err != nil {
		return err
	}

	codeSplit := strings.Split(string(decoded), ":")
	if len(codeSplit) == 1 {
		return errors.New("invalid confirmation code")
	}

	email := codeSplit[0]
	user, err := u.userSource.Select(email)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user does not exist anymore")
	}

	if user.ConfirmationCode != code {
		return errors.New("invalid confirmation code")
	}

	now := time.Now()
	user.ConfirmedAt = &now

	return u.userSource.Update(*user)
}

func (u *UserRegistration) Forgot(email string) error {
	if !u.HasMailSender() {
		return errors.New("no e-mail sender configured")
	}

	user, err := u.GetUser(email)
	if err != nil {
		return err
	}

	if user != nil {
		code, err := getCode("")
		if err != nil {
			return err
		}
		u.resetCodes[code] = resetCode{
			Email:  email,
			Expiry: time.Now().Add(time.Hour),
		}

		err = u.mailSender.Reset(email, code)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UserRegistration) passwordError() string {
	var errorItems []string

	minLength := defaultPasswordMinLength
	if u.passwordRequirements.MinLength != nil {
		minLength = *u.passwordRequirements.MinLength
	}
	errorItems = append(errorItems, fmt.Sprintf("at least %v characters", minLength))

	maxLength := defaultPasswordMaxLength
	if u.passwordRequirements.MaxLength != nil {
		maxLength = *u.passwordRequirements.MaxLength
	}
	errorItems = append(errorItems, fmt.Sprintf("at most %v characters", maxLength))

	if u.passwordRequirements.MinLowers != nil {
		errorItems = append(errorItems, fmt.Sprintf("at least %v lower case letter(s)", *u.passwordRequirements.MinLowers))
	}

	if u.passwordRequirements.MinUppers != nil {
		errorItems = append(errorItems, fmt.Sprintf("at least %v upper case letter(s)", *u.passwordRequirements.MinUppers))
	}

	if u.passwordRequirements.MinNumbers != nil {
		errorItems = append(errorItems, fmt.Sprintf("at least %v number(s)", *u.passwordRequirements.MinNumbers))
	}

	if u.passwordRequirements.MinSpecials != nil {
		errorItems = append(errorItems, fmt.Sprintf("at least %v special character(s)", *u.passwordRequirements.MinSpecials))
	}

	return fmt.Sprintf("Password does not fulfill one or more of the following requirements: %s.", strings.Join(errorItems, ", "))
}

func (u *UserRegistration) verifyPassword(s string) bool {
	minLength := defaultPasswordMinLength
	if u.passwordRequirements.MinLength != nil {
		minLength = *u.passwordRequirements.MinLength
	}
	if uint(len(s)) < minLength {
		return false
	}

	maxLength := defaultPasswordMaxLength
	if u.passwordRequirements.MaxLength != nil {
		maxLength = *u.passwordRequirements.MaxLength
	}
	if uint(len(s)) > maxLength {
		return false
	}

	if strings.Contains(s, " ") {
		return false
	}

	var lowers uint = 0
	var uppers uint = 0
	var numbers uint = 0
	var specials uint = 0

	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			numbers++
		case unicode.IsUpper(c):
			uppers++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			specials++
		case unicode.IsLetter(c):
			lowers++
		default:
			//return false, false, false, false
		}
	}

	if u.passwordRequirements.MinLowers != nil {
		if lowers < *u.passwordRequirements.MinLowers {
			return false
		}
	}

	if u.passwordRequirements.MinUppers != nil {
		if uppers < *u.passwordRequirements.MinUppers {
			return false
		}
	}

	if u.passwordRequirements.MinNumbers != nil {
		if numbers < *u.passwordRequirements.MinNumbers {
			return false
		}
	}

	if u.passwordRequirements.MinSpecials != nil {
		if specials < *u.passwordRequirements.MinSpecials {
			return false
		}
	}

	return true
}

func hashPassword(password string) (string, error) {
	b := []byte(password)
	var hashed []string

	for {
		b1 := b
		if len(b1) > maxBytesPerHash {
			b1 = b1[:maxBytesPerHash]
		}
		bytes, err := bcrypt.GenerateFromPassword(b1, 14)
		if err != nil {
			return "", err
		}

		hashed = append(hashed, string(bytes))

		if len(b) > maxBytesPerHash {
			b = b[maxBytesPerHash:]
		} else {
			break
		}
	}

	return strings.Join(hashed, " "), nil
}

func checkPasswordHash(password, hash string) bool {
	b := []byte(password)

	for _, h := range strings.Split(hash, " ") {
		if len(b) == 0 {
			return false
		}

		b1 := b
		if len(b1) > maxBytesPerHash {
			b1 = b1[:maxBytesPerHash]
		}
		err := bcrypt.CompareHashAndPassword([]byte(h), b1)
		if err != nil {
			return false
		}
		if len(b) > maxBytesPerHash {
			b = b[maxBytesPerHash:]
		} else {
			b = []byte{}
		}
	}

	return true
}
