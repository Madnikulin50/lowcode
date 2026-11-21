package options

const (
	AvatarInitialsFontPoppins    = "fonts/Poppins-Regular.ttf"
	AvatarInitialsFontMontserrat = "fonts/Montserrat-Regular.ttf"
)

// Cleanup upgrades the Latin-only Poppins default to Montserrat, which has
// Cyrillic glyphs. A custom AVATAR_INITIALS_FONT_PATH other than Poppins is kept.
func (o *AttachmentOpt) Cleanup() {
	if o.AvatarInitialsFontPath == "" || o.AvatarInitialsFontPath == AvatarInitialsFontPoppins {
		o.AvatarInitialsFontPath = AvatarInitialsFontMontserrat
	}
}
