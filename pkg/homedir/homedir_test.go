package homedir

import "testing"

func TestHomedir(t *testing.T) {
	home := HomeDir()

	if path := ExpandPath("~"); path != home {
		t.Errorf("homedir.Expand(~): expected %q but got %q", home, path)
	}
	if path := ExpandPath("~/"); path != home {
		t.Errorf("homedir.Expand(~/): expected %q but got %q", home, path)
	}
	if path := ExpandPath("~/Desktop"); path != home+"/Desktop" {
		t.Errorf("homedir.Expand(~): expected %q but got %q", home+"/Desktop", path)
	}
}
