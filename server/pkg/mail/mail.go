package mail

import (
    "bytes"
    "encoding/json"
    "net/http"

    _ "embed"
)

//go:embed emails/verification_code.html
var VerificationTemplate []byte

// SendEmail posts to dployr base to with given params.
// - toAddr, toName: recipient
// - subject: email subject line
// - templateBytes: raw HTML with {{KEY}} placeholders
// - vars: map of KEY→value to replace in the template
func SendEmail(to, subject, body, name string) error {
    payload := map[string]string{
        "to": to, 
        "subject": subject, 
        "body": body,
    }
    if name != "" {
        payload["name"] = name
    }
    jsonData, _ := json.Marshal(payload)
    _, err := http.Post(
        "https://dployr-base.tobimadehin.workers.dev/api/send-email",
        "application/json", 
        bytes.NewBuffer(jsonData),
    )
    return err
}
