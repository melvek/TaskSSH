package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	inv, err := Load("../../testdata/inventory.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 全局变量
	if inv.GlobalVars.Port != 22 {
		t.Errorf("expected port 22, got %d", inv.GlobalVars.Port)
	}
	if inv.GlobalVars.Username != "deploy" {
		t.Errorf("expected username deploy, got %s", inv.GlobalVars.Username)
	}
	if inv.GlobalVars.Extra["app_name"] != "my-app" {
		t.Errorf("expected app_name my-app, got %v", inv.GlobalVars.Extra["app_name"])
	}
	if _, ok := inv.GlobalVars.Extra["date"]; !ok {
		t.Error("expected date injected")
	}

	// 服务器组
	prod, ok := inv.Servers["prod"]
	if !ok {
		t.Fatal("expected group prod")
	}
	if len(prod.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(prod.Hosts))
	}
	if prod.Hosts["web_1"].Host != "192.168.1.1" {
		t.Errorf("expected web_1 host 192.168.1.1, got %s", prod.Hosts["web_1"].Host)
	}
	if prod.Hosts["web_2"].Port != 2222 {
		t.Errorf("expected web_2 port 2222, got %d", prod.Hosts["web_2"].Port)
	}

	// 任务
	release, ok := inv.Tasks["release"]
	if !ok {
		t.Fatal("expected task release")
	}
	if len(release.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(release.Steps))
	}
	if release.Steps[0].Action != "push" {
		t.Errorf("expected action push, got %s", release.Steps[0].Action)
	}
	if release.Steps[1].Delay != 5 {
		t.Errorf("expected delay 5, got %d", release.Steps[1].Delay)
	}
}
