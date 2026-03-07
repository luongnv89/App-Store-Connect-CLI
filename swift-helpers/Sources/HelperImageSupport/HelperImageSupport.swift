import CoreImage
import Foundation
import Metal

public enum HelperImageSupport {
    public static func makeCIContext(includeWorkingFormat: Bool = false) -> CIContext {
        let colorSpace = CGColorSpaceCreateDeviceRGB()

        if let device = MTLCreateSystemDefaultDevice() {
            var options: [CIContextOption: Any] = [
                .workingColorSpace: colorSpace,
                .outputColorSpace: colorSpace,
                .cacheIntermediates: false
            ]
            if includeWorkingFormat {
                options[.workingFormat] = CIFormat.RGBAf
            }
            return CIContext(mtlDevice: device, options: options)
        }

        return CIContext(options: [
            .workingColorSpace: colorSpace,
            .outputColorSpace: colorSpace
        ])
    }

    public static func loadImage(from path: String, errorFactory: (String) -> Error) throws -> CIImage {
        let url = URL(fileURLWithPath: path)
        guard let image = CIImage(contentsOf: url) else {
            throw errorFactory("Could not load image from \(path)")
        }
        return image
    }
}
