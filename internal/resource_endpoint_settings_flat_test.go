package internal

import (
	"net/http"
	"testing"
)

// TestEndpointSettingsUpdate_SendsSecuritySwitchesTopLevel verifies that the
// security switches reach PUT /endpoints/{id}/settings as top-level fields,
// which is where Portainer CE reads them, including the ones set to false.
func TestEndpointSettingsUpdate_SendsSecuritySwitchesTopLevel(t *testing.T) {
	mock := NewMockServer(t)

	mock.On("PUT", "/endpoints/7/settings", RespondJSON(http.StatusOK, map[string]interface{}{}))

	r := resourceEndpointSettings()
	d := r.TestResourceData()
	_ = d.Set("endpoint_id", 7)
	_ = d.Set("security_settings", []interface{}{
		map[string]interface{}{
			"allow_bind_mounts":            false,
			"allow_container_capabilities": false,
			"allow_device_mapping":         false,
			"allow_host_namespace":         false,
			"allow_privileged_mode":        false,
			"allow_stack_management":       false,
			"allow_sysctl_setting":         false,
			"allow_volume_browser":         true,
			"enable_host_management":       false,
		},
	})

	if err := rcCreate(r, d, mock.Client()); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	put := mock.FindRequest("PUT", "/endpoints/7/settings")
	if put == nil {
		t.Fatal("expected PUT /endpoints/7/settings")
	}
	var payload map[string]interface{}
	if err := put.DecodeJSON(&payload); err != nil {
		t.Fatalf("failed to decode PUT body: %v", err)
	}

	want := map[string]bool{
		"allowBindMountsForRegularUsers":            false,
		"allowContainerCapabilitiesForRegularUsers": false,
		"allowDeviceMappingForRegularUsers":         false,
		"allowHostNamespaceForRegularUsers":         false,
		"allowPrivilegedModeForRegularUsers":        false,
		"allowStackManagementForRegularUsers":       false,
		"allowSysctlSettingForRegularUsers":         false,
		"allowVolumeBrowserForRegularUsers":         true,
		"enableHostManagementFeatures":              false,
	}
	for key, value := range want {
		got, ok := payload[key]
		if !ok {
			t.Errorf("%s missing from the top level of the payload", key)
			continue
		}
		if got != value {
			t.Errorf("%s: expected %v, got %v", key, value, got)
		}
	}
}

// TestEndpointSettingsUpdate_NoSecurityBlockSendsNoSwitches verifies that
// without a security_settings block no switch is sent, so Portainer keeps the
// environment's current values.
func TestEndpointSettingsUpdate_NoSecurityBlockSendsNoSwitches(t *testing.T) {
	mock := NewMockServer(t)

	mock.On("PUT", "/endpoints/7/settings", RespondJSON(http.StatusOK, map[string]interface{}{}))

	r := resourceEndpointSettings()
	d := r.TestResourceData()
	_ = d.Set("endpoint_id", 7)

	if err := rcCreate(r, d, mock.Client()); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var payload map[string]interface{}
	if err := mock.FindRequest("PUT", "/endpoints/7/settings").DecodeJSON(&payload); err != nil {
		t.Fatalf("failed to decode PUT body: %v", err)
	}
	if _, ok := payload["allowStackManagementForRegularUsers"]; ok {
		t.Errorf("expected no security switches without a security_settings block, got %v", payload)
	}
}
