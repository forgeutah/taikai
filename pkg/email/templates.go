package email

// EmailVerificationTemplate is the template for email verification
var EmailVerificationTemplate = &Template{
	Subject: "Verify your email - Taikai",
	BodyHTML: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px;">
        <h1 style="color: #4f46e5; margin-bottom: 20px;">Welcome to Taikai!</h1>
        <p>Hi {{.Name}},</p>
        <p>Thank you for registering! Please verify your email address by clicking the button below:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationURL}}" style="background-color: #4f46e5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Verify Email</a>
        </div>
        <p style="color: #666; font-size: 14px;">Or copy and paste this link into your browser:</p>
        <p style="color: #666; font-size: 14px; word-break: break-all;">{{.VerificationURL}}</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">This link will expire in 24 hours.</p>
        <p style="color: #666; font-size: 14px;">If you didn't create an account, please ignore this email.</p>
    </div>
</body>
</html>
`,
	BodyText: "Welcome to Taikai! Please verify your email by visiting: {{.VerificationURL}}",
}

// PasswordResetTemplate is the template for password reset
var PasswordResetTemplate = &Template{
	Subject: "Reset your password - Taikai",
	BodyHTML: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px;">
        <h1 style="color: #4f46e5; margin-bottom: 20px;">Reset Your Password</h1>
        <p>Hi {{.Name}},</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background-color: #4f46e5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        <p style="color: #666; font-size: 14px;">Or copy and paste this link into your browser:</p>
        <p style="color: #666; font-size: 14px; word-break: break-all;">{{.ResetURL}}</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">This link will expire in 1 hour.</p>
        <p style="color: #666; font-size: 14px;">If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
    </div>
</body>
</html>
`,
	BodyText: "Reset your password by visiting: {{.ResetURL}}",
}

// WelcomeTemplate is sent after email verification
var WelcomeTemplate = &Template{
	Subject: "Welcome to Taikai!",
	BodyHTML: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px;">
        <h1 style="color: #4f46e5; margin-bottom: 20px;">Welcome to Taikai! 🎉</h1>
        <p>Hi {{.Name}},</p>
        <p>Your email has been verified successfully! You're all set to start exploring events and connecting with your community.</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.BaseURL}}/events" style="background-color: #4f46e5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Browse Events</a>
        </div>
        <p>Here's what you can do:</p>
        <ul>
            <li>Browse upcoming events</li>
            <li>RSVP to events you're interested in</li>
            <li>Subscribe to groups and organizations</li>
            <li>Manage your profile and preferences</li>
        </ul>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">Have questions? Feel free to reach out to our support team.</p>
    </div>
</body>
</html>
`,
	BodyText: "Welcome to Taikai! Your email has been verified. Start browsing events at: {{.BaseURL}}/events",
}
