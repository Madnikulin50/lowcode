package provision

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
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
	} else if err = migrateLegacyNavyDark(ctx, s, studioThemes, log); err != nil {
		return err
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

	// "Almost black" dark palette — keep in sync with
	// client3/lib/vue/src/scss/dark.scss and CUIBrandingEditor.vue.
	darkModeValues, _ := json.Marshal(almostBlackDark)

	lightModeValues, _ := json.Marshal(lightModeMap)

	themes = []types.Theme{
		{
			ID:     sass.LightTheme,
			Values: string(lightModeValues),
		},
		{
			ID:     sass.DarkTheme,
			Values: string(darkModeValues),
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

// almostBlackDark is the canonical dark palette. Keep in sync with
// client3/lib/vue/src/scss/dark.scss and CUIBrandingEditor.vue.
var almostBlackDark = map[string]string{
	"black":       "#EDEDED",
	"white":       "#121212",
	"primary":     "#6E8FF0",
	"secondary":   "#9A9A9A",
	"success":     "#43AA8B",
	"warning":     "#E27646",
	"danger":      "#F2555A",
	"light":       "#1A1A1A",
	"extra-light": "#242424",
	"body-bg":     "#0A0A0A",
	"sidebar-bg":  "#121212",
	"topbar-bg":   "#121212",
}

// legacyNavyDarkBodyBG / legacyNavyDarkWhite are the distinctive surface
// colors of the old Corteza "navy" dark theme. Provision used to write them
// as the default, and CUIBrandingEditor kept resetting to them on save.
const (
	legacyNavyDarkBodyBG = "#092B40"
	legacyNavyDarkWhite  = "#0B344E"
)

func hexEq(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func isLegacyNavyDark(vars map[string]string) bool {
	return hexEq(vars["body-bg"], legacyNavyDarkBodyBG) && hexEq(vars["white"], legacyNavyDarkWhite)
}

// migrateLegacyNavyDark replaces a stored dark theme that still has the old
// Corteza navy defaults with the almost-black palette. Custom dark themes
// (anything that is not exactly that old default pair of surface colors)
// are left alone. ui.studio.themes is only created when null, so without
// this, existing installs keep the navy forever.
func migrateLegacyNavyDark(ctx context.Context, s store.Storer, studioThemes *types.SettingValue, log *zap.Logger) error {
	if studioThemes == nil || studioThemes.IsNull() {
		return nil
	}

	var themes []types.Theme
	if err := studioThemes.Value.Unmarshal(&themes); err != nil {
		return err
	}

	changed := false
	for i := range themes {
		if themes[i].ID != sass.DarkTheme {
			continue
		}

		vars := map[string]string{}
		if err := json.Unmarshal([]byte(themes[i].Values), &vars); err != nil {
			log.Warn("skipping dark-theme migration; values are not a JSON object", zap.Error(err))
			continue
		}

		if !isLegacyNavyDark(vars) {
			continue
		}

		encoded, err := json.Marshal(almostBlackDark)
		if err != nil {
			return err
		}
		themes[i].Values = string(encoded)
		changed = true
	}

	if !changed {
		return nil
	}

	if err := studioThemes.SetSetting(themes); err != nil {
		return err
	}
	studioThemes.UpdatedAt = time.Now()
	if err := store.UpdateSettingValue(ctx, s, studioThemes); err != nil {
		log.Error("failed to migrate legacy navy dark theme", zap.Error(err))
		return err
	}

	log.Info("migrated legacy navy dark theme to almost-black")
	return nil
}
