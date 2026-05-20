import AppKit
import Foundation
import PDFKit
import Vision

struct OCRConfig {
    let pdfPath: String
    let startPage: Int
    let endPage: Int?
    let scale: CGFloat
}

func parseConfig() -> OCRConfig? {
    let args = CommandLine.arguments
    guard args.count >= 2 else {
        fputs("usage: ocr_pdf_vision.swift <pdf> [startPage] [endPage] [scale]\n", stderr)
        return nil
    }
    let start = args.count >= 3 ? max(Int(args[2]) ?? 1, 1) : 1
    let end = args.count >= 4 ? Int(args[3]) : nil
    let scale = args.count >= 5 ? CGFloat(Double(args[4]) ?? 2.0) : 2.0
    return OCRConfig(pdfPath: args[1], startPage: start, endPage: end, scale: scale)
}

func cgImage(for page: PDFPage, scale: CGFloat) -> CGImage? {
    let bounds = page.bounds(for: .mediaBox)
    let size = NSSize(width: bounds.width * scale, height: bounds.height * scale)
    let image = NSImage(size: size)
    image.lockFocus()
    guard let context = NSGraphicsContext.current?.cgContext else {
        image.unlockFocus()
        return nil
    }
    NSColor.white.setFill()
    NSRect(origin: .zero, size: size).fill()
    context.saveGState()
    context.scaleBy(x: scale, y: scale)
    page.draw(with: .mediaBox, to: context)
    context.restoreGState()
    image.unlockFocus()
    var rect = NSRect(origin: .zero, size: size)
    return image.cgImage(forProposedRect: &rect, context: nil, hints: nil)
}

func recognizeText(from image: CGImage) throws -> [String] {
    let request = VNRecognizeTextRequest()
    request.recognitionLevel = .accurate
    request.usesLanguageCorrection = true
    request.recognitionLanguages = ["zh-Hans", "en-US"]
    let handler = VNImageRequestHandler(cgImage: image, options: [:])
    try handler.perform([request])
    guard let observations = request.results else {
        return []
    }
    return observations
        .sorted {
            let yDelta = abs($0.boundingBox.minY - $1.boundingBox.minY)
            if yDelta > 0.01 {
                return $0.boundingBox.minY > $1.boundingBox.minY
            }
            return $0.boundingBox.minX < $1.boundingBox.minX
        }
        .compactMap { observation in
            observation.topCandidates(1).first?.string.trimmingCharacters(in: .whitespacesAndNewlines)
        }
        .filter { !$0.isEmpty }
}

guard let config = parseConfig() else {
    exit(2)
}

let url = URL(fileURLWithPath: config.pdfPath)
guard let document = PDFDocument(url: url) else {
    fputs("failed to open PDF: \(config.pdfPath)\n", stderr)
    exit(1)
}

let pageCount = document.pageCount
let startIndex = min(max(config.startPage - 1, 0), max(pageCount - 1, 0))
let requestedEnd = config.endPage ?? pageCount
let endIndex = min(max(requestedEnd - 1, startIndex), pageCount - 1)

print("--- file: \(url.lastPathComponent) ---")
print("--- pages: \(pageCount) ---")

for pageIndex in startIndex...endIndex {
    autoreleasepool {
        print("\n--- page \(pageIndex + 1) ---")
        guard let page = document.page(at: pageIndex), let image = cgImage(for: page, scale: config.scale) else {
            print("[OCR_RENDER_FAILED]")
            return
        }
        do {
            let lines = try recognizeText(from: image)
            if lines.isEmpty {
                print("[OCR_EMPTY]")
            } else {
                for line in lines {
                    print(line)
                }
            }
        } catch {
            print("[OCR_FAILED] \(error)")
        }
    }
}
