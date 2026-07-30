package config

import "testing"

func TestCheckExposureRefusesAPublicAddressWithoutAToken(t *testing.T) {
	if err := CheckExposure("0.0.0.0:7080", ""); err == nil {
		t.Fatal("want an error for a public address with no token")
	}
	if err := CheckExposure(":7080", ""); err == nil {
		t.Fatal("want an error for an empty host, which binds every interface")
	}
}

func TestCheckExposureAllowsAPublicAddressWithAToken(t *testing.T) {
	if err := CheckExposure("0.0.0.0:7080", "secret"); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestCheckExposureAllowsLoopbackWithoutAToken(t *testing.T) {
	if err := CheckExposure("127.0.0.1:7080", ""); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if err := CheckExposure("localhost:7080", ""); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}
