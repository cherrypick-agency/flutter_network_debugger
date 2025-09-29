import Cocoa
import FlutterMacOS

class MainFlutterWindow: NSWindow {
  override func awakeFromNib() {
    let flutterViewController = FlutterViewController()
    var windowFrame = self.frame
    if let screen = self.screen ?? NSScreen.main {
      let screenWidth = screen.visibleFrame.width
      // Target width similar to the screenshot (~1440), but cap to screen width
      let targetWidth: CGFloat = min(screenWidth, 1440)
      windowFrame.size.width = targetWidth
      // Keep height as-is but ensure it does not exceed screen height
      let maxHeight = screen.visibleFrame.height
      if windowFrame.size.height > maxHeight { windowFrame.size.height = maxHeight - 40 }
      // Center the window on screen
      let x = screen.visibleFrame.origin.x + (screenWidth - windowFrame.size.width) / 2
      let y = screen.visibleFrame.origin.y + (maxHeight - windowFrame.size.height) / 2
      windowFrame.origin = CGPoint(x: x, y: y)
    }
    self.contentViewController = flutterViewController
    self.setFrame(windowFrame, display: true)

    RegisterGeneratedPlugins(registry: flutterViewController)

    super.awakeFromNib()
  }
}
