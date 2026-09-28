import Cocoa

let width: CGFloat = 660
let height: CGFloat = 400

func createBitmap(scale: CGFloat) -> NSBitmapImageRep {
    let pixelWidth = Int(width * scale)
    let pixelHeight = Int(height * scale)
    
    let rep = NSBitmapImageRep(
        bitmapDataPlanes: nil,
        pixelsWide: pixelWidth,
        pixelsHigh: pixelHeight,
        bitsPerSample: 8,
        samplesPerPixel: 4,
        hasAlpha: true,
        isPlanar: false,
        colorSpaceName: .deviceRGB,
        bytesPerRow: 0,
        bitsPerPixel: 0
    )!
    
    rep.size = NSSize(width: width, height: height) // 72 DPI logical size
    
    NSGraphicsContext.saveGraphicsState()
    let context = NSGraphicsContext(bitmapImageRep: rep)!
    NSGraphicsContext.current = context
    let cg = context.cgContext
    
    // Solid Dark Background
    let colorSpace = CGColorSpaceCreateDeviceRGB()
    let bgColors = [
        NSColor(red: 0.08, green: 0.09, blue: 0.12, alpha: 1.0).cgColor,
        NSColor(red: 0.06, green: 0.07, blue: 0.10, alpha: 1.0).cgColor
    ] as CFArray
    let bgGrad = CGGradient(colorsSpace: colorSpace, colors: bgColors, locations: [0.0, 1.0])!
    cg.drawLinearGradient(bgGrad, start: CGPoint(x: 0, y: height), end: CGPoint(x: 0, y: 0), options: [])

    // Subtle Outer Frame
    let frameRect = CGRect(x: 10, y: 10, width: width - 20, height: height - 20)
    let framePath = NSBezierPath(roundedRect: frameRect, xRadius: 14, yRadius: 14)
    NSColor(white: 1.0, alpha: 0.07).setStroke()
    framePath.lineWidth = 1.0
    framePath.stroke()

    // Header Title
    let title = "AetherGrok Desktop Studio"
    let titleAttrs: [NSAttributedString.Key: Any] = [
        .font: NSFont.systemFont(ofSize: 20, weight: .bold),
        .foregroundColor: NSColor(white: 0.95, alpha: 1.0)
    ]
    let titleSize = (title as NSString).size(withAttributes: titleAttrs)
    (title as NSString).draw(at: CGPoint(x: (width - titleSize.width) / 2, y: height - 48), withAttributes: titleAttrs)

    // Subtitle instruction
    let subtitle = "Drag AetherGrok to Applications to install"
    let subAttrs: [NSAttributedString.Key: Any] = [
        .font: NSFont.systemFont(ofSize: 12, weight: .medium),
        .foregroundColor: NSColor(red: 0.58, green: 0.64, blue: 0.72, alpha: 1.0)
    ]
    let subSize = (subtitle as NSString).size(withAttributes: subAttrs)
    (subtitle as NSString).draw(at: CGPoint(x: (width - subSize.width) / 2, y: height - 70), withAttributes: subAttrs)

    // Center Arrow Graphic (aligned horizontally between icon centers)
    // Left icon center is X: 160, Right icon center is X: 500. Center is X: 330.
    // In Finder with window height 400 and icons at AppleScript position {X, 150} (from top):
    // Finder coordinate Y=150 from top corresponds to NSGraphicsContext Y = 400 - 150 = 250 from bottom!
    let arrowY: CGFloat = 250

    cg.saveGState()
    cg.setLineWidth(3.0)
    cg.setLineCap(.round)
    let lineGradientColors = [
        NSColor(red: 0.23, green: 0.51, blue: 0.96, alpha: 0.2).cgColor,
        NSColor(red: 0.38, green: 0.65, blue: 0.98, alpha: 1.0).cgColor,
        NSColor(red: 0.58, green: 0.77, blue: 0.99, alpha: 0.8).cgColor
    ] as CFArray
    let lineGrad = CGGradient(colorsSpace: colorSpace, colors: lineGradientColors, locations: [0.0, 0.6, 1.0])!
    
    let arrowPath = CGMutablePath()
    arrowPath.move(to: CGPoint(x: 275, y: arrowY))
    arrowPath.addLine(to: CGPoint(x: 375, y: arrowY))
    cg.addPath(arrowPath)
    cg.replacePathWithStrokedPath()
    cg.clip()
    cg.drawLinearGradient(lineGrad, start: CGPoint(x: 275, y: arrowY), end: CGPoint(x: 375, y: arrowY), options: [])
    cg.restoreGState()

    // Chevrons
    let chevronColor = NSColor(red: 0.45, green: 0.70, blue: 1.0, alpha: 1.0)
    chevronColor.setStroke()
    
    let chv1 = NSBezierPath()
    chv1.move(to: CGPoint(x: 350, y: arrowY + 11))
    chv1.line(to: CGPoint(x: 365, y: arrowY))
    chv1.line(to: CGPoint(x: 350, y: arrowY - 11))
    chv1.lineWidth = 3.5
    chv1.lineCapStyle = .round
    chv1.lineJoinStyle = .round
    chv1.stroke()

    let chv2 = NSBezierPath()
    chv2.move(to: CGPoint(x: 366, y: arrowY + 11))
    chv2.line(to: CGPoint(x: 381, y: arrowY))
    chv2.line(to: CGPoint(x: 366, y: arrowY - 11))
    chv2.lineWidth = 3.5
    chv2.lineCapStyle = .round
    chv2.lineJoinStyle = .round
    chv2.stroke()

    // Pulse dots
    let dot1 = NSBezierPath(ovalIn: CGRect(x: 295, y: arrowY - 3, width: 6, height: 6))
    NSColor(red: 0.38, green: 0.65, blue: 0.98, alpha: 0.6).setFill()
    dot1.fill()

    let dot2 = NSBezierPath(ovalIn: CGRect(x: 320, y: arrowY - 3.5, width: 7, height: 7))
    NSColor(red: 0.38, green: 0.65, blue: 0.98, alpha: 0.9).setFill()
    dot2.fill()

    // Footer
    let footer = "v1.0.1 • Universal macOS Application"
    let footerAttrs: [NSAttributedString.Key: Any] = [
        .font: NSFont.systemFont(ofSize: 11, weight: .regular),
        .foregroundColor: NSColor(red: 0.42, green: 0.48, blue: 0.58, alpha: 0.9)
    ]
    let footerSize = (footer as NSString).size(withAttributes: footerAttrs)
    (footer as NSString).draw(at: CGPoint(x: (width - footerSize.width) / 2, y: 20), withAttributes: footerAttrs)

    NSGraphicsContext.restoreGraphicsState()
    return rep
}

let rep1x = createBitmap(scale: 1.0)
let rep2x = createBitmap(scale: 2.0)

let cwd = FileManager.default.currentDirectoryPath
let resURL = URL(fileURLWithPath: "\(cwd)/resources")

let png1xData = rep1x.representation(using: .png, properties: [:])!
let png2xData = rep2x.representation(using: .png, properties: [:])!

try! png1xData.write(to: resURL.appendingPathComponent("dmg-background.png"))
try! png2xData.write(to: resURL.appendingPathComponent("dmg-background@2x.png"))

print("Saved PNGs successfully.")
