package validator

import (
	"regexp"
	"testing"
)

func TestIsValidOrderNumber(t *testing.T) {
	type args struct {
		number string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid Luhan number",
			args: args{
				number: "9278923470",
			},
			want: true,
		},
		{
			name: "invalid Luhan number",
			args: args{
				number: "23772224",
			},
			want: false,
		},
		{
			name: "empty string passed",
			args: args{
				number: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidOrderNumber(tt.args.number); got != tt.want {
				t.Errorf("IsValidOrderNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	type args struct {
		values []any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Unique integers",
			args: args{values: []any{1, 2, 3, 4, 5}},
			want: true,
		},
		{
			name: "Non-unique integers",
			args: args{values: []any{1, 2, 3, 3, 5}},
			want: false,
		},
		{
			name: "Unique strings",
			args: args{values: []any{"apple", "banana", "cherry"}},
			want: true,
		},
		{
			name: "Non-unique strings",
			args: args{values: []any{"apple", "banana", "banana"}},
			want: false,
		},
		{
			name: "Empty slice",
			args: args{values: []any{}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.args.values); got != tt.want {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	type args struct {
		value string
		rx    *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid email address",
			args: args{value: "mihai@example.com", rx: EmailRX},
			want: true,
		},
		{
			name: "invalid email address, missing @ symbol",
			args: args{value: "mihaiexample.com", rx: EmailRX},
			want: false,
		},
		{
			name: "invalid email address, incorrect domain",
			args: args{value: "mihai@examplecom", rx: EmailRX},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.args.value, tt.args.rx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermittedValue(t *testing.T) {
	type args struct {
		value           any
		permittedValues []any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "allowed value found in permitted value",
			args: args{value: "admin", permittedValues: []any{"admin", "super-admin"}},
			want: true,
		},
		{
			name: "value not found in permitted value",
			args: args{value: "user", permittedValues: []any{"admin", "super-admin"}},
			want: false,
		},
		{
			name: "empty value and not empty permitted values",
			args: args{value: "", permittedValues: []any{"admin", "super-admin"}},
			want: false,
		},
		{
			name: "empty value and not empty permitted values",
			args: args{value: "", permittedValues: []any{}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PermittedValue(tt.args.value, tt.args.permittedValues...); got != tt.want {
				t.Errorf("PermittedValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_Valid(t *testing.T) {
	type args struct {
		value string
		rx    *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid email address",
			args: args{value: "mihai@example.com", rx: EmailRX},
			want: true,
		},
		{
			name: "invalid email address, missing @ symbol",
			args: args{value: "mihaiexample.com", rx: EmailRX},
			want: false,
		},
		{
			name: "invalid email address, incorrect domain",
			args: args{value: "mihai@examplecom", rx: EmailRX},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()

			v.Check(Matches(tt.args.value, tt.args.rx), "email", "invalid email address")

			if !tt.want && v.Valid() {
				t.Error("validater received an error although should not have any validation errors")
			}
		})
	}
}
