package main

import "testing"

func TestGetAppInfo(t *testing.T) {
	app := NewApp()

	got, err := app.GetAppInfo()
	if err != nil {
		t.Fatalf("GetAppInfo() returned an error: %v", err)
	}

	want := AppInfo{Name: "StudyDesk", Version: "0.1.0"}
	if got != want {
		t.Errorf("GetAppInfo() = %+v, want %+v", got, want)
	}
}
