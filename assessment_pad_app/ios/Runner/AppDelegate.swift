import UIKit
import Flutter

@UIApplicationMain
@objc class AppDelegate: FlutterAppDelegate {
  private let fileExporter = IepFileExporter()

  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)
    if let controller = window?.rootViewController as? FlutterViewController {
      let channel = FlutterMethodChannel(
        name: "cn.irts.children.assessmentassistant/file_exporter",
        binaryMessenger: controller.binaryMessenger
      )
      channel.setMethodCallHandler { [weak self] call, result in
        guard call.method == "saveFile" else {
          result(FlutterMethodNotImplemented)
          return
        }
        self?.fileExporter.saveFile(call: call, result: result, presenter: controller)
      }
    }
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}

private final class IepFileExporter: NSObject, UIDocumentPickerDelegate {
  private var pendingResult: FlutterResult?
  private var temporaryFileURL: URL?

  func saveFile(
    call: FlutterMethodCall,
    result: @escaping FlutterResult,
    presenter: UIViewController
  ) {
    guard pendingResult == nil else {
      result(FlutterError(
        code: "export_in_progress",
        message: "已有文件正在导出，请稍后再试",
        details: nil
      ))
      return
    }
    guard
      let arguments = call.arguments as? [String: Any],
      let bytes = arguments["bytes"] as? FlutterStandardTypedData
    else {
      result(FlutterError(code: "invalid_arguments", message: "导出文件参数不正确", details: nil))
      return
    }

    let fileName = normalizedDocxName(arguments["fileName"] as? String)
    let directory = FileManager.default.temporaryDirectory
      .appendingPathComponent("iep-word-export", isDirectory: true)

    do {
      try FileManager.default.createDirectory(
        at: directory,
        withIntermediateDirectories: true,
        attributes: nil
      )
      let fileURL = uniqueURL(in: directory, fileName: fileName)
      try bytes.data.write(to: fileURL, options: [.atomic])

      pendingResult = result
      temporaryFileURL = fileURL
      let picker: UIDocumentPickerViewController
      if #available(iOS 14.0, *) {
        picker = UIDocumentPickerViewController(forExporting: [fileURL], asCopy: true)
      } else {
        picker = UIDocumentPickerViewController(url: fileURL, in: .exportToService)
      }
      picker.delegate = self
      picker.modalPresentationStyle = .formSheet
      presenter.present(picker, animated: true)
    } catch {
      result(FlutterError(
        code: "export_failed",
        message: "导出文件写入失败：\(error.localizedDescription)",
        details: nil
      ))
    }
  }

  func documentPicker(
    _ controller: UIDocumentPickerViewController,
    didPickDocumentsAt urls: [URL]
  ) {
    finish(true)
  }

  func documentPickerWasCancelled(_ controller: UIDocumentPickerViewController) {
    finish(false)
  }

  private func finish(_ saved: Bool) {
    pendingResult?(saved)
    pendingResult = nil
    if let temporaryFileURL {
      try? FileManager.default.removeItem(at: temporaryFileURL)
    }
    temporaryFileURL = nil
  }

  private func normalizedDocxName(_ raw: String?) -> String {
    let fallback = "IEP计划.docx"
    let trimmed = raw?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
    let base = trimmed.isEmpty ? fallback : trimmed
    let invalidCharacters = CharacterSet(charactersIn: "\\/:*?\"<>|")
    let cleaned = base.components(separatedBy: invalidCharacters).joined(separator: "_")
    return cleaned.lowercased().hasSuffix(".docx") ? cleaned : "\(cleaned).docx"
  }

  private func uniqueURL(in directory: URL, fileName: String) -> URL {
    let baseURL = directory.appendingPathComponent(fileName)
    if !FileManager.default.fileExists(atPath: baseURL.path) {
      return baseURL
    }
    let ext = (fileName as NSString).pathExtension
    let stem = (fileName as NSString).deletingPathExtension
    let suffix = Int(Date().timeIntervalSince1970)
    return directory.appendingPathComponent("\(stem)-\(suffix).\(ext)")
  }
}
