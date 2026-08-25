package provision

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/sass"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/system/types"
	"go.uber.org/zap"
)

// updateWebappTheme is a function that provisions new webapp themes,
// and migrates the old custom css and branding sass settings to new webapp themes setting.
func updateWebappTheme(ctx context.Context, log *zap.Logger, s store.Storer) (err error) {
	log.Info("provision start")
	defer log.Info("provision end")

	vv, _, err := store.SearchSettingValues(ctx, s, types.SettingsFilter{})
	if err != nil {
		return err
	}

	oldCustomCSS := vv.FindByName("ui.custom-css")
	oldBranding := vv.FindByName("ui.studio.branding-sass")
	studioThemes := vv.FindByName("ui.studio.themes")
	customCSSThemes := vv.FindByName("ui.studio.custom-css")

	//get branding themes setting value
	brandingTheme, err := processBrandingTheme(oldBranding)
	if err != nil {
		return err
	}

	//get custom css themes setting value
	customCSSTheme, err := processCustomCSSTheme(oldCustomCSS)
	if err != nil {
		return err
	}

	// provision new themes
	if studioThemes.IsNull() {
		// provision branding themes setting
		err = provisionTheme(ctx, s, "ui.studio.themes", brandingTheme, log)
		if err != nil {
			return err
		}

		if !oldBranding.IsNull() {
			// delete old branding sass settings from the database
			err = store.DeleteSettingValue(ctx, s, oldBranding)
			if err != nil {
				return err
			}
		}
	}

	if customCSSThemes.IsNull() {
		// provision custom CSS themes setting
		err = provisionTheme(ctx, s, "ui.studio.custom-css", customCSSTheme, log)
		if err != nil {
			return err
		}

		if !oldCustomCSS.IsNull() {
			// delete old custom css settings from the database
			err = store.DeleteSettingValue(ctx, s, oldCustomCSS)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func provisionTheme(ctx context.Context, s store.Storer, name string, themes []types.Theme, log *zap.Logger) (err error) {
	value, err := json.Marshal(themes)
	if err != nil {
		return err
	}

	newThemeSetting := &types.SettingValue{
		Name:      name,
		Value:     value,
		UpdatedAt: time.Now(),
	}

	err = store.CreateSettingValue(ctx, s, newThemeSetting)
	if err != nil {
		log.Error("failed to provision webapp themes", zap.Error(err))
		return err
	}

	return nil
}

func processBrandingTheme(oldBranding *types.SettingValue) (themes []types.Theme, err error) {
	var brandingMap map[string]string

	lightModeMap := map[string]string{
		"black":       "#0B344E",
		"white":       "#FFFFFF",
		"primary":     "#4e73df",
		"secondary":   "#858796",
		"success":     "#1cc88a",
		"warning":     "#f6c23e",
		"danger":      "#e74a3b",
		"light":       "#f8f9fc",
		"extra-light": "#f8f9fc",
		"body-bg":     "#F3F5F7",
		"sidebar-bg":  "#FFFFFF",
		"topbar-bg":   "#F3F5F7",
	}

	// process old branding sass settings and match them with the new branding themes setting
	if !oldBranding.IsNull() {
		oldBrandingString, err := strconv.Unquote(oldBranding.Value.String())
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(oldBrandingString), &brandingMap); err != nil {
			return nil, err
		}

		for key, bmValue := range brandingMap {
			if key == "light" || key == "extra-light" {
				continue
			}

			if _, ok := lightModeMap[key]; ok {
				lightModeMap[key] = bmValue
			}
		}
	}

	// "Almost black" dark palette — matches client3/web/*/src/themes/corteza-base/dark.scss
	// and CUIBrandingEditor.vue's darkModeVariables, so a freshly provisioned
	// environment's dark theme matches what ships client-side by default.
	darkModeValues := `
    {
        "black":"#EDEDED",
        "white":"#121212",
        "primary":"#6E8FF0",
        "secondary":"#9A9A9A",
        "success":"#43AA8B",
        "warning":"#E27646",
        "danger":"#F2555A",
        "light":"#1A1A1A",
        "extra-light":"#242424",
        "body-bg":"#0A0A0A",
        "sidebar-bg": "#121212",
        "topbar-bg": "#121212"
    }`

	lightModeValues, _ := json.Marshal(lightModeMap)

	themes = []types.Theme{
		{
			ID:     sass.LightTheme,
			Values: string(lightModeValues),
		},
		{
			ID:     sass.DarkTheme,
			Values: darkModeValues,
		},
	}

	return themes, nil
}

func processCustomCSSTheme(oldValue *types.SettingValue) (themes []types.Theme, err error) {
	var generalCSS string

	if oldValue.IsNull() {
		generalCSS = ""
	} else {
		generalCSS, err = strconv.Unquote(oldValue.Value.String())
		if err != nil {
			return nil, err
		}
	}

	themes = []types.Theme{
		{
			ID:     sass.GeneralTheme,
			Values: generalCSS,
		},
		{
			ID:     sass.LightTheme,
			Values: "",
		},
		{
			ID:     sass.DarkTheme,
			Values: "",
		},
	}

	return themes, nil
}
