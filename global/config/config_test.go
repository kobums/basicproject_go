package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	oldSecret, oldName, oldMode := JwtSecret, Database.Name, Mode
	t.Cleanup(func() { JwtSecret, Database.Name, Mode = oldSecret, oldName, oldMode })

	Database.Name = "project"
	JwtSecret = "too-short"
	if Validate() == nil {
		t.Error("short JWT_SECRET accepted")
	}
	JwtSecret = "0123456789abcdef0123456789abcdef"
	if err := Validate(); err != nil {
		t.Error(err)
	}

	JwtSecret = "change-me-to-a-long-random-string-32chars"
	Mode = "develop"
	if err := Validate(); err != nil {
		t.Errorf("develop placeholder: %v", err)
	}
	Mode = "production"
	if Validate() == nil {
		t.Error("production accepted the placeholder secret")
	}
}

func TestLoadDotEnv(t *testing.T) {
	file := filepath.Join(t.TempDir(), ".env")
	content := "# comment\nBP_TEST_A=1\nexport BP_TEST_B=\"two words\"\nBP_TEST_C=from-file\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BP_TEST_C", "from-env")
	os.Unsetenv("BP_TEST_A")
	os.Unsetenv("BP_TEST_B")
	t.Cleanup(func() { os.Unsetenv("BP_TEST_A"); os.Unsetenv("BP_TEST_B") })

	loadDotEnv(file)
	if os.Getenv("BP_TEST_A") != "1" || os.Getenv("BP_TEST_B") != "two words" {
		t.Errorf("A=%q B=%q", os.Getenv("BP_TEST_A"), os.Getenv("BP_TEST_B"))
	}
	if os.Getenv("BP_TEST_C") != "from-env" {
		t.Errorf(".env overrode the environment: C=%q", os.Getenv("BP_TEST_C"))
	}
}

func TestSplitList(t *testing.T) {
	got := splitList(" http://a , ,http://b,")
	if len(got) != 2 || got[0] != "http://a" || got[1] != "http://b" {
		t.Errorf("splitList = %q", got)
	}
}
