package rbac

import "testing"

func TestAllowed(t *testing.T) {
	if !Allowed("admin", "anything") {
		t.Fatal("admin should be allowed")
	}
	if !Allowed("analyst", "alerts:write") {
		t.Fatal("analyst should ack alerts")
	}
	if !Allowed("analyst", "replay:write") {
		t.Fatal("analyst should trigger replay")
	}
	if Allowed("analyst", "model:deploy") {
		t.Fatal("analyst must not deploy models")
	}
	if Allowed("viewer", "alerts:write") {
		t.Fatal("viewer must not acknowledge alerts")
	}
	if !Allowed("viewer", "read") {
		t.Fatal("viewer should be read-only")
	}
	if Allowed("viewer", "replay:write") {
		t.Fatal("viewer must not trigger replay")
	}
}
