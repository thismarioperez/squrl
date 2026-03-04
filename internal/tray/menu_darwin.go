//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics

#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <objc/runtime.h>

// openSystrayMenu accesses the NSStatusItem stored as an ivar on systray's
// AppDelegate (retrieved via [NSApp delegate]) and simulates a button click on
// the main thread, which opens the tray dropdown menu.
// No Accessibility permission is required because we are clicking our own
// UI element from within the same process.
void openSystrayMenu(void) {
    id delegate = [NSApp delegate];
    if (delegate == nil) return;

    Ivar ivar = class_getInstanceVariable([delegate class], "statusItem");
    if (ivar == nil) return;

    NSStatusItem *item = (__bridge NSStatusItem *)object_getIvar(delegate, ivar);
    if (item == nil) return;

    dispatch_async(dispatch_get_main_queue(), ^{
        [item.button performClick:nil];
    });
}

// suppressDockIcon explicitly sets the app activation policy to Accessory so
// that no Dock tile or app-switcher entry is shown. LSUIElement=true in the
// Info.plist should be sufficient, but macOS Sequoia (15+) may still briefly
// display the icon during startup unless the policy is also set in code.
void suppressDockIcon(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    });
}

// preflightScreenCapture returns true if Screen Recording permission has
// already been granted. The permission only becomes active after a relaunch,
// so this returning false is the signal to prompt and ask the user to restart.
bool preflightScreenCapture(void) {
    return CGPreflightScreenCaptureAccess();
}

// requestScreenCaptureAccess opens System Settings to the Screen Recording
// pane so the user can grant permission. The new permission takes effect only
// after the app is relaunched.
void requestScreenCaptureAccess(void) {
    CGRequestScreenCaptureAccess();
}
*/
import "C"

// openMenu programmatically opens the squrl tray dropdown menu by simulating
// a click on the status bar button via the Objective-C runtime.
func openMenu() {
	C.openSystrayMenu()
}

// initPlatform suppresses the Dock icon on macOS Sequoia and later, where
// LSUIElement alone may not prevent the icon from appearing at startup.
func initPlatform() {
	C.suppressDockIcon()
}

// hasScreenCapturePermission returns true if Screen Recording permission is
// already active. On macOS the permission only takes effect after a relaunch.
func hasScreenCapturePermission() bool {
	return bool(C.preflightScreenCapture())
}

// requestScreenCapturePermission opens System Settings to the Screen Recording
// pane. The granted permission takes effect only after the app is relaunched.
func requestScreenCapturePermission() {
	C.requestScreenCaptureAccess()
}
