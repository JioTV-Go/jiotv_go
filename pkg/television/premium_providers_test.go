package television

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// TestExtractPremiumProvidersFromPackageInfo covers the current plans API shape
// (v2.1/plans/get): PackageInfo[].packageDetail.providers[] with id/name/image.
func TestExtractPremiumProvidersFromPackageInfo(t *testing.T) {
	plansResponse := PlansResponse{
		PackageInfo: []ActiveSubscriptionPlan{
			{
				PlanID:   "1019037",
				IsActive: true,
				PackageDetail: ActiveSubscriptionPack{
					PackageName: "Sports Pack",
					Providers: []PlanProvider{
						{ID: "Z0177", Name: "FanCode", Image: "fancode.png"},
					},
				},
			},
			{
				PlanID:   "1019038",
				IsActive: true,
				PackageDetail: ActiveSubscriptionPack{
					Providers: []PlanProvider{
						{ID: "Z0177", Name: "FanCode"},
					},
				},
			},
		},
	}

	premiumProviders := extractPremiumProviders(plansResponse)
	if len(premiumProviders) != 1 {
		t.Fatalf("expected 1 premium provider, got %d", len(premiumProviders))
	}
	if premiumProviders[0].ProviderID != "Z0177" {
		t.Fatalf("expected canonical provider ID Z0177, got %s", premiumProviders[0].ProviderID)
	}
	if premiumProviders[0].Name != "FanCode" {
		t.Fatalf("expected provider name FanCode, got %s", premiumProviders[0].Name)
	}
	if premiumProviders[0].Image != "fancode.png" {
		t.Fatalf("expected provider image fancode.png, got %s", premiumProviders[0].Image)
	}
	if premiumProviders[0].URL != "https://www.fancode.com/" {
		t.Fatalf("expected FanCode URL, got %s", premiumProviders[0].URL)
	}
}

// TestExtractPremiumProvidersKeepsUnknownProviders ensures providers absent from
// the local registry are still returned, since the plans API only lists
// providers the account is entitled to.
func TestExtractPremiumProvidersKeepsUnknownProviders(t *testing.T) {
	plansResponse := PlansResponse{
		PackageInfo: []ActiveSubscriptionPlan{
			{
				IsActive: true,
				PackageDetail: ActiveSubscriptionPack{
					Providers: []PlanProvider{
						{ID: "Z0999", Name: "Some Other Provider"},
					},
				},
			},
		},
	}

	premiumProviders := extractPremiumProviders(plansResponse)
	if len(premiumProviders) != 1 {
		t.Fatalf("expected 1 premium provider, got %d", len(premiumProviders))
	}
	if premiumProviders[0].Name != "Some Other Provider" {
		t.Fatalf("expected provider name to be preserved, got %s", premiumProviders[0].Name)
	}
	if premiumProviders[0].URL != "" {
		t.Fatalf("expected no URL for unknown provider, got %s", premiumProviders[0].URL)
	}
}

// TestExtractPremiumProvidersWithoutIsActive covers responses that omit the
// "isactive" flag; every listed package should then be considered.
func TestExtractPremiumProvidersWithoutIsActive(t *testing.T) {
	plansResponse := PlansResponse{
		PackageInfo: []ActiveSubscriptionPlan{
			{
				PackageDetail: ActiveSubscriptionPack{
					Providers: []PlanProvider{
						{ID: "Z0177", Name: "FanCode"},
					},
				},
			},
		},
	}

	premiumProviders := extractPremiumProviders(plansResponse)
	if len(premiumProviders) != 1 {
		t.Fatalf("expected 1 premium provider, got %d", len(premiumProviders))
	}
}

// providerDirectoryFixture mirrors the real /cnf/provider directory: keys are
// normalised provider taxonomy names, values are content provider IDs.
var providerDirectoryFixture = map[string]string{
	"fancode":       "200169",
	"jiocinema":     "200006",
	"sonyliv":       "200200",
	"zee5":          "200022",
	"sunnxt":        "200138",
	"discoveryplus": "200141",
	"lionsgate":     "200140",
	"kancchalannka": "206666",
	"etvwin":        "201541",
	"tarangplus":    "300010",
	"shemaroome":    "200152",
	"isj":           "2",
}

// TestResolveContentProviderIDsRealWorldNames covers the plan-name spellings
// seen in live plan data, which differ from the directory's taxonomy names.
func TestResolveContentProviderIDsRealWorldNames(t *testing.T) {
	testCases := []struct {
		planName           string
		taxonomyName       string
		expectedProviderID string
	}{
		// Exact taxonomy match, display name carries a qualifier.
		{"Fancode (Via JioTV)", "Fancode", "200169"},
		// No taxonomy name at all: fall back to the display name.
		{"SonyLIV", "", "200200"},
		{"Sony LIV", "", "200200"},
		{"Zee5 (sports excluded)", "", "200022"},
		{"SunNXT", "", "200138"},
		{"ETV Win", "", "201541"},
		{"Tarang Plus", "", "300010"},
		// "+" style differences.
		{"Discovery+", "", "200141"},
		// Extra qualifier resolved by containment.
		{"JioCinema Premium", "", "200006"},
		{"LionsgatePlay", "", "200140"},
	}

	for _, testCase := range testCases {
		provider := PremiumProvider{ID: "X", Name: testCase.planName}
		if testCase.taxonomyName != "" {
			provider.matchNames = []string{testCase.taxonomyName, testCase.planName}
		}
		premiumProviders := []PremiumProvider{provider}

		resolveContentProviderIDs(premiumProviders, providerDirectoryFixture)

		if premiumProviders[0].ProviderID != testCase.expectedProviderID {
			t.Errorf("plan name %q: expected content provider ID %s, got %q",
				testCase.planName, testCase.expectedProviderID, premiumProviders[0].ProviderID)
		}
	}
}

// TestResolveContentProviderIDsKeepsEntitlementIDWhenUnmatched ensures an
// unknown provider is left alone rather than mapped to something wrong.
func TestResolveContentProviderIDsKeepsEntitlementIDWhenUnmatched(t *testing.T) {
	premiumProviders := []PremiumProvider{
		{ID: "Z9999", ProviderID: "Z9999", Name: "Totally Unknown Service"},
	}

	resolveContentProviderIDs(premiumProviders, providerDirectoryFixture)

	if premiumProviders[0].ProviderID != "Z9999" {
		t.Fatalf("expected unmatched provider to keep entitlement ID, got %q", premiumProviders[0].ProviderID)
	}
}

// TestResolveContentProviderIDsShortNameNotOverMatched guards the length limit
// in the containment fallback: a short directory name such as "isj" must not
// swallow unrelated providers.
func TestResolveContentProviderIDsShortNameNotOverMatched(t *testing.T) {
	premiumProviders := []PremiumProvider{
		{ID: "Z0001", ProviderID: "Z0001", Name: "Disj"},
	}

	resolveContentProviderIDs(premiumProviders, providerDirectoryFixture)

	if premiumProviders[0].ProviderID != "Z0001" {
		t.Fatalf("expected no match for short name, got %q", premiumProviders[0].ProviderID)
	}
}

func TestNormalizeProviderName(t *testing.T) {
	testCases := map[string]string{
		"Fancode":        "fancode",
		"Discovery+":     "discoveryplus",
		"Tarang Plus":    "tarangplus",
		"Sony LIV":       "sonyliv",
		"ETV Win":        "etvwin",
		"jiotv_aajtak":   "jiotvaajtak",
		"  Zee5  ":       "zee5",
		"Kanchha-Lannka": "kanchhalannka",
	}

	for input, expected := range testCases {
		if actual := normalizeProviderName(input); actual != expected {
			t.Errorf("normalizeProviderName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

// TestExtractPremiumProvidersLegacyShape covers the older nested response shape.
func TestExtractPremiumProvidersLegacyShape(t *testing.T) {
	plansResponse := PlansResponse{
		Result: PlansResult{
			Plans: []Plan{
				{
					Providers: []PlanProvider{
						{ProviderID: "", ProviderName: "JioCinema Premium"},
					},
				},
			},
		},
	}

	premiumProviders := extractPremiumProviders(plansResponse)
	if len(premiumProviders) != 1 {
		t.Fatalf("expected 1 premium provider, got %d", len(premiumProviders))
	}
	if premiumProviders[0].Name != "JioCinema Premium" {
		t.Fatalf("expected provider name JioCinema Premium, got %s", premiumProviders[0].Name)
	}
	if premiumProviders[0].URL != "https://www.jiocinema.com/" {
		t.Fatalf("expected JioCinema URL, got %s", premiumProviders[0].URL)
	}
}

func TestExtractPremiumProvidersFromAccessToken(t *testing.T) {
	claims := map[string]interface{}{
		"data": map[string]interface{}{
			"extra": "{\"plandetails\":{\"PackageInfo\":[{\"planid\":\"1019037\"},{\"planid\":\"1\"}]}}",
		},
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}

	accessToken := "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
	premiumProviders := extractPremiumProvidersFromAccessToken(accessToken)

	if len(premiumProviders) != 1 {
		t.Fatalf("expected 1 premium provider from token, got %d", len(premiumProviders))
	}
	if premiumProviders[0].Name != "FanCode" {
		t.Fatalf("expected FanCode from token, got %s", premiumProviders[0].Name)
	}
	if premiumProviders[0].ProviderID != "Z0177" {
		t.Fatalf("expected canonical provider ID Z0177 from token, got %s", premiumProviders[0].ProviderID)
	}
	if premiumProviders[0].URL != "https://www.fancode.com/" {
		t.Fatalf("expected FanCode URL from token, got %s", premiumProviders[0].URL)
	}
}
