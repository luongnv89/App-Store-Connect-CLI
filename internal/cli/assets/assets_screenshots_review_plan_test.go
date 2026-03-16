package assets

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	reviewshots "github.com/rudrankriyam/App-Store-Connect-CLI/internal/screenshots"
)

func TestExecuteScreenshotReviewPlanPlanModeReturnsBeforeUploadWhenBlockingIssuesExist(t *testing.T) {
	setupAssetsPlanAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	outputDir := t.TempDir()
	filePath := writeAssetsTestPNG(t, outputDir, "01-home.png")

	manifestPath := filepath.Join(outputDir, defaultReviewManifestFile)
	manifest := reviewshots.ReviewManifest{
		GeneratedAt: "2026-03-16T00:00:00Z",
		FramedDir:   outputDir,
		OutputDir:   outputDir,
		Entries: []reviewshots.ReviewEntry{
			{
				Key:               "ready-entry",
				ScreenshotID:      "home",
				Locale:            "en-US",
				FramedPath:        filePath,
				FramedRelative:    "01-home.png",
				DisplayTypes:      []string{"APP_IPHONE_65"},
				ValidAppStoreSize: true,
				Status:            "ready",
			},
			{
				Key:               "blocked-entry",
				ScreenshotID:      "details",
				Locale:            "en-US",
				FramedPath:        filePath,
				FramedRelative:    "01-home.png",
				DisplayTypes:      []string{"APP_IPHONE_65"},
				ValidAppStoreSize: true,
				Status:            "invalid-size",
			},
		},
	}
	writeAssetsReviewManifest(t, manifestPath, manifest)
	if err := reviewshots.SaveApprovals(filepath.Join(outputDir, defaultReviewApprovalFile), map[string]bool{
		"ready-entry":   true,
		"blocked-entry": true,
	}); err != nil {
		t.Fatalf("SaveApprovals() error: %v", err)
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = assetsUploadRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/v1/appStoreVersions/version-123":
			return assetsJSONResponse(http.StatusOK, `{
				"data": {
					"type": "appStoreVersions",
					"id": "version-123",
					"attributes": {
						"versionString": "1.2.3",
						"platform": "IOS"
					},
					"relationships": {
						"app": {
							"data": {
								"type": "apps",
								"id": "123456789"
							}
						}
					}
				}
			}`)
		case req.Method == http.MethodGet && req.URL.Path == "/v1/appStoreVersions/version-123/appStoreVersionLocalizations":
			return assetsJSONResponse(http.StatusOK, `{
				"data": [
					{
						"type": "appStoreVersionLocalizations",
						"id": "loc-en",
						"attributes": {
							"locale": "en-US"
						}
					}
				],
				"links": {}
			}`)
		case req.Method == http.MethodGet && req.URL.Path == "/v1/appStoreVersionLocalizations/loc-en/appScreenshotSets":
			t.Fatal("plan mode with blocking issues must not fetch screenshot sets")
			return nil, nil
		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})
	t.Cleanup(func() {
		http.DefaultTransport = origTransport
	})

	result, err := executeScreenshotReviewPlan(context.Background(), screenshotReviewPlanOptions{
		AppID:           "123456789",
		VersionID:       "version-123",
		Platform:        "IOS",
		ReviewOutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("executeScreenshotReviewPlan() error: %v", err)
	}

	if result.ErrorCount == 0 {
		t.Fatal("expected blocking issue count to be reported")
	}
	if result.PlannedGroups != 1 {
		t.Fatalf("expected one planned group, got %d", result.PlannedGroups)
	}
	if len(result.Groups) != 1 {
		t.Fatalf("expected one returned group without uploads, got %d", len(result.Groups))
	}
	if result.Groups[0].Result.SetID != "" || len(result.Groups[0].Result.Results) != 0 {
		t.Fatalf("expected plan mode to skip upload results when blocking issues exist, got %+v", result.Groups[0].Result)
	}
}

func setupAssetsPlanAuth(t *testing.T) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if pemBytes == nil {
		t.Fatal("encode pem: nil")
	}

	t.Setenv("ASC_BYPASS_KEYCHAIN", "1")
	t.Setenv("ASC_KEY_ID", "KEY_ID")
	t.Setenv("ASC_ISSUER_ID", "ISSUER_ID")
	t.Setenv("ASC_PRIVATE_KEY", string(pemBytes))
}

func writeAssetsReviewManifest(t *testing.T, path string, manifest reviewshots.ReviewManifest) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
}
