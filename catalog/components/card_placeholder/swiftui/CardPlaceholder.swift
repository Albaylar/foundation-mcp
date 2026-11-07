import SwiftUI

/// Foundation CardPlaceholder. Holds a card's place in the real layout until its media is ready. Pulses gently; still with Reduce Motion.
/// Usage: `FCardPlaceholder(height: 240)`
public struct FCardPlaceholder: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var dim = false

    private let height: CGFloat
    private let cornerRadius: CGFloat

    public init(height: CGFloat, cornerRadius: CGFloat = FoundationTokens.Radius.media) {
        self.height = height
        self.cornerRadius = cornerRadius
    }

    public var body: some View {
        RoundedRectangle(cornerRadius: cornerRadius)
            .fill(FoundationTokens.Color.surface2)
            .frame(maxWidth: .infinity)
            .frame(height: height)
            .opacity(dim ? 0.55 : 1)
            .accessibilityHidden(true)
            .task {
                guard !reduceMotion else { return }
                withAnimation(.easeInOut(duration: FoundationTokens.Motion.base * 3).repeatForever(autoreverses: true)) {
                    dim = true
                }
            }
    }
}

#Preview {
    FCardPlaceholder(height: FoundationTokens.Space.s6 * 8)
        .padding(FoundationTokens.Space.s4)
}
