import SwiftUI

/// Foundation Toast. A short message above the content. The caller decides when it shows and hides, and announces it.
/// Usage: `FToast("This is no longer available")`
public struct FToast: View {
    private let message: String

    public init(_ message: String) {
        self.message = message
    }

    public var body: some View {
        Text(message)
            .font(FoundationTokens.Typography.body)
            .foregroundStyle(FoundationTokens.Color.text1)
            .multilineTextAlignment(.center)
            .padding(.vertical, FoundationTokens.Space.s2)
            .padding(.horizontal, FoundationTokens.Space.s4)
            .background(FoundationTokens.Color.surface2, in: RoundedRectangle(cornerRadius: FoundationTokens.Radius.pill))
            .overlay {
                RoundedRectangle(cornerRadius: FoundationTokens.Radius.pill).strokeBorder(FoundationTokens.Color.line)
            }
            .accessibilityAddTraits(.isStaticText)
    }
}

#Preview {
    FToast("This is no longer available")
        .padding(FoundationTokens.Space.s4)
}
