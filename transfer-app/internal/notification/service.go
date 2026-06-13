package notification

import (
	"fmt"
	"log"
	"transfer-app/internal/user"
	"transfer-app/pkg/otp"
)

type Service struct {
	smsClient *otp.Client
}

func NewService(smsClient *otp.Client) *Service {
	return &Service{smsClient: smsClient}
}

func (s *Service) SendTransferNotification(sender, receiver *user.User, amount, fee float64, ref string) {
	debitMsg := fmt.Sprintf(
		"NairaTransfer: Debit of ₦%.2f (fee: ₦%.2f) to %s %s. Ref: %s. If not you, call support.",
		amount, fee, receiver.FirstName, receiver.LastName, ref,
	)
	creditMsg := fmt.Sprintf(
		"NairaTransfer: Credit of ₦%.2f from %s %s. Ref: %s.",
		amount, sender.FirstName, sender.LastName, ref,
	)
	if err := s.smsClient.SendSMS(sender.Phone, debitMsg); err != nil {
		log.Printf("❌ debit SMS to %s: %v", sender.Phone, err)
	}
	if err := s.smsClient.SendSMS(receiver.Phone, creditMsg); err != nil {
		log.Printf("❌ credit SMS to %s: %v", receiver.Phone, err)
	}
}

func (s *Service) SendBankTransferNotification(sender *user.User, amount, fee float64, accountName, ref string) {
	msg := fmt.Sprintf(
		"NairaTransfer: ₦%.2f sent to %s (fee: ₦%.2f). Ref: %s. If not you, call support immediately.",
		amount, accountName, fee, ref,
	)
	if err := s.smsClient.SendSMS(sender.Phone, msg); err != nil {
		log.Printf("❌ bank transfer SMS to %s: %v", sender.Phone, err)
	}
}

func (s *Service) SendFundingSuccessNotification(u *user.User, amount float64, ref string) {
	msg := fmt.Sprintf(
		"NairaTransfer: Your wallet has been funded with ₦%.2f. Ref: %s. New balance updated.",
		amount, ref,
	)
	if err := s.smsClient.SendSMS(u.Phone, msg); err != nil {
		log.Printf("❌ funding SMS to %s: %v", u.Phone, err)
	}
}

func (s *Service) SendLoginAlert(u *user.User) {
	msg := fmt.Sprintf(
		"NairaTransfer: New login detected on your account. If not you, change your password immediately.",
	)
	if err := s.smsClient.SendSMS(u.Phone, msg); err != nil {
		log.Printf("❌ login alert SMS to %s: %v", u.Phone, err)
	}
}
