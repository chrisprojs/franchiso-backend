package service

import (
	"bytes"
	"fmt"
	"net/smtp"

	"github.com/chrisprojs/Franchiso/config"
)

// SendVerificationEmail sends verification email with code
func SendVerificationEmail(emailConfig *config.EmailConfig, toEmail, toName, verificationCode string) error {
	// Set up authentication
	auth := smtp.PlainAuth("", emailConfig.SMTPUsername, emailConfig.SMTPPassword, emailConfig.SMTPHost)

	// Email subject
	subject := "Kode Verifikasi Email - Franchiso"

	// Email body (HTML format)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body {
			font-family: Arial, sans-serif;
			line-height: 1.6;
			color: #333;
			max-width: 600px;
			margin: 0 auto;
			padding: 20px;
		}
		.container {
			background-color: #f9f9f9;
			border-radius: 8px;
			padding: 30px;
			margin: 20px 0;
		}
		.header {
			text-align: center;
			margin-bottom: 30px;
		}
		.header h1 {
			color: #2c3e50;
			margin: 0;
		}
		.code-box {
			background-color: #ffffff;
			border: 2px dashed #3498db;
			border-radius: 8px;
			padding: 20px;
			text-align: center;
			margin: 30px 0;
		}
		.code {
			font-size: 32px;
			font-weight: bold;
			letter-spacing: 5px;
			color: #2c3e50;
			font-family: 'Courier New', monospace;
		}
		.footer {
			margin-top: 30px;
			padding-top: 20px;
			border-top: 1px solid #ddd;
			font-size: 12px;
			color: #777;
			text-align: center;
		}
		.warning {
			background-color: #fff3cd;
			border-left: 4px solid #ffc107;
			padding: 15px;
			margin: 20px 0;
			border-radius: 4px;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Verifikasi Email Anda</h1>
		</div>
		
		<p>Halo <strong>%s</strong>,</p>
		
		<p>Terima kasih telah mendaftar di Franchiso. Untuk menyelesaikan proses registrasi, silakan gunakan kode verifikasi berikut:</p>
		
		<div class="code-box">
			<div class="code">%s</div>
		</div>
		
		<div class="warning">
			<strong>Perhatian:</strong> Kode ini hanya berlaku selama 10 menit. Jangan bagikan kode ini kepada siapapun.
		</div>
		
		<p>Jika Anda tidak melakukan registrasi ini, silakan abaikan email ini.</p>
		
		<div class="footer">
			<p>Email ini dikirim otomatis, mohon tidak membalas email ini.</p>
			<p>&copy; 2024 Franchiso. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`, toName, verificationCode)

	// Build email message
	msg := bytes.Buffer{}
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", emailConfig.FromName, emailConfig.FromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Send email
	addr := fmt.Sprintf("%s:%s", emailConfig.SMTPHost, emailConfig.SMTPPort)
	err := smtp.SendMail(addr, auth, emailConfig.FromEmail, []string{toEmail}, msg.Bytes())
	if err != nil {
		return fmt.Errorf("gagal mengirim email: %v", err)
	}

	return nil
}
