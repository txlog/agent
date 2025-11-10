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
