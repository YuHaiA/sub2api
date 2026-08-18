//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvedAccountProxyURLFailsClosedForInvalidBindings(t *testing.T) {
	proxyID := int64(7)
	tests := []struct {
		name    string
		account *Account
	}{
		{
			name:    "configured proxy is missing",
			account: &Account{ID: 1, ProxyID: &proxyID},
		},
		{
			name: "loaded proxy does not match configured proxy",
			account: &Account{
				ID:      2,
				ProxyID: &proxyID,
				Proxy:   &Proxy{ID: 8, Protocol: "http", Host: "127.0.0.1", Port: 8080},
			},
		},
		{
			name: "mihomo standby account",
			account: &Account{ID: 3, Extra: map[string]any{
				"mihomo_pool_managed": true,
				"mihomo_pool_standby": true,
			}},
		},
		{
			name: "mihomo managed account without egress",
			account: &Account{ID: 4, Extra: map[string]any{
				"mihomo_pool_managed": true,
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proxyURL, err := resolvedAccountProxyURL(test.account)
			require.Error(t, err)
			require.Empty(t, proxyURL)
		})
	}
}

func TestResolvedAccountProxyURLAllowsDirectAndBoundAccounts(t *testing.T) {
	directURL, err := resolvedAccountProxyURL(&Account{ID: 1})
	require.NoError(t, err)
	require.Empty(t, directURL)

	proxyID := int64(7)
	bound := &Account{
		ID:      2,
		ProxyID: &proxyID,
		Proxy:   &Proxy{ID: proxyID, Protocol: "http", Host: "127.0.0.1", Port: 8080},
	}
	proxyURL, err := resolvedAccountProxyURL(bound)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:8080", proxyURL)
}

func TestHydrateSelectedAccountRejectsInvalidEgressWithoutSnapshot(t *testing.T) {
	proxyID := int64(7)
	account := &Account{ID: 1, ProxyID: &proxyID}

	_, err := (&OpenAIGatewayService{}).hydrateSelectedAccount(context.Background(), account)
	require.ErrorContains(t, err, "invalid egress")

	_, err = (&GatewayService{}).hydrateSelectedAccount(context.Background(), account)
	require.ErrorContains(t, err, "invalid egress")
}
