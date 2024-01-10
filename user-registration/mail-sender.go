package user_registration

type MailSender interface {
	Confirm(email, code string) error
	Reset(email, code string) error
}
