package cmd

import (
	"testing"
)

func TestReadSecretInput_WithArg(t *testing.T) {
	got, err := readSecretInput([]string{"hello"}, "Text: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestReadPlainInput_WithArg(t *testing.T) {
	got, err := readPlainInput([]string{"cipher"}, "Ciphertext: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "cipher" {
		t.Errorf("got %q, want %q", got, "cipher")
	}
}

func TestReadSecretInput_MultipleArgsUsesFirst(t *testing.T) {
	// 理论上 cobra.MaximumNArgs(1) 会拦掉，但函数本身只取第一个
	got, err := readSecretInput([]string{"first", "second"}, "Text: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "first" {
		t.Errorf("got %q, want %q", got, "first")
	}
}

func TestReadPlainInput_MultipleArgsUsesFirst(t *testing.T) {
	got, err := readPlainInput([]string{"first", "second"}, "Ciphertext: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "first" {
		t.Errorf("got %q, want %q", got, "first")
	}
}
