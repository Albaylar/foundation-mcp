import SwiftUI

/// Foundation Pill. Plain information in a capsule: a status note or a fact on a surface, or a label with a colored dot on media.
/// The media variant always sits on a dark ground, so its tint needs 4.5:1 against that ground only.
/// Usage: `FPill("You're offline", systemImage: "wifi.slash")`, `FPill("Trend", tint: accent)`
public struct FPill: View {
    private let text: String
    private let systemImage: String?
    private let tint: Color?

    public init(_ text: String, systemImage: String? = nil) {
        self.text = text
        self.systemImage = systemImage
        self.tint = nil
    }

    public init(_ text: String, tint: Color) {
        self.text = text
        self.systemImage = nil
        self.tint = tint
    }

    public var body: some View {
        HStack(spacing: FoundationTokens.Space.s1) {
            if let systemImage {
                Image(systemName: systemImage).accessibilityHidden(true)
            }
            if let tint {
                Circle().fill(tint)
                    .frame(width: FoundationTokens.Space.s2, height: FoundationTokens.Space.s2)
                    .accessibilityHidden(true)
            }
            Text(text)
        }
        .font(FoundationTokens.Typography.caption)
        .foregroundStyle(tint ?? FoundationTokens.Color.text1)
        .padding(.vertical, FoundationTokens.Space.s1)
        .padding(.horizontal, FoundationTokens.Space.s3)
        .background(
            tint == nil ? FoundationTokens.Color.surface2 : FoundationTokens.Color.mediaGround.opacity(0.8),
            in: RoundedRectangle(cornerRadius: FoundationTokens.Radius.pill)
        )
    }
}

#Preview {
    VStack(spacing: FoundationTokens.Space.s2) {
        FPill("You're offline", systemImage: "wifi.slash")
        FPill("Trend", tint: FoundationTokens.Color.brand)
    }
    .padding(FoundationTokens.Space.s4)
}
