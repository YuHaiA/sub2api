package service

import (
	"context"
	"errors"
	"net/http"
)

// buildOpenAIAuthenticationHeaders is the non-Agent-Identity compatibility
// path used by Live and other OpenAI OAuth calls in this fork.
func (s *OpenAIGatewayService) buildOpenAIAuthenticationHeaders(_ context.Context, account *Account, token string) (http.Header, error) {
	if account == nil {
		return nil, errors.New("account is nil")
	}
	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)
	return headers, nil
}
