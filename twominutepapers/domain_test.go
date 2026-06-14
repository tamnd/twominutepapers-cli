package twominutepapers

import (
	"testing"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "twominutepapers" {
		t.Errorf("Scheme = %q, want twominutepapers", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "twominutepapers" {
		t.Errorf("Identity.Binary = %q, want twominutepapers", info.Identity.Binary)
	}
}
