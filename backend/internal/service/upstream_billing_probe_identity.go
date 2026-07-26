package service

func upstreamBillingProbeIdentity(account *Account) map[string]any {
	if account == nil {
		return nil
	}
	identity := map[string]any{"platform": account.Platform, "type": account.Type, "proxy_id": nil}
	if account.ProxyID != nil {
		identity["proxy_id"] = *account.ProxyID
	}
	for _, key := range []string{"api_key", "base_url", credKeyHeaderOverrideEnabled, credKeyHeaderOverrides} {
		if value, ok := account.Credentials[key]; ok {
			identity[key] = value
		}
	}
	return identity
}
