package options

import "testing"

func TestAttachmentOptCleanupUsesMontserrat(t *testing.T) {
	o := &AttachmentOpt{AvatarInitialsFontPath: AvatarInitialsFontPoppins}
	o.Cleanup()
	if o.AvatarInitialsFontPath != AvatarInitialsFontMontserrat {
		t.Fatalf("got %q", o.AvatarInitialsFontPath)
	}

	o = &AttachmentOpt{AvatarInitialsFontPath: "fonts/Custom.ttf"}
	o.Cleanup()
	if o.AvatarInitialsFontPath != "fonts/Custom.ttf" {
		t.Fatalf("custom font was overwritten: %q", o.AvatarInitialsFontPath)
	}
}
