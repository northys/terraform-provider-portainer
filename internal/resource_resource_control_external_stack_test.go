package internal

import (
	"net/http"
	"testing"
)

// TestResourceControl_ExternalStackLifecycle covers a stack deployed outside
// Portainer: its resource ID is "<endpointId>_<compose project>", which
// GET /stacks/{id} cannot resolve. Create must POST the control, Read must keep
// it in state, and Delete must remove it by the stored ID.
func TestResourceControl_ExternalStackLifecycle(t *testing.T) {
	mock := NewMockServer(t)

	mock.On("POST", "/resource_controls", RespondJSON(http.StatusOK, map[string]interface{}{"Id": 17}))
	mock.On("DELETE", "/resource_controls/17", RespondString(http.StatusNoContent, "application/json", ``))

	r := resourceResourceControl()
	d := r.TestResourceData()
	_ = d.Set("resource_id", "4_server")
	_ = d.Set("type", 6)
	_ = d.Set("teams", []interface{}{3})

	if err := rcCreate(r, d, mock.Client()); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if d.Id() != "17" {
		t.Fatalf("expected ID %q after create, got %q", "17", d.Id())
	}

	post := mock.FindRequest("POST", "/resource_controls")
	if post == nil {
		t.Fatal("expected POST /resource_controls")
	}
	var payload map[string]interface{}
	if err := post.DecodeJSON(&payload); err != nil {
		t.Fatalf("failed to decode POST body: %v", err)
	}
	if payload["ResourceID"] != "4_server" || payload["Type"] != float64(6) {
		t.Errorf("unexpected POST payload: %v", payload)
	}

	if err := rcRead(r, d, mock.Client()); err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if d.Id() != "17" {
		t.Errorf("Read dropped the external stack control from state (ID %q)", d.Id())
	}
	if mock.FindRequest("GET", "/stacks/4_server") != nil {
		t.Error("Read must not look up an external stack through GET /stacks/{id}")
	}

	if err := rcDelete(r, d, mock.Client()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if mock.FindRequest("DELETE", "/resource_controls/17") == nil {
		t.Error("expected DELETE /resource_controls/17")
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after delete, got %q", d.Id())
	}
}
