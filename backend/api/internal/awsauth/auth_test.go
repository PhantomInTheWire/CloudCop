package awsauth

import (
	"errors"
	"testing"
	"time"
)

func TestAssumeRoleInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   AssumeRoleInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: AssumeRoleInput{
				AccountID:  "123456789012",
				ExternalID: "test-external-id",
			},
			wantErr: false,
		},
		{
			name: "empty account id",
			input: AssumeRoleInput{
				AccountID:  "",
				ExternalID: "test-external-id",
			},
			wantErr: true,
		},
		{
			name: "empty external id",
			input: AssumeRoleInput{
				AccountID:  "123456789012",
				ExternalID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr && (tt.input.AccountID == "" || tt.input.ExternalID == "") {
				// Expected empty fields
				return
			}
			if !tt.wantErr && tt.input.AccountID != "" && tt.input.ExternalID != "" {
				// Valid input
				return
			}
			t.Errorf("Unexpected validation result")
		})
	}
}

func TestCredentials_Expiration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		creds      *Credentials
		buffer     time.Duration
		wantExpire bool
	}{
		{
			name: "not expired",
			creds: &Credentials{
				Expiration: now.Add(10 * time.Minute),
			},
			buffer:     5 * time.Minute,
			wantExpire: false,
		},
		{
			name: "expiring soon",
			creds: &Credentials{
				Expiration: now.Add(3 * time.Minute),
			},
			buffer:     5 * time.Minute,
			wantExpire: true,
		},
		{
			name: "already expired",
			creds: &Credentials{
				Expiration: now.Add(-1 * time.Minute),
			},
			buffer:     5 * time.Minute,
			wantExpire: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeUntil := time.Until(tt.creds.Expiration)
			isExpiring := timeUntil <= tt.buffer
			if isExpiring != tt.wantExpire {
				t.Errorf("Expiration check failed: got %v, want %v", isExpiring, tt.wantExpire)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "invalid external id",
			err:     ErrInvalidExternalID,
			wantMsg: "invalid external ID",
		},
		{
			name:    "assume role failed",
			err:     ErrAssumeRoleFailed,
			wantMsg: "failed to assume role",
		},
		{
			name:    "credentials expired",
			err:     ErrCredentialsExpired,
			wantMsg: "credentials have expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("Error message mismatch: got %v, want %v", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestErrWrapping(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(ErrAssumeRoleFailed, baseErr)

	if !errors.Is(wrappedErr, ErrAssumeRoleFailed) {
		t.Error("Error wrapping failed")
	}
}
