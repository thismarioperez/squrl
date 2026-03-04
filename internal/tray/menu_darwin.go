//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>
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
