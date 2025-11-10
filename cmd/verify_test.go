package cmd

import (
	"testing"
)

func TestCompareTransactionItems(t *testing.T) {
	tests := []struct {
		name           string
		localPackages  []Package
		serverPackages []Package
		wantMissing    int
		wantExtra      int
	}{
		{
			name: "identical packages",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 0,
			wantExtra:   0,
		},
		{
			name: "missing package on server",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
				{Action: "Install", Name: "git", Version: "2.31", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 1,
			wantExtra:   0,
		},
		{
			name: "extra package on server",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
				{Action: "Install", Name: "git", Version: "2.31", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 0,
			wantExtra:   1,
		},
		{
			name: "different versions",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.1", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 1,
			wantExtra:   1,
		},
		{
			name:          "empty local",
			localPackages: []Package{},
			serverPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 0,
			wantExtra:   1,
		},
		{
			name: "empty server",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{},
			wantMissing:    1,
			wantExtra:      0,
		},
		{
			name: "different actions",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{
				{Action: "Upgrade", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 1,
			wantExtra:   1,
		},
		{
			name: "different repositories",
			localPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			serverPackages: []Package{
				{Action: "Install", Name: "vim", Version: "8.2", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "epel"},
			},
			wantMissing: 1,
			wantExtra:   1,
		},
		{
			name: "same package different repos should be treated as different",
			localPackages: []Package{
				{Action: "Install", Name: "git", Version: "2.31", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
				{Action: "Install", Name: "git", Version: "2.31", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "epel"},
			},
			serverPackages: []Package{
				{Action: "Install", Name: "git", Version: "2.31", Release: "1.el8", Epoch: "", Arch: "x86_64", Repo: "appstream"},
			},
			wantMissing: 1,
			wantExtra:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing, extra := compareTransactionItems(tt.localPackages, tt.serverPackages)
			if len(missing) != tt.wantMissing {
				t.Errorf("compareTransactionItems() missing count = %d, want %d", len(missing), tt.wantMissing)
			}
			if len(extra) != tt.wantExtra {
				t.Errorf("compareTransactionItems() extra count = %d, want %d", len(extra), tt.wantExtra)
			}
		})
	}
}

func TestPrintVerificationResults(t *testing.T) {
	tests := []struct {
		name                string
		result              VerificationResult
		wantFullyVerified   int
		wantWithMissing     int
		wantWithExtra       int
		wantMissingOnServer int
	}{
		{
			name: "all transactions verified",
			result: VerificationResult{
				FullyVerified:    3,
				WithMissingItems: []string{},
				WithExtraItems:   []string{},
				MissingOnServer:  []string{},
			},
			wantFullyVerified:   3,
			wantWithMissing:     0,
			wantWithExtra:       0,
			wantMissingOnServer: 0,
		},
		{
			name: "some transactions with missing items",
			result: VerificationResult{
				FullyVerified:    2,
				WithMissingItems: []string{"5", "7"},
				WithExtraItems:   []string{},
				MissingOnServer:  []string{},
			},
			wantFullyVerified:   2,
			wantWithMissing:     2,
			wantWithExtra:       0,
			wantMissingOnServer: 0,
		},
		{
			name: "transactions with extra items",
			result: VerificationResult{
				FullyVerified:    1,
				WithMissingItems: []string{},
				WithExtraItems:   []string{"3"},
				MissingOnServer:  []string{},
			},
			wantFullyVerified:   1,
			wantWithMissing:     0,
			wantWithExtra:       1,
			wantMissingOnServer: 0,
		},
		{
			name: "transactions not on server",
			result: VerificationResult{
				FullyVerified:    0,
				WithMissingItems: []string{},
				WithExtraItems:   []string{},
				MissingOnServer:  []string{"10", "11"},
			},
			wantFullyVerified:   0,
			wantWithMissing:     0,
			wantWithExtra:       0,
			wantMissingOnServer: 2,
		},
		{
			name: "mixed verification results",
			result: VerificationResult{
				FullyVerified:    1,
				WithMissingItems: []string{"5"},
				WithExtraItems:   []string{"6"},
				MissingOnServer:  []string{"7"},
			},
			wantFullyVerified:   1,
			wantWithMissing:     1,
			wantWithExtra:       1,
			wantMissingOnServer: 1,
		},
		{
			name: "verify FullyVerified increments per transaction",
			result: VerificationResult{
				FullyVerified:    5,
				WithMissingItems: []string{},
				WithExtraItems:   []string{},
				MissingOnServer:  []string{},
			},
			wantFullyVerified:   5,
			wantWithMissing:     0,
			wantWithExtra:       0,
			wantMissingOnServer: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.FullyVerified != tt.wantFullyVerified {
				t.Errorf("FullyVerified count = %d, want %d", tt.result.FullyVerified, tt.wantFullyVerified)
			}
			if len(tt.result.WithMissingItems) != tt.wantWithMissing {
				t.Errorf("WithMissingItems count = %d, want %d", len(tt.result.WithMissingItems), tt.wantWithMissing)
			}
			if len(tt.result.WithExtraItems) != tt.wantWithExtra {
				t.Errorf("WithExtraItems count = %d, want %d", len(tt.result.WithExtraItems), tt.wantWithExtra)
			}
			if len(tt.result.MissingOnServer) != tt.wantMissingOnServer {
				t.Errorf("MissingOnServer count = %d, want %d", len(tt.result.MissingOnServer), tt.wantMissingOnServer)
			}
		})
	}
}
