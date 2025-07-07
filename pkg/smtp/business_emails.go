package smtp

import (
	"fmt"
	"time"
	model "greenlync-api-gateway/model/common/v1"
)

func (s *SMTP) SendEmailReport(user *model.User, to string, reportDesc string, accountName string, downloadURL string, createdAt time.Time) error {
	// Cannabis boilerplate - simplified email template
	date := createdAt.Format("Jan 2 2006")
	companyName := user.CompanyName
	if companyName == "" {
		companyName = "GreenLync"
	}

	htmlContent := fmt.Sprintf(`
		<html>
		<body>
			<h2>%s - Your Report is Ready!</h2>
			<p>Hello %s,</p>
			<p>Your %s report has been generated successfully.</p>
			<p>Download Link: <a href="%s">Download Report</a></p>
			<p>Generated on: %s</p>
			<br>
			<p>Best regards,<br>%s Team</p>
		</body>
		</html>
		`, companyName, accountName, reportDesc, downloadURL, date, companyName)

	subject := fmt.Sprintf("%s, Your Report is Ready!!", companyName)
	err := s.QueueEmail(to, subject, htmlContent)
	if err != nil {
		s.Log.Logger.Error(err)
		return err
	}

	return nil
}

func (s *SMTP) SendApproveDemoAccountEmail(user *model.User, to, accountName, username, password string) error {
	// Cannabis boilerplate - simplified demo account email
	companyName := user.CompanyName
	if companyName == "" {
		companyName = "GreenLync"
	}

	htmlContent := fmt.Sprintf(`
		<html>
		<body>
			<h2>%s - Demo Account Created</h2>
			<p>Hello %s,</p>
			<p>Your demo account has been created successfully.</p>
			<p>Username: %s</p>
			<p>Password: %s</p>
			<p>Please keep this information secure.</p>
			<br>
			<p>Best regards,<br>%s Team</p>
		</body>
		</html>
		`, companyName, accountName, username, password, companyName)

	subject := fmt.Sprintf("%s, Thank you for creating a demo account with us!", companyName)
	err := s.QueueEmail(to, subject, htmlContent)
	if err != nil {
		s.Log.Logger.Error(err)
		return err
	}

	return nil
}

func (s *SMTP) SendVerifyChangePasswordEmail(user *model.User, to string, accountName, code string, expiryDate time.Duration) error {
	// Cannabis boilerplate - simplified password change email
	companyName := user.CompanyName
	if companyName == "" {
		companyName = "GreenLync"
	}

	htmlContent := fmt.Sprintf(`
		<html>
		<body>
			<h2>%s - Password Change Verification</h2>
			<p>Hello %s,</p>
			<p>We received a request to change your password.</p>
			<p>Verification Code: <strong>%s</strong></p>
			<p>This code will expire in %d minutes.</p>
			<p>If you did not request this change, please ignore this email.</p>
			<br>
			<p>Best regards,<br>%s Team</p>
		</body>
		</html>
		`, companyName, accountName, code, int(expiryDate.Minutes()), companyName)

	subject := fmt.Sprintf("%s, Verify your password change request", companyName)
	err := s.QueueEmail(to, subject, htmlContent)
	if err != nil {
		s.Log.Logger.Error(err)
		return err
	}

	return nil
}
