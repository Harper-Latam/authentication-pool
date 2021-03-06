package authentication_pool

import (
	"errors"
	"sync"

	"github.com/Harper-Latam/authentication-pool/codes"
)

type TemplateName string

const (
	Validation TemplateName = "validation-email"
	Reminder                = "remind-email"
)

type CodeSender interface {
	Send(templateName TemplateName, to string, code *codes.Code) error
}

type LocalAccountManager struct {
	localAPI      LocalAPI
	localProvider ProviderWithStore
	codeHandler   codes.Manager
	codeSender    CodeSender
}

func NewLocalAccountManager(localAPI LocalAPI, localProvider ProviderWithStore, codeHandler codes.Manager, codeSender CodeSender) *LocalAccountManager {
	return &LocalAccountManager{localAPI, localProvider, codeHandler, codeSender}
}

func (l LocalAccountManager) SendValidationCode(input *SendValidationCodeInput) error {
	user, err := l.localAPI.User(input.Nickname)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("the given user does not exist")
	}

	if user.ValidatedAt != nil {
		return errors.New("the given account has been validated already")
	}

	output, err := l.codeHandler.Issue(&codes.IssueInput{Issuer: user.Email})
	if err != nil {
		return err
	}

	if err = l.codeSender.Send(Validation, user.Email, output.Code); err != nil {
		return err
	}

	return nil
}

func (l LocalAccountManager) ValidateAccount(input *ValidateAccountInput) (*CustomerAccount, error) {
	_, err := l.codeHandler.Used(&codes.CheckCodeInput{
		Issuer: input.Nickname,
		Code:   input.Code,
	})

	if err != nil {
		return nil, err
	}

	return l.localProvider.ValidatedEmail(&ValidateEmailInput{Email: input.Nickname})
}

func (l LocalAccountManager) RemindPassword(input *RemindPasswordInput) error {
	user, err := l.localAPI.User(input.Nickname)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("the given user does not exist")
	}

	if user.ValidatedAt == nil {
		return errors.New("the given account has not been validated")
	}

	output, err := l.codeHandler.Issue(&codes.IssueInput{Issuer: user.Email})
	if err != nil {
		return err
	}

	if err = l.codeSender.Send(Reminder, user.Email, output.Code); err != nil {
		return err
	}

	return nil
}

func (l LocalAccountManager) ResetPassword(input *ResetPasswordInput) (*CustomerAccount, error) {
	_, err := l.codeHandler.Used(&codes.CheckCodeInput{
		Issuer: input.Nickname,
		Code:   input.Code,
	})

	if err != nil {
		return nil, err
	}

	return l.localProvider.UpdatePassword(&UpdatePasswordInput{Email: input.Nickname, Password: input.Password})
}

type send struct {
	code             *codes.Code
	to, templateName string
}

type TestCodeSender struct {
	store map[string]*send
	mx    sync.Mutex
}

func NewTestCodeSender() *TestCodeSender {
	return &TestCodeSender{store: map[string]*send{}}
}

func (t *TestCodeSender) Send(templateName TemplateName, to string, code *codes.Code) error {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.store[to] = &send{
		to:           to,
		code:         code,
		templateName: string(templateName),
	}

	return nil
}
