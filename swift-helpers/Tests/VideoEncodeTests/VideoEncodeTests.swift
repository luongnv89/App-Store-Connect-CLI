import AVFoundation
import XCTest
@testable import asc_video_encode

final class VideoEncodeTests: XCTestCase {
    func testVideoPresetUsesDistinctExportPresets() {
        XCTAssertEqual(VideoPreset.store.exportPresetName, AVAssetExportPresetHighestQuality)
        XCTAssertEqual(VideoPreset.preview.exportPresetName, AVAssetExportPreset1920x1080)
        XCTAssertEqual(VideoPreset.compact.exportPresetName, AVAssetExportPreset1280x720)
    }
    
    func testRenderSizeDoesNotUpscaleOriginal() {
        let original = CGSize(width: 640, height: 360)
        let rendered = renderSize(for: original, maxResolution: CGSize(width: 1920, height: 1080))
        
        XCTAssertEqual(rendered.width, original.width)
        XCTAssertEqual(rendered.height, original.height)
    }
    
    func testRenderSizeDownscalesToFitPreset() {
        let original = CGSize(width: 3840, height: 2160)
        let rendered = renderSize(for: original, maxResolution: CGSize(width: 1280, height: 720))
        
        XCTAssertEqual(rendered.width, 1280)
        XCTAssertEqual(rendered.height, 720)
    }

    func testCompressionRatioMatchesGoStyleSemantics() {
        XCTAssertEqual(compressionRatio(originalFileSize: 200, outputFileSize: 100), 2.0)
        XCTAssertEqual(compressionRatio(originalFileSize: 100, outputFileSize: 200), 1.0)
        XCTAssertEqual(compressionRatio(originalFileSize: 100, outputFileSize: 0), 1.0)
    }

    func testOrientedSizeAppliesTransform() {
        let portrait = orientedSize(
            for: CGSize(width: 1920, height: 1080),
            applying: CGAffineTransform(rotationAngle: .pi / 2)
        )
        
        XCTAssertEqual(portrait.width, 1080, accuracy: 0.001)
        XCTAssertEqual(portrait.height, 1920, accuracy: 0.001)
    }
}
