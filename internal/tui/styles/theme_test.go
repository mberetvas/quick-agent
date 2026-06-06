package styles

import "testing"

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme.Header.Render("x") == "" {
		t.Fatal("expected populated default theme")
	}
}

func TestLightTheme(t *testing.T) {
	theme := LightTheme()
	if theme.Title.Render("x") == "" {
		t.Fatal("expected populated light theme")
	}
}

func TestThemeForConfig(t *testing.T) {
	if ThemeForConfig("light").Primary != LightTheme().Primary {
		t.Error("light config should return LightTheme")
	}
	if ThemeForConfig("dark").Primary != DefaultTheme().Primary {
		t.Error("dark config should return DefaultTheme")
	}
	if ThemeForConfig("auto").Primary != DefaultTheme().Primary {
		t.Error("auto config should return DefaultTheme")
	}
}
