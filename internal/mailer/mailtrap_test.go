package mailer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMailTrapClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		fromEmail string
		wantErr   bool
	}{
		{
			name:      "valid client creation",
			apiKey:    "c52bdad0a3dd8628e54d0ef69d4ff646",
			fromEmail: "noreply@hello@demomailtrap.co",
			wantErr:   false,
		},
		{
			name:      "empty api key should error",
			apiKey:    "",
			fromEmail: "test@hello@demomailtrap.co",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMailTrapClient(tt.apiKey, tt.fromEmail)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, mailtrapClient{}, client)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.apiKey, client.apiKey)
				assert.Equal(t, tt.fromEmail, client.fromEmail)
			}
		})
	}
}

func TestMailtrapClient_Send_TemplateProcessing(t *testing.T) {
	client, err := NewMailTrapClient("c52bdad0a3dd8628e54d0ef69d4ff646", "hello@demomailtrap.co")
	require.NoError(t, err)

	ctx := context.Background()

	status, err := client.Send(ctx, UserWelcomeTemplate, "easybizwithai@gmail.com", map[string]string{
		"Username": "John Doe",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestMailtrapClient_Send_InvalidTemplate(t *testing.T) {
	client, err := NewMailTrapClient("c52bdad0a3dd8628e54d0ef69d4ff646", "noreply@hello@demomailtrap.co")
	require.NoError(t, err)

	ctx := context.Background()

	// Test with non-existent template
	_, err = client.Send(ctx, "nonexistent.tmpl", "recipient@example.com", map[string]string{
		"Name": "John",
	})

	assert.Error(t, err)
	// Should be a template parsing error
	assert.Contains(t, err.Error(), "no such file") // or similar template error
}
