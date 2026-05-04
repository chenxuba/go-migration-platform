import Cocoa
import FlutterMacOS

class MainFlutterWindow: NSWindow {
  override func awakeFromNib() {
    let flutterViewController = FlutterViewController()
    self.contentViewController = flutterViewController

    let targetSize = NSSize(width: 1366, height: 768)
    if let screenFrame = self.screen?.visibleFrame ?? NSScreen.main?.visibleFrame {
      let origin = NSPoint(
        x: screenFrame.midX - targetSize.width / 2,
        y: screenFrame.midY - targetSize.height / 2
      )
      self.setFrame(NSRect(origin: origin, size: targetSize), display: true)
    } else {
      self.setContentSize(targetSize)
    }
    self.contentMinSize = targetSize

    RegisterGeneratedPlugins(registry: flutterViewController)

    super.awakeFromNib()
  }
}
