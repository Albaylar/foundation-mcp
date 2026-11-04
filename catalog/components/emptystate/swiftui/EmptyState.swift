import SwiftUI

/// Foundation EmptyState. Three poster shapes on a soft brand glow, then a title and one sentence.
/// The glow fills the space it is given, so a Liquid Glass tab bar and toolbar above it have color to pick up.
/// Usage: `FEmptyState("No results yet", message: "Finished results will appear here.", systemImage: "photo.on.rectangle")`
public struct FEmptyState: View {
    @Environment(\.colorScheme) private var colorScheme

    private let title: String
    private let message: String
    private let systemImage: String

    public init(_ title: String, message: String, systemImage: String) {
        self.title = title
        self.message = message
        self.systemImage = systemImage
    }

    public var body: some View {
        VStack(spacing: FoundationTokens.Space.s5) {
            posters
                .accessibilityHidden(true)
            VStack(spacing: FoundationTokens.Space.s2) {
                Text(title)
                    .font(FoundationTokens.Typography.title)
                    .foregroundStyle(FoundationTokens.Color.text1)
                Text(message)
                    .font(FoundationTokens.Typography.body)
                    .foregroundStyle(FoundationTokens.Color.text2)
            }
            .multilineTextAlignment(.center)
        }
        .padding(.horizontal, FoundationTokens.Space.s5)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background {
            RadialGradient(
                colors: [FoundationTokens.Color.brand.opacity(colorScheme == .dark ? 0.3 : 0.16), .clear],
                center: .center,
                startRadius: 0,
                endRadius: FoundationTokens.Space.s6 * 11
            )
            .ignoresSafeArea()
        }
        .accessibilityElement(children: .combine)
    }

    private var posters: some View {
        let width = FoundationTokens.Space.s6 * 3
        let height = FoundationTokens.Space.s6 * 4.5
        let shape = RoundedRectangle(cornerRadius: FoundationTokens.Radius.card)
        let side = LinearGradient(
            colors: [FoundationTokens.Color.surface2, FoundationTokens.Color.line],
            startPoint: .top,
            endPoint: .bottom
        )
        let center = LinearGradient(
            colors: [FoundationTokens.Color.brand, FoundationTokens.Color.brandStrong],
            startPoint: .top,
            endPoint: .bottom
        )
        return ZStack {
            shape.fill(side)
                .frame(width: width, height: height)
                .scaleEffect(0.92)
                .rotationEffect(.degrees(-10))
                .offset(x: -FoundationTokens.Space.s6 * 2 + FoundationTokens.Space.s1)
            shape.fill(side)
                .frame(width: width, height: height)
                .scaleEffect(0.92)
                .rotationEffect(.degrees(10))
                .offset(x: FoundationTokens.Space.s6 * 2 - FoundationTokens.Space.s1)
            shape.fill(center)
                .frame(width: width, height: height)
            Image(systemName: systemImage)
                .font(FoundationTokens.Typography.title)
                .imageScale(.large)
                .foregroundStyle(FoundationTokens.Color.onBrand)
        }
    }
}

#Preview {
    FEmptyState("No results yet", message: "Finished results will appear here.", systemImage: "photo.on.rectangle")
}
