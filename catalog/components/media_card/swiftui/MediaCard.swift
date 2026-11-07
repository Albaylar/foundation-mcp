import SwiftUI

/// Foundation MediaCard. The picture or video is the card, and its title sits on a ground that travels with the text,
/// so it holds 4.5:1 over any picture and any number of lines. At accessibility text sizes the title moves under the media.
/// The caller gives the media its size. Usage: `FMediaCard("Golden hour", label: "Trend", action: open) { poster.frame(height: 240) }`
public struct FMediaCard<Media: View, Badge: View>: View {
    @Environment(\.dynamicTypeSize) private var typeSize
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    private let title: String
    private let label: String?
    private let titleFont: Font
    private let cornerRadius: CGFloat
    private let action: () -> Void
    private let media: Media
    private let badge: Badge

    public init(
        _ title: String,
        label: String? = nil,
        titleFont: Font = FoundationTokens.Typography.label,
        cornerRadius: CGFloat = FoundationTokens.Radius.media,
        action: @escaping () -> Void,
        @ViewBuilder media: () -> Media,
        @ViewBuilder badge: () -> Badge = { EmptyView() }
    ) {
        self.title = title
        self.label = label
        self.titleFont = titleFont
        self.cornerRadius = cornerRadius
        self.action = action
        self.media = media()
        self.badge = badge()
    }

    public var body: some View {
        Button(action: action) {
            if typeSize.isAccessibilitySize {
                VStack(alignment: .leading, spacing: FoundationTokens.Space.s2) {
                    picture
                    titleText.foregroundStyle(FoundationTokens.Color.text1)
                }
            } else {
                picture
                    .overlay(alignment: .bottomLeading) { caption }
                    .clipShape(RoundedRectangle(cornerRadius: cornerRadius))
            }
        }
        .buttonStyle(PressStyle(reduceMotion: reduceMotion))
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(label.map { "\(title), \($0)" } ?? title)
        .accessibilityAddTraits(.isButton)
    }

    private var picture: some View {
        media
            .frame(maxWidth: .infinity)
            .overlay(alignment: .topLeading) {
                badge.padding(FoundationTokens.Space.s2)
            }
            .clipShape(RoundedRectangle(cornerRadius: cornerRadius))
    }

    private var titleText: some View {
        Text(title)
            .font(titleFont)
            .multilineTextAlignment(.leading)
            .fixedSize(horizontal: false, vertical: true)
    }

    private var caption: some View {
        titleText
            .foregroundStyle(FoundationTokens.Color.onMedia)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(FoundationTokens.Space.s3)
            .padding(.top, FoundationTokens.Space.s6)
            .background {
                LinearGradient(
                    stops: [
                        .init(color: FoundationTokens.Color.mediaGround.opacity(0), location: 0),
                        .init(color: FoundationTokens.Color.mediaGround.opacity(0.62), location: 0.4),
                        .init(color: FoundationTokens.Color.mediaGround.opacity(0.9), location: 1)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )
            }
    }
}

private struct PressStyle: ButtonStyle {
    let reduceMotion: Bool

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed && !reduceMotion ? 0.97 : 1)
            .animation(.easeOut(duration: FoundationTokens.Motion.fast), value: configuration.isPressed)
    }
}

#Preview {
    FMediaCard("Golden hour", label: "Trend", action: {}) {
        FoundationTokens.Color.brand.frame(height: FoundationTokens.Space.s6 * 8)
    }
    .padding(FoundationTokens.Space.s4)
}
